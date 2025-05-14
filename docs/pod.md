# 处理Pod的状态

可以给Pause容器打上标签
通过Label先筛选出Pause容器，然后就能搞出Pod的Namespace和Name
然后啥状态啥的就都能判断了！

K8s 的 Pod 同步机制
Watch 机制：
K8s 的 API Server 支持通过 HTTP 长连接（基于 watch 参数）提供事件流。Kubelet 通过 Watch API 订阅与本节点相关的 Pod 变化（如创建、更新、删除）。
当 API Server 检测到 Pod 变化时，会通过 Watch 通道推送事件（如 ADDED、MODIFIED、DELETED），kubelet 实时接收这些事件。
Watch 机制基于 HTTP/1.1 的 chunked 编码或 gRPC 流，效率较高，且与 K8s 的 API 模型无缝集成。
Informer 和 Lister：
Kubelet 使用客户端库（如 client-go）中的 Informer 机制，维护一个本地缓存（PodCache），存储与节点相关的 Pod 信息。
Informer 通过 Watch 通道监听 Pod 变化，更新本地缓存，并触发回调函数处理事件（如创建或删除 Pod）。
Kubelet 定期通过 List API 拉取全量 Pod 列表（ListPods），以确保缓存与 API Server 一致（防止 Watch 事件丢失）。
Pod 同步循环：
Kubelet 运行一个 PodSyncLoop，定期调用 SyncPods 方法，比较本地缓存的 desiredPods（从 API Server 获取）与运行时的 actualPods（从容器运行时获取）。
收到 Watch 事件时，kubelet 会触发增量同步，立即处理特定 Pod 的变化。
Kubelet 还维护一个 PodManager，管理 Pod 的状态（Pending、Running 等），并通过容器运行时（如 containerd、CRI-O）执行创建、删除等操作。
自治性：
当 API Server 不可访问时，kubelet 依靠本地缓存（PodCache）继续管理现有 Pod。
Kubelet 会定期尝试重新连接 API Server，并在恢复连接后重新同步状态。