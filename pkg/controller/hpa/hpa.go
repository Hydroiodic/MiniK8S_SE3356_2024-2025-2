package hpa

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	cadvisorutils "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/cadvisor"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/container"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type HPAController struct {
	// 认为HPA对象创建后只读，做更新之后它的UID一定发生变化
	HpasMap map[string]*object.HorizontalPodAutoscaler
	// HPA_ID -> Ticker的映射，用于定期检查某个HPA对象
	Tickers map[string]*time.Ticker
	// 新增一个map来存储每个HPA对象的退出通道
	QuitChs map[string]chan struct{}
	Ci      *apiserver.APIClient
	Cs      *container.ContainerService
}

func (hpaC *HPAController) Start() {
	hpaC.Cs, _ = container.NewContainerService()
	_, err := hpaC.Cs.RunCadvisorContainer()

	if err != nil {
		fmt.Println(err)
	}

	hpaC.HpasMap = make(map[string]*object.HorizontalPodAutoscaler)
	hpaC.Tickers = make(map[string]*time.Ticker)
	hpaC.QuitChs = make(map[string]chan struct{})
	// 建议比replicaSet处理时间长一点

	ticker := time.NewTicker(15 * time.Second)

	go func() {
		for range ticker.C {
			hpaC.CheckAllHPA()
		}
	}()

	hpaC.Ci = apiserver.NewAPIClient("")
}

// 增量式地同步HPA对象
func (hpaC *HPAController) CheckAllHPA() {
	//获取所有hpa对象
	hpas, err := hpaC.Ci.GetHpas()
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("开始检查 :::%d", len(hpas))

	updatedHpas := make(map[string]*object.HorizontalPodAutoscaler)

	for _, hpa := range hpas {
		// 用UID唯一标识一个HPA对象
		hpaKey := hpa.Metadata.Namespace + hpa.Metadata.Name
		updatedHpas[hpaKey] = &hpa

		// 如果是新的HPA对象，创建新的定时器和退出通道
		if _, ok := hpaC.HpasMap[hpaKey]; !ok {
			hpaC.Tickers[hpaKey] = time.NewTicker(
				5 * time.Second,
			)
			// 退出通道用于释放资源
			hpaC.QuitChs[hpaKey] = make(chan struct{})
			go func(hpa object.HorizontalPodAutoscaler) {
				// 创建后立即先检查一次，否则等待时间太长了！
				hpaC.CheckOneHPA(hpa)

				for {
					select {
					case <-hpaC.Tickers[hpaKey].C:
						hpaC.CheckOneHPA(hpa) // 检查单个HPA对象
					case <-hpaC.QuitChs[hpaKey]:
						delete(hpaC.Tickers, hpaKey)
						delete(hpaC.QuitChs, hpaKey)
						delete(hpaC.HpasMap, hpaKey)

						return // 收到退出信号，结束协程
					}
				}
			}(hpa)

			hpaC.HpasMap[hpaKey] = &hpa
		}
	}

	// 对于HpasMap中有但是在新的HPA列表中没有的，停止并删除对应的定时器和退出通道
	for hpaKey, ticker := range hpaC.Tickers {
		if _, ok := updatedHpas[hpaKey]; !ok {
			ticker.Stop()
			close(hpaC.QuitChs[hpaKey]) // 关闭退出通道，通知协程退出
		}
	}
}

// 这个函数用于在一个ticker协程里，定期检查一个HPA对象；需要各自维护状态
func (hpaC *HPAController) CheckOneHPA(hpa object.HorizontalPodAutoscaler) {
	// 获取所有pod对象
	pods, err := hpaC.Ci.GetPods()
	if err != nil {
		fmt.Println(err)
		return
	}
	// 根据hpa的配置，获得replicaSet
	replicaset, err := hpaC.Ci.GetReplicasetyName(
		hpa.Spec.ScaleTargetRef.Name,
	)
	if err != nil {
		fmt.Println(err)
		// return
		fmt.Println("该hpa已被删除")
	}

	// 现在已经获取到了replicaset，接下来需要获取到replicaset下的所有pod
	var matchPods []object.Pod

	for _, p := range pods {
		if apiserver.HasMatchingLabels(
			replicaset.Metadata.Labels,
			p.Metadata.Labels,
		) {
			matchPods = append(matchPods, p)
		}
	}

	// // 计算replicaset级别的所有pod的资源利用情况，只是做一个简单的算术平均
	podsMetricsEntries := make(map[string]object.PodMetrics, 0)

	for _, pod := range matchPods {
		containers := pod.Spec.Containers

		var cpuUsage = 0.0

		var memoryUsage = 0.0

		for _, container := range containers {
			fmt.Println(
				pod.Metadata.Namespace + "_" + pod.Metadata.Name + "_" + container.Name,
			)

			containercpuUsage, containermemoryUsage, err := cadvisorutils.GetContainerCPUandMem(
				"localhost",
				"8090",
				pod.Metadata.Namespace+"_"+pod.Metadata.Name+"_"+container.Name,
			)

			if err != nil {
				fmt.Println(err)
				continue
			}

			cpuUsage += containercpuUsage
			memoryUsage += containermemoryUsage
		}

		fmt.Printf("cpu利用率 %f\n", cpuUsage)
		fmt.Printf("内存利用率 %f\n", memoryUsage)

		podsMetricsEntries[pod.Metadata.Namespace+pod.Metadata.Name] =
			object.PodMetrics{
				Resources: map[string]float64{
					"cpu":    cpuUsage,
					"memory": memoryUsage,
				},
			}
	}

	rsMetricsResult := CalculateReplicaMetrics(
		&hpa,
		podsMetricsEntries,
	)

	// 当前replicaset的副本数，应该取spec静态期望值，而非真实值，可以留给replicaSetController一些调整时间
	currentReplicaNum := replicaset.Spec.Replicas
	// 期望副本数
	desiredReplicaNum := CalculateDesiredReplicas(
		&hpa,
		currentReplicaNum,
		rsMetricsResult,
	)

	fmt.Printf("当前replicaset的数量: %d\n", currentReplicaNum)

	fmt.Printf("期望replicaset的数量: %d\n", desiredReplicaNum)
	// 比较
	if currentReplicaNum == desiredReplicaNum {
		// 如果已经相等，则不管
		return
	} else {
		if currentReplicaNum < desiredReplicaNum {
			// 如果当前副本数小于期望副本数，则副本数自增1
			replicaset.Spec.Replicas = currentReplicaNum + 1
		} else {
			replicaset.Spec.Replicas = currentReplicaNum - 1
		}

		// 发给api-server，为这个replicaSet应用新的副本数；注意这里直接使用create方法，传递相同的静态配置（不需要管UID的问题，重建一份也没事），只是副本数不同
		err = hpaC.Ci.UpdateReplicaset(&replicaset)
		if err != nil {
			fmt.Printf("update replicaset failed : %s\n", err)
		}
	}
}

// 从这个ReplicaSet管理的所有Pods的监控指标百分比中（直接由参数给出），简单加和平均，计算出HPA需要的ReplicaSet级别的监控指标
func CalculateReplicaMetrics(
	h *object.HorizontalPodAutoscaler,
	metrics map[string]object.PodMetrics,
) object.PodMetrics {
	result := object.PodMetrics{
		Resources: make(map[string]float64), // 显式初始化
	}

	for _, metric := range h.Spec.Metrics {
		var total = 0.0

		var count = 0.0

		for _, podMetric := range metrics {
			switch metric.Resource.Name {
			case "cpu":
				total += podMetric.Resources["cpu"]
			case "memory":
				total += podMetric.Resources["memory"]
			default:
				fmt.Printf("Unknown metric: %s\n", metric.Resource.Name)
				continue
			}

			count++
		}

		if count > 0 {
			result.Resources[metric.Resource.Name] = total / count
		}
	}

	return result
}

// 有多项监控指标时，例如replicaSet级别的cpu和memory使用百分比（就是各个Pods的使用百分比均值）；需要先计算每一个单项的disiredReplicas，然后取最大值，这样作为整个HPA的disiredReplicas
func CalculateDesiredReplicas(
	h *object.HorizontalPodAutoscaler,
	curReplicaNum int,
	curMetrics object.PodMetrics,
) int {
	var maxDesiredReplicas = 0

	// 遍历每项指标
	for _, oneTargetMtc := range h.Spec.Metrics {
		// 这里oneCurMtc从map拿出来直接就是一个0~1的浮点值了！
		if oneCurMtc, exists := curMetrics.Resources[oneTargetMtc.Resource.Name]; exists {
			expectedReplicas := int(
				math.Ceil(
					float64(
						curReplicaNum,
					) * oneCurMtc / *oneTargetMtc.Resource.Target.AverageUtilization,
				),
			)

			fmt.Printf(
				"%s 目标指标值:%f\n",
				oneTargetMtc.Resource.Name,
				*oneTargetMtc.Resource.Target.AverageUtilization,
			)

			fmt.Printf("%s 当前指标值:%f\n", oneTargetMtc.Resource.Name, oneCurMtc)

			if expectedReplicas > maxDesiredReplicas {
				maxDesiredReplicas = expectedReplicas
			}
		}
	}

	if maxDesiredReplicas < int(h.Spec.MinReplicas) {
		maxDesiredReplicas = int(h.Spec.MinReplicas)
	} else if maxDesiredReplicas > int(h.Spec.MaxReplicas) {
		maxDesiredReplicas = int(h.Spec.MaxReplicas)
	}

	fmt.Printf(
		"CurReplicaNum: %v, final DesiredReplicas: %v\n",
		curReplicaNum,
		maxDesiredReplicas,
	)

	return maxDesiredReplicas
}
