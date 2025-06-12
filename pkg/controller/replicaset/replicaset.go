package replicaset

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/google/uuid"
)

type ReplicasetController struct {
	// TODO: Please add a field `client` to interact with the API server.
}

func (rsc *ReplicasetController) Start() {
	ticker := time.NewTicker(5 * time.Second) // 每3秒触发一次
	defer ticker.Stop()                       // 确保程序退出时停止Ticker

	for range ticker.C {
		rsc.CheckAllReplicaset()
	}
}

func (rsc *ReplicasetController) NewUid() string {
	uid := uuid.New()
	uidstr := strings.Split(uid.String(), "-")[0]

	return uidstr
}

// pod数量不足，创建pod，需要给pod name加随机5位后缀
func (rsc *ReplicasetController) CreatePod(
	num int,
	rs object.ReplicaSet,
) {
	ci := apiserver.NewAPIClient("")

	fmt.Println("开始创建pod")

	for range num {
		var pod object.Pod
		pod.Metadata.Labels = rs.Metadata.Labels
		pod.Metadata = rs.Spec.Template.Metadata
		pod.Metadata.Name = "Replica" + "-" + rsc.NewUid()[:5]
		pod.Spec = rs.Spec.Template.Spec
		pod.Status = object.PodStatus{
			Phase:     "Running",
			StartTime: time.Now(), // 使用当前时间
		}
		pod.Kind = "Pod"

		err := ci.CreatePod(&pod)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (rsc *ReplicasetController) DeletePod(
	pods []object.Pod,
	num int,
) {
	fmt.Println("开始删除pod")

	ci := apiserver.NewAPIClient("")
	for i := range num {
		err := ci.DeletePod(&pods[i])
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func (rsc *ReplicasetController) CheckAllReplicaset() {
	ci := apiserver.NewAPIClient("")

	fmt.Println("开始检查所有replicaset")

	var replicasets []object.ReplicaSet

	var pods []object.Pod

	replicasets, err := ci.GetReplicasets()

	if err != nil {
		fmt.Println("get Replicaset from api server failed")
	}

	pods, err = ci.GetPods()
	if err != nil {
		fmt.Println("get Pods from api server failed")
	}

	for _, rs := range replicasets {
		curLabels := rs.Spec.Template.Metadata.Labels

		var matchPods []object.Pod

		for _, p := range pods {
			if apiserver.HasMatchingLabels(curLabels, p.Metadata.Labels) {
				matchPods = append(matchPods, p)
			}
		}

		rs.Status.AvailableReplicas = len(matchPods)

		fmt.Println("当前replica数量为")
		fmt.Println(matchPods)

		if len(matchPods) == rs.Spec.Replicas {
			rs.Status.AvailableReplicas = rs.Spec.Replicas
			return
		} else if len(matchPods) < rs.Spec.Replicas {
			// 创建新的pod
			log.Printf("数量不够 : %d", rs.Spec.Replicas-len(matchPods))
			rsc.CreatePod(rs.Spec.Replicas-len(matchPods), rs)
			rs.Status.AvailableReplicas = rs.Spec.Replicas
		} else {
			log.Printf("数量太多了 : %d", len(matchPods)-rs.Spec.Replicas)
			rsc.DeletePod(matchPods, len(matchPods)-rs.Spec.Replicas)
			rs.Status.AvailableReplicas = rs.Spec.Replicas
		}

		err = ci.UpdateReplicaset(&rs)
		log.Printf("更新数量 : %d", rs.Status.AvailableReplicas)

		if err != nil {
			fmt.Println(err)
		}
	}
}
