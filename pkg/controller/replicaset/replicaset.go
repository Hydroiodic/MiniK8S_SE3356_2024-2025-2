package replicaset

import (
	"fmt"
	"log"
	"strings"
	"time"

	client "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/google/uuid"
)

type ReplicasetController struct {
}

func (rsc *ReplicasetController) Start() {
	ticker := time.NewTicker(5 * time.Second) // 每3秒触发一次
	defer ticker.Stop()                       // 确保程序退出时停止Ticker

	for {
		select {
		case <-ticker.C:
			rsc.CheckAllReplicaset()
		}
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
	ci := client.NewAPIClient("")

	for range num {
		var pod object.Pod
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
	ci := client.NewAPIClient("")
	for i := range num {
		ci.DeletePod(&pods[i])
	}
}

func (rsc *ReplicasetController) CheckAllReplicaset() {

	ci := client.NewAPIClient("")

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
			if HasMatchingLabels(curLabels, p.Metadata.Labels) {
				matchPods = append(matchPods, p)
			}
		}

		rs.Status.AvailableReplicas = len(matchPods)

		if len(matchPods) == rs.Spec.Replicas {
			rs.Status.AvailableReplicas = len(matchPods)
		} else if len(matchPods) < rs.Spec.Replicas {
			//创建新的pod
			log.Printf("数量不够 : %d", rs.Spec.Replicas-len(matchPods))
			rsc.CreatePod(rs.Spec.Replicas-len(matchPods), rs)
			rs.Status.AvailableReplicas = len(matchPods)
		} else {
			log.Printf("数量太多了 : %d", len(matchPods)-rs.Spec.Replicas)
			rsc.DeletePod(matchPods, len(matchPods)-rs.Spec.Replicas)
			rs.Status.AvailableReplicas = len(matchPods)
		}

		err = ci.UpdateReplicaset(rs)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// 判断两个 Label 是否有相同的键值对
func HasMatchingLabels(
	rsLabels, podLabels map[string]string,
) bool {
	for key, value := range rsLabels {
		if podValue, exists := podLabels[key]; exists && podValue == value {
			return true
		}
	}

	return false
}
