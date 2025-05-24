package interfaces

import (
	"log"
	"net/http"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

// KubeProxyRegister registers a new KubeProxy or updates an existing one.
func KubeProxyRegister(c *gin.Context) { //nolint
	var kp object.KubeProxy
	if err := c.BindJSON(&kp); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if kp.Config.Name == "" {
		c.JSON(http.StatusBadRequest, "The kubeproxy name is empty.")
		return
	}

	// Create store.
	store, err := object.NewKubeProxyStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create kubeproxy store: "+err.Error(),
		)

		return
	}

	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("Failed to close kubeproxy store: %v", err)
		}
	}()

	// Check if kubeproxy already exists.
	oldKp, err := store.GetKubeProxy(c.Request.Context(), kp.Config.Name)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get kubeproxy: "+err.Error(),
		)

		return
	}

	if oldKp != nil {
		// Use old services.
		kp.Services = oldKp.Services
		log.Printf(
			"KubeProxy already exists: %s, use the previous service list\n",
			kp.Config.Name,
		)
	} else {
		// If not found, no services.
		kp.Services = make([]object.Service, 0)
		log.Printf("Registering KubeProxy %s with an empty service list\n", kp.Config.Name)
	}

	// Update heartbeat.
	kp.Heartbeat()

	// Write to etcd.
	if err := store.AddKubeProxy(c.Request.Context(), &kp); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to register kubeproxy: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "KubeProxy registered: "+kp.Config.Name)
	log.Println("KubeProxy registered:", kp.Config.Name)
}

// KubeProxyHeartbeat receives a heartbeat and updates services/endpoints as needed.
func KubeProxyHeartbeat(c *gin.Context) {
	var kp object.KubeProxy
	if err := c.BindJSON(&kp); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if kp.Config.Name == "" {
		c.JSON(http.StatusBadRequest, "The kubeproxy name is empty.")
		return
	}

	// Update heartbeat time.
	kp.Heartbeat()

	// Create store.
	store, err := object.NewKubeProxyStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create kubeproxy store: "+err.Error(),
		)

		return
	}

	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("Failed to close kubeproxy store: %v", err)
		}
	}()

	// Get Desired KubeProxy.
	desiredKp, err := store.GetKubeProxy(c.Request.Context(), kp.Config.Name)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get kubeproxy: "+err.Error(),
		)

		return
	}

	// Create Pod Store
	podStore, err := object.NewPodStore([]string{})

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create pod store: "+err.Error(),
		)

		return
	}

	defer func() {
		if err := podStore.Close(); err != nil {
			log.Printf("Failed to close pod store: %v", err)
		}
	}()

	// Get All the Pods from the store.
	pods, err := podStore.ListPods(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to list pods from store: "+err.Error(),
		)

		return
	}

	// TODO: 对于 DesiredServices，一律更新 EndPoints，如果乱序等等，后面再比较
	// 重新构建每个 Service 的 Endpoints
	for i := range desiredKp.Services {
		svc := &desiredKp.Services[i]
		// 跳过没有 selector 的 Service
		if svc.Spec.Selector == nil {
			continue
		}

		managedEps := make([]object.Endpoint, 0)

		for _, pod := range pods {
			// 只选 Running 且有 IP 的 Pod
			if pod.Status.Phase != object.PodRunning || pod.Status.IP == "" {
				continue
			}
			// 检查 Pod 是否符合 Service 的 Selector
			if isSelectedPod(pod, svc.Spec.Selector) {
				// 生成 Endpoint
				eps := getEndpointsFromPods([]*object.Pod{pod})
				managedEps = append(managedEps, eps...)
			}
		}

		svc.Status.Endpoints = managedEps
	}

	desiredServiceMap := make(map[string]object.Service)
	kubeProxyServiceMap := make(map[string]object.Service)

	// 1. 遍历 desiredKp.Services，填充 desiredServiceMap
	if desiredKp != nil {
		for _, desiredService := range desiredKp.Services {
			// Namespace + Name 作为唯一标识
			key := desiredService.Metadata.Namespace + "/" + desiredService.Metadata.Name
			desiredServiceMap[key] = desiredService
		}
	}
	// 2. 遍历 kp.Services，填充 kubeProxyServiceMap
	for _, kubeProxyService := range kp.Services {
		// Namespace + Name 作为唯一标识
		key := kubeProxyService.Metadata.Namespace + "/" + kubeProxyService.Metadata.Name
		kubeProxyServiceMap[key] = kubeProxyService
	}

	// Determine services to add, update, or delete.
	servicesToAdd := make([]object.Service, 0)
	servicesToUpdate := make([]object.Service, 0)
	servicesToDelete := make([]object.Service, 0)

	// 3. 遍历 desiredServiceMap 和 kubeProxyServiceMap，决定添加、更新或删除服务
	for key, desiredService := range desiredServiceMap {
		kubeProxyService, exists := kubeProxyServiceMap[key]
		if !exists {
			// Service not in kubeProxy, add it
			servicesToAdd = append(servicesToAdd, desiredService)
		} else {
			// Service exists, check if it needs to be updated
			if !endpointsEqual(desiredService.Status.Endpoints, kubeProxyService.Status.Endpoints) {
				// Service exists but is different, update it
				servicesToUpdate = append(servicesToUpdate, desiredService)
			}
		}
	}

	for key, kubeProxyService := range kubeProxyServiceMap {
		_, exists := desiredServiceMap[key]
		if !exists {
			// Service not in desiredKp, delete it
			servicesToDelete = append(servicesToDelete, kubeProxyService)
		}
	}

	log.Printf(
		"KubeProxy %s Heartbeat: %d services, %d to add, %d to update, %d to delete\n",
		kp.Config.Name,
		len(kp.Services),
		len(servicesToAdd),
		len(servicesToUpdate),
		len(servicesToDelete),
	)

	// TODO: 根据需求同步 servicesToAdd, servicesToUpdate, servicesToDelete
	// 可参照 KubeletHeartbeat 中的逻辑，如 internalUpdatePods / internalDeletePods

	// 更新 etcd 里的 kubeproxy
	if err := store.UpdateKubeProxy(c.Request.Context(), &kp); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to update kubeproxy: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "KubeProxy heartbeat: "+kp.Config.Name)
}

func getEndpointsFromPods(pods []*object.Pod) []object.Endpoint {
	eps := make([]object.Endpoint, 0)

	for _, pod := range pods {
		// 跳过无效的Pod
		if pod.Status.IP == "" || pod.Status.Phase != object.PodRunning {
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

func endpointsEqual(a, b []object.Endpoint) bool {
	if len(a) != len(b) {
		return false
	}

	amap := make(map[object.Endpoint]struct{})
	for _, ep := range a {
		amap[ep] = struct{}{}
	}

	for _, ep := range b {
		if _, ok := amap[ep]; !ok {
			return false
		}
	}

	return true
}
