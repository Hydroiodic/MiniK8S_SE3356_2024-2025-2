/**
 * Heartbeat由Kublet发送，此处只需管理消息队列的处理
 */
package kubeproxy

import (
	"encoding/json"
	"log"
	"path"
	"time"

	"slices"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/ipvs_ops"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/mqtemplate"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type ServiceController struct {
	kubelet   *object.Kubelet
	IpvsOps   ipvs_ops.IpvsOpsInterface
	apiClient *apiserver.APIClient
}

func NewServiceController(
	kubelet *object.Kubelet,
	ipvsOps ipvs_ops.IpvsOpsInterface,
	apiClient *apiserver.APIClient,
) *ServiceController {
	return &ServiceController{
		kubelet:   kubelet,
		IpvsOps:   ipvsOps,
		apiClient: apiClient,
	}
}

// CreateServiceHandler 处理服务创建请求
// Service应该完整，携带所有信息，包括Endpoints
func (c *ServiceController) CreateServiceHandler(svc object.Service) error {
	// 检查服务是否已经存在
	c.kubelet.Mu.Lock()
	for _, s := range c.kubelet.Services {
		if s.Metadata.Namespace == svc.Metadata.Namespace &&
			s.Metadata.Name == svc.Metadata.Name {
			// TODO: 服务已经存在，进行更新
			// TODO: 服务已经存在，但是API Server又要我创建？
			// 调用 ipvs_ops 更新服务
			c.IpvsOps.UpdateServiceEps(&s, &svc)

			// 更新服务映射
			for i, s := range c.kubelet.Services {
				if s.Metadata.Name == svc.Metadata.Name {
					c.kubelet.Services[i] = svc
					break
				}
			}

			c.kubelet.LastUpdateTime = time.Now()
			c.kubelet.Mu.Unlock()

			return nil // 服务已存在，直接返回
		}
	}
	c.kubelet.Mu.Unlock()

	/** 服务不存在，直接添加 */

	// 调用 ipvs_ops 添加服务
	c.IpvsOps.AddService(&svc)

	// 将服务添加到服务映射中
	c.kubelet.Mu.Lock()
	c.kubelet.Services = append(c.kubelet.Services, svc)
	c.kubelet.Mu.Unlock()

	return nil
}

func (c *ServiceController) DeleteServiceHandler(svc *object.Service) error {
	// 调用 ipvs_ops 删除服务
	c.IpvsOps.DelService(svc)

	c.kubelet.Mu.Lock()
	// 从服务映射中删除服务
	for i, s := range c.kubelet.Services {
		if s.Metadata.Name == svc.Metadata.Name {
			c.kubelet.Services = slices.Delete(
				c.kubelet.Services, i,
				i+1,
			)

			break
		}
	}
	c.kubelet.Mu.Unlock()

	return nil
}

func (c *ServiceController) Run(stopCh <-chan struct{}) {
	// 处理消息队列中的 Service 创建和更新请求
	go func() {
		queueName := path.Join(mqtemplate.KubeProxyCreateServiceQueue,
			c.kubelet.Config.Name)
		err := mqtemplate.ConsumeMessageOnQueue(
			queueName,
			func(msg map[string]any) error {
				// 解析消息体
				msgBody, _ := json.Marshal(msg)

				var svc object.Service
				_ = json.Unmarshal(msgBody, &svc)

				// 调用处理函数
				if err := c.CreateServiceHandler(svc); err != nil {
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

	// 处理消息队列中的 Service 删除请求
	go func() {
		queueName := path.Join(
			mqtemplate.KubeProxyDeleteServiceQueue,
			c.kubelet.Config.Name,
		)
		err := mqtemplate.ConsumeMessageOnQueue(
			queueName,
			func(msg map[string]any) error {
				// 解析消息体
				msgBody, _ := json.Marshal(msg)

				var svc object.Service
				_ = json.Unmarshal(msgBody, &svc)

				// 调用处理函数
				if err := c.DeleteServiceHandler(&svc); err != nil {
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

	for range stopCh {
		return
	}
}
