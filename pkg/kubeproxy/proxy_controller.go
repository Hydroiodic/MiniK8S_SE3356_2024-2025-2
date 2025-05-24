package kubeproxy

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/ipvs_ops"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/mqtemplate"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func NewKubeProxy(config object.KubeProxyConfig) *object.KubeProxy {
	return &object.KubeProxy{
		Config:         config,
		LastUpdateTime: time.Now(),
	}
}

type KubeProxyService struct {
	kubeProxy  *object.KubeProxy
	syncPeriod time.Duration

	IpvsOps   ipvs_ops.IpvsOpsInterface
	apiClient *apiserver.APIClient
}

func NewKubeProxyService(
	config object.KubeProxyConfig,
	ipvsOps ipvs_ops.IpvsOpsInterface,
	apiClient *apiserver.APIClient,
	syncPeriod time.Duration,
) *KubeProxyService {
	return &KubeProxyService{
		kubeProxy:  NewKubeProxy(config),
		IpvsOps:    ipvsOps,
		apiClient:  apiClient,
		syncPeriod: syncPeriod,
	}
}

// CreateServiceHandler 处理服务创建请求
// Service应该完整，携带所有信息，包括Endpoints
func (kp *KubeProxyService) CreateServiceHandler(svc object.Service) error {
	kp.kubeProxy.Mu.Lock()
	kp.kubeProxy.LastUpdateTime = time.Now()
	kp.kubeProxy.Mu.Unlock()

	// 检查服务是否已经存在
	kp.kubeProxy.Mu.Lock()
	for _, s := range kp.kubeProxy.Services {
		if s.Metadata.Name == svc.Metadata.Name {
			kp.kubeProxy.Mu.Unlock()
			log.Printf(
				"Service %s already exists, skipping creation",
				svc.Metadata.Name,
			)

			return nil // 服务已存在，直接返回
		}
	}

	// 调用 ipvs_ops 添加服务
	kp.IpvsOps.AddService(&svc)

	// 将服务添加到服务映射中
	kp.kubeProxy.Mu.Lock()
	kp.kubeProxy.Services = append(kp.kubeProxy.Services, svc)
	kp.kubeProxy.Mu.Unlock()

	return nil
}

func (kp *KubeProxyService) DeleteServiceHandler(svc *object.Service) error {
	kp.kubeProxy.Mu.Lock()
	kp.kubeProxy.LastUpdateTime = time.Now()
	kp.kubeProxy.Mu.Unlock()

	// 调用 ipvs_ops 删除服务
	kp.IpvsOps.DelService(svc)

	kp.kubeProxy.Mu.Lock()
	// 从服务映射中删除服务
	for i, s := range kp.kubeProxy.Services {
		if s.Metadata.Name == svc.Metadata.Name {
			kp.kubeProxy.Services = append(
				kp.kubeProxy.Services[:i],
				kp.kubeProxy.Services[i+1:]...,
			)

			break
		}
	}
	kp.kubeProxy.Mu.Unlock()

	return nil
}

func (kp *KubeProxyService) UpdateServiceHandler(svc *object.Service) error {
	var oldSvc *object.Service

	kp.kubeProxy.Mu.Lock()
	// 遍历服务列表，找到要更新的服务
	for _, s := range kp.kubeProxy.Services {
		if s.Metadata.Name == svc.Metadata.Name {
			oldSvc = &s
			break
		}
	}
	kp.kubeProxy.Mu.Unlock()

	if oldSvc == nil {
		log.Printf("Service %s not found", svc.Metadata.Name)
		return nil // 服务不存在，返回
	}

	// 调用 ipvs_ops 更新服务
	kp.IpvsOps.UpdateServiceEps(oldSvc, svc)

	kp.kubeProxy.Mu.Lock()
	// 更新服务映射
	for i, s := range kp.kubeProxy.Services {
		if s.Metadata.Name == svc.Metadata.Name {
			kp.kubeProxy.Services[i] = *svc
			break
		}
	}

	kp.kubeProxy.LastUpdateTime = time.Now()
	kp.kubeProxy.Mu.Unlock()

	return nil
}

func (kp *KubeProxyService) Run(stopCh <-chan struct{}) {
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
				if err := kp.CreateServiceHandler(svc); err != nil {
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

	// 定时拉取最新的 Pod 和 Service 列表
	ticker := time.NewTicker(kp.syncPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// TODO: Send Heartbeat to APIServer
		case <-stopCh:
			log.Println("Stopping KubeProxy...")
			return
		}
	}
}

// TODO：添加一个通知函数，定时向APIServer汇报本地的Service
func (kp *KubeProxyService) NotifyAPIServer() {
}
