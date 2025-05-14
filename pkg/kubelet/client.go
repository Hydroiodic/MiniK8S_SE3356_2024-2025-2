package kubelet

type APIServerClient interface {
	// 发送心跳
	SendHeartbeat(kubelet *Kubelet) error
}
