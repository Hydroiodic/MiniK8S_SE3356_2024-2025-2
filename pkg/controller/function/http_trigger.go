package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/Knetic/govaluate"
	"github.com/gin-gonic/gin"
)

type Serverless_controller struct {
	fcController     *FucntionController
	r                *gin.Engine
	route_map        map[string]string        //存储function的namespace/name和service的ip的映射
	timeVisitedQueue map[string]([]time.Time) // 每个function一个队列，记录最近的访问
	savedTimeLen     time.Duration            // 队列的时间窗口长度
	funcPodNums      map[string]int           // 每个function的pod数量
	lastVisitTime    map[string]time.Time     // 每个function的最后一次访问
	ToZeroTime       time.Duration            // 函数未被调用多久后触发缩容到零
	lastScaleTime    map[string]time.Time     // 最后一次scale的时间
	minScaleGap      time.Duration            //两次扩缩容操作之间的最小间隔时间
	ci               *apiserver.APIClient
}

// 所有的key都是namespace/name

func NewServerlessController() *Serverless_controller {
	return &Serverless_controller{

		fcController:     NewFucntionController(),
		r:                gin.Default(),
		route_map:        make(map[string]string),
		timeVisitedQueue: make(map[string]([]time.Time)),
		savedTimeLen:     time.Minute * 1,
		funcPodNums:      make(map[string]int),
		lastVisitTime:    make(map[string]time.Time),
		ToZeroTime:       time.Minute * 5,
		lastScaleTime:    make(map[string]time.Time),
		minScaleGap:      time.Second * 15,
		ci:               apiserver.NewAPIClient(""),
	}
}

func (s *Serverless_controller) Start() {
	go s.fcController.Start()
	ticker := time.NewTicker(2 * time.Second)

	go func() {
		for range ticker.C {
			//检查是否需要scale-to-zero
			s.CheckAllFunction()
			s.UpdateRoute()
		}
	}()

	s.r.POST(
		"/triggerFunction/:functionNamespace/:functionName",
		s.TriggerFunction,
	)
	s.r.POST(
		"/triggerWorkflow/:workflowNamespace/:workflowName",
		s.TriggerWorkflow,
	)
	err := s.r.Run(":8060")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func (s *Serverless_controller) UpdateRoute() {
	fmt.Println("开始更新route")
	serviceList, err := s.ci.GetServices()
	if err != nil {
		fmt.Println(err)
		return
	}
	remoteService := make(map[string]string)
	//更新本地没有，但是etcd中有的
	for _, service := range serviceList {
		if val, ok := service.Metadata.Labels["FunctionMetadata"]; ok {
			s.route_map[val] = service.Status.ClusterIP + ":" + strconv.Itoa(
				service.Spec.Ports[0].Port,
			)
			fmt.Printf(
				"func: %s ,route%s\n",
				val,
				service.Status.ClusterIP+":"+strconv.Itoa(
					service.Spec.Ports[0].Port,
				),
			)
			remoteService[val] = service.Status.ClusterIP + strconv.Itoa(
				service.Spec.Ports[0].Port,
			)

			if s.timeVisitedQueue[val] == nil {
				s.timeVisitedQueue[val] = make([]time.Time, 0)
				s.funcPodNums[val] = 1 // 启动时有一个Pod，认为在server把它干掉
				s.lastVisitTime[val] = time.Now().Add(-60 * time.Minute)
				s.lastScaleTime[val] = time.Now().Add(-60 * time.Minute)
			}

		}
	}
	//删除本地有，但是etcd中没有的
	for key := range s.route_map {
		if _, ok := remoteService[key]; !ok {
			delete(s.route_map, key)
			delete(s.timeVisitedQueue, key)
			delete(s.funcPodNums, key)
			delete(s.lastVisitTime, key)
			delete(s.lastScaleTime, key)
		}
	}

}
func (s *Serverless_controller) CheckAllFunction() {
	for key := range s.timeVisitedQueue {
		s.ScaleToZeroFunction(key)
	}
}

// 周期性检查是否需要缩容（或 scale-to-zero
func (s *Serverless_controller) ScaleToZeroFunction(name string) {
	q := s.timeVisitedQueue[name]

	for {
		if len(q) == 0 {
			//说明很长时间没人访问，已经scale to zero
			break
		}
		if time.Since(q[0]) > s.savedTimeLen {
			//更新timequeue，使得访问记录都是在savedtimelen范围内
			q = q[1:]
		}
	}
	s.timeVisitedQueue[name] = q
	nameStr := strings.Split(name, "/")
	if s.funcPodNums[name] > 0 &&
		time.Since(s.lastVisitTime[name]) > s.ToZeroTime {
		fmt.Println("开始缩容")
		// scale to zero
		fmt.Println(name, "scale to zero")
		s.lastScaleTime[name] = time.Now()
		s.funcPodNums[name] = 0

		//更改replicaset的内容
		rs, err := s.ci.GetReplicasetyName(nameStr[1])
		fmt.Println(rs)
		if err != nil {
			fmt.Println(err)
			return
		}
		rs.Spec.Replicas = 0
		err = s.ci.UpdateReplicaset(&rs)
		fmt.Println("更新replicaset为0")
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func (s *Serverless_controller) TriggerFunction(c *gin.Context) {

	functionName := c.Param("functionName")
	functionNamespace := c.Param("functionNamespace")
	fmt.Printf(
		"functionName: %s, functionNamespace: %s\n",
		functionName,
		functionNamespace,
	)

	name := functionNamespace + "/" + functionName
	functionServiceIP, ok := s.route_map[name]

	if !ok {
		c.JSON(404, gin.H{"error": "Function not found"})
		return
	}

	s.visitFunction(name)
	sendPath := "http://" + functionServiceIP
	fmt.Println("triggerFunction", sendPath)
	request_body, _ := io.ReadAll(c.Request.Body)
	fmt.Printf("TriggerFunction request is: %s\n", string(request_body))
	req, err := http.NewRequest(
		"POST",
		sendPath,
		bytes.NewBuffer(request_body),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("send post request failed", err.Error())
		return
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Println("Failed to close response body: ", cerr)
		}
	}()

	// Check if the response status code is OK (200).
	if resp.StatusCode != http.StatusOK {
		fmt.Println(err)
		return
	}
	bodyBytes, err := io.ReadAll(resp.Body) // 读取整个响应体
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close() // 确保关闭 Body

	bodyString := string(bodyBytes) // 转换为字符串
	fmt.Println(bodyString)
}

func (s *Serverless_controller) visitFunction(name string) {
	s.timeVisitedQueue[name] = append(s.timeVisitedQueue[name], time.Now())
	s.lastVisitTime[name] = time.Now()

	scaleNewNum := int(
		math.Ceil(math.Sqrt(float64(len(s.timeVisitedQueue[name])))),
	)

	scaleOldNum := s.funcPodNums[name]
	if scaleOldNum == 0 ||
		scaleNewNum > scaleOldNum { //只有scale up 和 刚启动才更新
		fmt.Printf("更新数量，新数量为%d\n", scaleNewNum)
		s.funcPodNums[name] = scaleNewNum

		s.lastScaleTime[name] = time.Now()

		nameStr := strings.Split(name, "/")

		rs, err := s.ci.GetReplicasetyName(nameStr[1])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		rs.Spec.Replicas = scaleNewNum
		err = s.ci.UpdateReplicaset(&rs)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

	}
}

func (s *Serverless_controller) TriggerWorkflow(c *gin.Context) {
	workflowName := c.Param("workflowName")
	workflowNamespace := c.Param("workflowNamespace")
	fmt.Printf(
		"workflowName: %s, workflowNamespace: %s\n",
		workflowName,
		workflowNamespace,
	)

	workflows, err := s.ci.GetAllWorkflow()
	if err != nil {
		fmt.Println(err)
	}

	var cur_w object.Workflow
	for _, w := range workflows {
		fmt.Println(w.Metadata.Name)
		if workflowName == w.Metadata.Name {
			cur_w = w
			break
		}
	}
	fmt.Println(cur_w)
	request_body, _ := io.ReadAll(c.Request.Body)
	var curParamsMap map[string]interface{}
	// 如果请求体不为空，那么直接使用请求体作为此时的参数
	if len(request_body) > 0 {
		fmt.Printf("TriggerWorkflow request is: %s\n", string(request_body))
		err = json.Unmarshal(request_body, &curParamsMap)
		if err != nil {
			fmt.Println(err.Error())
			c.JSON(400, gin.H{"error": "Parse Entry Params Error"})

			return
		}

	} else if len(cur_w.Spec.EntryParams) > 0 {
		// 请求体为空，但是workflow对象中有默认参数，那么使用默认参数
		curParamsMap = cur_w.Spec.EntryParams
	} else {
		// 上述两者都为空，那么直接使用空对象
		err = json.Unmarshal([]byte("{}"), &curParamsMap)
		if err != nil {
			fmt.Println(err.Error())
			c.JSON(400, gin.H{"error": "Parse Entry Params Error"})

			return
		}
	}

	fmt.Printf("EntryParams: %v\n", curParamsMap)

	fmt.Printf("now EntryNode is: %s\n", cur_w.Spec.EntryNode)

	// // 开始递归调用
	isStartFunc := true
	curNodeName := cur_w.Spec.EntryNode
	lastFuncResultMap := make(map[string]interface{})
	for {
		fmt.Printf(
			"Workflow %s/%s goto Node %s\n",
			workflowNamespace,
			workflowName,
			curNodeName,
		)

		curNode, ok := cur_w.Spec.Nodes[curNodeName]
		//当前workflow结束
		if curNodeName == "" || !ok {
			fmt.Printf(
				"Workflow %s/%s completed! final ResponseStr is %s\n",
				workflowNamespace,
				workflowName,
				lastFuncResultMap,
			)
			c.JSON(200, lastFuncResultMap)
			return
		}

		if !isStartFunc {
			// 如果之前调用过函数，那么更新此时的新参数
			curParamsMap = lastFuncResultMap
		}

		if curNode.Type == "func" {

			// 查看本次调用的Func的IP
			curFuncNamespace := curNode.FuncNodeRef.Metadata.Namespace
			curFuncName := curNode.FuncNodeRef.Metadata.Name

			functionServiceIP, ok := s.route_map[curFuncNamespace+"/"+curFuncName]
			if !ok {
				errStr := fmt.Sprintf(
					"In Workflow %s/%s, function %s/%s not found",
					workflowNamespace,
					workflowName,
					curFuncNamespace,
					curFuncName,
				)
				fmt.Println(errStr)
				c.JSON(404, gin.H{"error": errStr})
				return
			}
			sendPath := "http://" + functionServiceIP
			curRequestBody, err := json.Marshal(curParamsMap)
			if err != nil {
				errStr := fmt.Sprintf(
					"In Workflow %s/%s, function %s/%s params marshal err",
					workflowNamespace,
					workflowName,
					curFuncNamespace,
					curFuncName,
				)
				fmt.Println(errStr)
				return
			}

			fmt.Printf(
				"Workflow %s/%s start do function %s/%s, request is: %s\n",
				workflowNamespace,
				workflowName,
				curFuncNamespace,
				curFuncName,
				string(curRequestBody),
			)

			s.visitFunction(curFuncNamespace + "/" + curFuncName)
			req, err := http.NewRequest(
				"POST",
				sendPath,
				bytes.NewBuffer(curRequestBody),
			)
			if err != nil {
				fmt.Println(err)
				return
			}

			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("send post request failed", err.Error())
				return
			}
			defer func() {
				if cerr := resp.Body.Close(); cerr != nil {
					fmt.Println("Failed to close response body: ", cerr)
				}
			}()

			// Check if the response status code is OK (200).
			if resp.StatusCode != http.StatusOK {
				fmt.Println(err)
				return
			}

			// 获取响应的字节形式，应该提取为map
			var curFuncResultMap map[string]interface{}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(body, &curFuncResultMap)

			if err != nil {
				fmt.Println(err.Error())
				errStr := fmt.Sprintf(
					"In Workflow %s/%s, function %s/%s response parse error",
					workflowNamespace,
					workflowName,
					curFuncNamespace,
					curFuncName,
				)
				fmt.Println(errStr)
				c.JSON(400, gin.H{"error": errStr})
				return
			}

			// 将这次的结果存入上一次的结果中，以便下一次的计算
			lastFuncResultMap = curFuncResultMap
			isStartFunc = false
			// 更新转移节点
			curNodeName = curNode.FuncNodeRef.Next
			// 休息一小会，防止太快
			time.Sleep(1 * time.Second)
			fmt.Printf(
				"Workflow %s/%s done one function %s/%s, response is: %s\n",
				workflowNamespace,
				workflowName,
				curFuncNamespace,
				curFuncName,
				string(body),
			)
		} else if curNode.Type == "choice" {

			// 从上到下获取选择条件
			isSomeConditionMatched := false
			for _, condition := range curNode.ChoiceNodeRef.Conditons {
				// 利用govaluate库进行表达式计算
				// 获取表达式
				expressionStr := condition.Expression
				expr, err := govaluate.NewEvaluableExpression(expressionStr)
				if err != nil {
					fmt.Println(err.Error())
					errStr := fmt.Sprintf("In Workflow %s/%s, choice node %s parse expression error", workflowNamespace, workflowName, curNodeName)
					fmt.Println(errStr)
					c.JSON(400, gin.H{"error": errStr})
					return
				}

				calculateRes, _ := expr.Evaluate(curParamsMap)
				resIsOk := false
				// 这里计算出来可能是bool，也可能是算术
				if boolRes, ok := calculateRes.(bool); ok {
					resIsOk = boolRes
				} else if numRes, ok := calculateRes.(float64); ok {
					if numRes != 0 {
						resIsOk = true
					} else {
						resIsOk = false
					}
				} else if strRes, ok := calculateRes.(string); ok {
					resIsOk = strRes != ""
				}

				if resIsOk { //condition 为 true,可以进行下一步
					fmt.Printf("Workflow %s/%s choice node %s, expression %s matched\n", workflowNamespace, workflowName, curNodeName, expressionStr)

					curNodeName = condition.Next
					isSomeConditionMatched = true

					break
				} else {
					continue
				}
			}

			// condition为false，
			if !isSomeConditionMatched {
				fmt.Printf("Workflow %s/%s choice node %s, no expression matched, default goto end\n", workflowNamespace, workflowName, curNodeName)
				curNodeName = ""
			}
		} else {
			// 不支持种类的节点，直接返回
			errStr := fmt.Sprintf("In Workflow %s/%s, node %s type %s not supported", workflowNamespace, workflowName, curNodeName, curNode.Type)
			fmt.Println(errStr)
			c.JSON(400, gin.H{"error": errStr})
			return
		}
	}
}
