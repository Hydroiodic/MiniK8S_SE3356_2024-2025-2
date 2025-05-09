package hpa

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/google/cadvisor/client"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type HPAController struct {
	HPAQueue      chan *object.HorizontalPodAutoscaler
	MetricsClient *client.Client   // cAdvisor 客户端
	EtcdClient    *clientv3.Client // etcd 客户端
}

func NewHPAController(etcdEndpoints []string) (*HPAController, error) {
	// 初始化 etcd 客户端
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %v", err)
	}

	// 初始化 cAdvisor 客户端
	metricsClient, err := client.NewClient("http://localhost:8080/")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cAdvisor: %v", err)
	}

	return &HPAController{
		HPAQueue:      make(chan *object.HorizontalPodAutoscaler, 100),
		MetricsClient: metricsClient,
		EtcdClient:    etcdClient,
	}, nil
}

func (c *HPAController) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(15 * time.Second) // 每 15 秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.reconcileHPA()
		case hpa := <-c.HPAQueue:
			c.processHPA(hpa)
		case <-stopCh:
			return
		}
	}
}

// processHPA 是一个处理每个 HPA 事件的函数，它将会从队列中取出一个 HPA 对象并进行处理。
func (c *HPAController) processHPA(hpa *object.HorizontalPodAutoscaler) {

}

// 获取所有 HPA 配置（从 etcd 获取）
func (c *HPAController) getAllHPA() ([]*object.HorizontalPodAutoscaler, error) {
	// 假设 HPA 数据存储在 etcd 中某个特定的键下，键名为 "/hpas/"
	resp, err := c.EtcdClient.Get(
		context.Background(),
		"/hpas/",
		clientv3.WithPrefix(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HPAs from etcd: %v", err)
	}

	var hpas []*object.HorizontalPodAutoscaler

	for _, kv := range resp.Kvs {
		hpa := &object.HorizontalPodAutoscaler{}
		// 假设 hpa 数据是以 JSON 格式存储的
		err := json.Unmarshal(kv.Value, hpa)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal HPA data: %v", err)
		}

		hpas = append(hpas, hpa)
	}

	return hpas, nil
}

// 核心扩缩容逻辑
func (c *HPAController) reconcileHPA() {
	// 获取所有 HPA 配置
	hpas, err := c.getAllHPA()
	if err != nil {
		fmt.Printf("Error getting HPA: %v\n", err)
		return
	}

	// 2. 遍历每个 HPA，检查指标
	for _, hpa := range hpas {
		// 获取目标资源（如 ReplicaSet）
		target := getTargetWorkload(hpa.Spec.ScaleTargetRef)

		// 获取当前指标值
		currentMetrics := c.getCurrentMetrics(hpa)

		// 计算期望副本数
		desiredReplicas := calculateDesiredReplicas(hpa, currentMetrics)

		// 应用扩缩容策略
		adjustedReplicas := applyScalingPolicy(hpa, desiredReplicas)

		// 更新副本数
		if (int)(adjustedReplicas) != target.Status.Replicas {
			updateReplicaSet(target, adjustedReplicas)
		}
	}
}

// 该函数会根据 HPA 的 ScaleTargetRef 获取目标工作负载（如 ReplicaSet）。你可能需要查询 Kubernetes API 来实现这一点。
func getTargetWorkload(
	scaleTargetRef object.CrossVersionObjectReference,
) *object.ReplicaSet {
	return nil
}

// 该函数已经有了一个框架，可以通过 cAdvisor 获取目标 Pod 的资源使用情况（如 CPU 和内存），你需要实现获取目标指标的逻辑。
func (c *HPAController) getCurrentMetrics(
	hpa *object.HorizontalPodAutoscaler,
) map[string]float64 {
	return nil
}

// 这个函数会更新目标工作负载（ReplicaSet）的副本数。
func updateReplicaSet(target *object.ReplicaSet, adjustedReplicas int32) {

}

// 该函数根据实际的资源使用情况来计算期望的副本数。
func calculateDesiredReplicas(
	hpa *object.HorizontalPodAutoscaler,
	metrics map[string]float64,
) int32 {
	return 0
}

// 应用扩缩容策略
func applyScalingPolicy(
	hpa *object.HorizontalPodAutoscaler,
	desired int32,
) int32 {
	return 0
}
