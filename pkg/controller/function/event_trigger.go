package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v2"
)

type EventTriggerController struct {
	EventMap map[string]*object.Event
	Crons    map[string]*cron.Cron
	QuitChs  map[string]chan struct{}
	Ci       *apiserver.APIClient
}

func NewEventController() *EventTriggerController {
	return &EventTriggerController{
		EventMap: make(map[string]*object.Event),
		Crons:    make(map[string]*cron.Cron),
		QuitChs:  make(map[string]chan struct{}),
		Ci:       apiserver.NewAPIClient(""),
	}
}
func (ec *EventTriggerController) Start() {
	ticker := time.NewTicker(15 * time.Second)

	go func() {
		for range ticker.C {
			ec.CheckAllEventTrigger()
		}
	}()
}

func (ec *EventTriggerController) CheckAllEventTrigger() {
	fmt.Println("开始检查所有的event")

	es, err := ec.Ci.GetAllEvent()
	if err != nil {
		fmt.Println(err)
		return
	}

	updateEvent := make(map[string]*object.Event)

	for _, ps := range es {
		psKey := ps.Metadata.Namespace + "/" + ps.Metadata.Name
		updateEvent[psKey] = &ps

		if _, ok := ec.EventMap[psKey]; !ok {
			//当前event还没有被实行
			c := cron.New(cron.WithSeconds())
			// 配置可退出的定时器函数
			_, err := c.AddFunc(ps.Schedule, func() {
				select {
				case <-ec.QuitChs[psKey]:
					return
				default:
					ec.CheckOneEvent(ps)
				}
			})
			if err != nil {
				fmt.Printf("Failed to create cron for Event: %s\n", err)
				continue
			}

			c.Start()
			ec.Crons[psKey] = c
			ec.QuitChs[psKey] = make(chan struct{})
			ec.EventMap[psKey] = &ps
		}
	}

	for psKey, c := range ec.Crons {
		if _, ok := updateEvent[psKey]; !ok {
			//处理已经删除的event
			data, _ := yaml.Marshal(ec.EventMap[psKey])
			fmt.Printf(
				"Detect current PingSource removed, stop cron for it: \n%s\n",
				string(data),
			)
			c.Stop()
			close(ec.QuitChs[psKey])
			delete(ec.Crons, psKey)
			delete(ec.QuitChs, psKey)
			delete(ec.EventMap, psKey)
		}
	}
}

func (ec *EventTriggerController) CheckOneEvent(ps object.Event) {
	fmt.Printf(
		"CheckOnePingSource %s/%s\n",
		ps.Metadata.Namespace,
		ps.Metadata.Name,
	)
	// 默认只支持向某个函数/workflow发请求
	kindStr := strings.ToLower(ps.Target.Kind)
	if kindStr != "function" && kindStr != "workflow" {
		fmt.Printf("kind is not function or workflow, do nothing")
		return
	}

	evNamespace := ps.Target.Namespace
	if evNamespace == "" {
		evNamespace = "default"
	}

	evName := ps.Target.Name
	evUrl := evNamespace + "/" + evName

	if kindStr == "workflow" {
		evUrl = "http://localhost:8060/triggerWorkflow/" + evUrl
	} else {
		evUrl = "http://localhost:8060/triggerFunction/" + evUrl
	}
	//先转成map，再转成byte
	var dataMapping map[string]interface{}
	err := json.Unmarshal([]byte(ps.JsonData), &dataMapping)

	if err != nil {
		fmt.Printf("Failed to unmarshal json data: %s\n", err)
		return
	}

	data, err := json.Marshal(dataMapping)

	if err != nil {
		fmt.Printf("Failed to unmarshal json data: %s\n", err)
	}

	req, err := http.NewRequest(
		"POST",
		evUrl,
		bytes.NewBuffer(data),
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

	if resp.StatusCode != http.StatusOK {
		fmt.Println(err)
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body) // 读取整个响应体

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Println("Failed to close response body: ", cerr)
		}
	}()

	bodyString := string(bodyBytes) // 转换为字符串
	fmt.Printf(
		"do trigger %s, response from serverless function: %s\n",
		evUrl,
		bodyString,
	)
}
