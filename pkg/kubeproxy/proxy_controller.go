package kubeproxy

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	ctr_pod "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/ipvs_ops"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/mqtemplate"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type KubeProxy struct {
	IpvsOps   ipvs_ops.IpvsOpsInterface
	apiClient *apiserver.APIClient

	ServiceMap map[string]*object.Service // FIXME: 服务名称到服务对象的映射（不管命名空间了，懒了）
	PodMap     map[string]*object.Pod
	syncPeriod time.Duration
	// 建议加一个锁
	Mu sync.Mutex
}

func NewKubeProxy(
	ipvsOps ipvs_ops.IpvsOpsInterface,
	apiClient *apiserver.APIClient,
	syncPeriod time.Duration,
) *KubeProxy {
	return &KubeProxy{
		IpvsOps:    ipvsOps,
		apiClient:  apiClient,
		ServiceMap: make(map[string]*object.Service),
		PodMap:     make(map[string]*object.Pod),
		syncPeriod: syncPeriod,
	}
}

func getEndpointsFromPods(pods []*object.Pod) []object.Endpoint {
	eps := make([]object.Endpoint, 0)

	for _, pod := range pods {
		// 跳过无效的Pod
		if pod.Status.IP == "" || pod.Status.Phase != ctr_pod.PodStatusRunning {
			continue
		}

		for _, container := range pod.Spec.Containers {
			// 如果这个容器暴露的端口不为空
			if len(container.Ports) > 0 {
				// 逐一绑定
				for _, port := range container.Ports {
					ep := object.Endpoint{
						IP:   pod.Status.IP,
						Port: port.ContainerPort,
					}
					eps = append(eps, ep)
				}
			}
		}
	}

	return eps
}

// 单独检查一个Pod是否满足给定的Selector，上面这个函数只是一个数组的封装
func isSelectedPod(pod *object.Pod, selector map[string]string) bool {
	for key, val := range selector {
		if pod.Metadata.Labels[key] != val {
			return false
		}
	}

	return true
}

func (kp *KubeProxy) CreateServiceHandler(svc *object.Service) error {
	kp.Mu.Lock()
	defer kp.Mu.Unlock()

	// 检查服务是否已经存在
	if _, exists := kp.ServiceMap[svc.Metadata.Name]; exists {
		return nil // 服务已存在，返回
	}

	// TODO: 填充 EndPoints
	if len(svc.Status.Endpoints) == 0 {
		var pods []*object.Pod

		for _, pod := range kp.PodMap {
			if pod.Metadata.Namespace == svc.Metadata.Namespace {
				pods = append(pods, pod)
			}
		}

		svc.Status.Endpoints = getEndpointsFromPods(pods)
	}

	// 调用 ipvs_ops 添加服务
	kp.IpvsOps.AddService(svc)

	// 添加服务到服务映射
	kp.ServiceMap[svc.Metadata.Name] = svc

	return nil
}

func (kp *KubeProxy) DeleteServiceHandler(svc *object.Service) error {
	kp.Mu.Lock()
	defer kp.Mu.Unlock()

	// 检查服务是否存在
	if _, exists := kp.ServiceMap[svc.Metadata.Name]; !exists {
		return nil // 服务不存在，返回
	}

	// 调用 ipvs_ops 删除服务
	kp.IpvsOps.DelService(svc)

	// 从服务映射中删除服务
	delete(kp.ServiceMap, svc.Metadata.Name)

	return nil
}

func (kp *KubeProxy) SyncPodsAndServices(
	pods []object.Pod,
	svcs []object.Service,
) {
	kp.Mu.Lock()
	defer kp.Mu.Unlock()

	// 清空 PodMap，重新填充
	kp.PodMap = make(map[string]*object.Pod)
	for _, pod := range pods {
		kp.PodMap[pod.Metadata.Name] = &pod
	}

	// 创建一个新的服务映射，用于增量更新
	updatedSvcs := make(map[string]*object.Service)

	for _, svc := range svcs {
		if svc.Spec.Selector == nil {
			continue
		}

		// 遍历所有 Pod，挑选出能被该 Service 管理的所有 Endpoints
		managedEps := make([]object.Endpoint, 0)

		for _, pod := range kp.PodMap {
			// 跳过无效的 Pod
			if pod.Status.IP == "" ||
				pod.Status.Phase != ctr_pod.PodStatusRunning {
				continue
			}

			// 检查 Pod 是否匹配 Service 的 Selector
			if isSelectedPod(pod, svc.Spec.Selector) {
				newEps := getEndpointsFromPods([]*object.Pod{pod})
				managedEps = append(managedEps, newEps...)
			}
		}

		// 深拷贝 Service 并更新 Endpoints
		data, _ := json.Marshal(&svc)

		var svcCopy object.Service
		_ = json.Unmarshal(data, &svcCopy)
		svcCopy.Status.Endpoints = managedEps
		updatedSvcs[svc.Metadata.Name] = &svcCopy
	}

	// 同步服务：先删除旧的服务，再添加或更新新的服务
	for name, svc := range kp.ServiceMap {
		if _, ok := updatedSvcs[name]; !ok {
			// 如果旧的服务不在新的服务列表中，删除它
			kp.IpvsOps.DelService(svc)
			delete(kp.ServiceMap, name)
		}
	}

	for name, svc := range updatedSvcs {
		if _, ok := kp.ServiceMap[name]; !ok {
			// 如果是新的服务，直接添加
			kp.IpvsOps.AddService(svc)
		} else {
			// 如果服务已存在，更新其 Endpoints
			kp.IpvsOps.UpdateServiceEps(kp.ServiceMap[name], svc)
		}
		// 更新服务映射
		kp.ServiceMap[name] = svc
	}
}

func (kp *KubeProxy) Run(stopCh <-chan struct{}) {
	// 处理消息队列中的 Service 创建请求
	go func() {
		err := mqtemplate.ConsumeMessageOnQueue(
			mqtemplate.CreateServiceQueueName,
			func(msg map[string]interface{}) error {
				// 解析消息体
				msgBody, _ := json.Marshal(msg)

				var svc object.Service
				_ = json.Unmarshal(msgBody, &svc)

				// 调用处理函数
				if err := kp.CreateServiceHandler(&svc); err != nil {
					log.Printf("Failed to create service: %v", err)
				}

				log.Printf("Service created: %s", svc.Metadata.Name)

				return nil
			},
		)
		if err != nil {
			log.Printf("Failed to consume message: %v", err)
		}
	}()

	go func() {
		err := mqtemplate.ConsumeMessageOnQueue(
			mqtemplate.DeleteServiceQueueName,
			func(msg map[string]interface{}) error {
				// 解析消息体
				msgBody, _ := json.Marshal(msg)

				var svc object.Service
				_ = json.Unmarshal(msgBody, &svc)

				// 调用处理函数
				if err := kp.DeleteServiceHandler(&svc); err != nil {
					log.Printf("Failed to delete service: %v", err)
				}

				log.Printf("Service deleted: %s", svc.Metadata.Name)

				return nil
			},
		)
		if err != nil {
			log.Printf("Failed to consume message: %v", err)
		}
	}()

	// TODO: 奇怪的立即更新队列

	// 定时拉取最新的 Pod 和 Service 列表
	ticker := time.NewTicker(kp.syncPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// TODO
			pods, err := kp.apiClient.GetPods()
			if err != nil {
				log.Printf("Failed to get pods: %v", err)
				continue
			}

			svcs, err := kp.apiClient.GetServices()
			if err != nil {
				log.Printf("Failed to get services: %v", err)
				continue
			}
			// 同步 Pod 和 Service
			kp.SyncPodsAndServices(pods, svcs)
		case <-stopCh:
			log.Println("Stopping KubeProxy...")
			return
		}
	}
}
