# 目前的 Pod 抽象


使用`Pod.Metadata.Namesapce`作为`Containerd`的`NameSpace`
`Pod`名称在同一命名空间下名称应该唯一。

`Container`在Containerd下的名称为`Pod名称-Container名称`

+ 容器
+ 镜像
+ 快照

# Pod抽象

共享网络和存储

生命周期与Pause容器相同

其余容器地位对等，加入到Pause容器的NetworkNamespace中

# Network Namespace

网络命名空间（Network Namespace）是 Linux 内核提供的一种隔离机制，用于隔离网络相关的资源，如网络接口、IP 地址、路由表、端口等。
每个网络命名空间是一个独立的网络栈，拥有自己的网络设备、IP 地址、防火墙规则等。

容器运行时（如 Containerd）可以通过 oci.WithLinuxNamespace("network", path) 指定容器加入某个网络命名空间，或者通过 oci.WithHostNamespace(oci.NetworkNamespace) 使用宿主机的网络命名空间。

TODO: 学习 Network Namespace?

```
> $ ls -l /proc/1470/ns/net                                                              [±feat/kubelet ●●]
lrwxrwxrwx 1 liu liu 0 Apr 22 22:03 /proc/1470/ns/net -> 'net:[4026533291]'
```

可以隔离网络栈

在修改后的 CreateContainer 函数中，你通过 oci.WithLinuxNamespace("network", netNSPath) 让业务容器加入 Pause 容器的网络命名空间（netNSPath 指向 /proc/<pause_pid>/ns/net）。
这确保 Pod 内的所有容器共享同一个网络命名空间，从而拥有相同的 IP 地址和网络配置。

Containerd Namespace：
在 Kubernetes 中，**Containerd namespace 通常与 Kubernetes namespace 对齐**，用于隔离不同命名空间的 Pod。例如，你的代码通过 pod.Metadata.Namespace 设置 Containerd namespace，确保 Pod 的容器操作在正确的隔离范围内。
同一个 Containerd namespace 可以包含多个 Pod，每个 Pod 有自己的网络命名空间。

在 Kubernetes 的 Pod 模型中，Pod 内的所有容器共享同一个网络命名空间，由 Pause 容器创建和持有。
Pause 容器的网络命名空间通过 CNI 插件配置网络（例如分配 Pod 的 IP 地址），Pod 内的其他容器通过加入这个网络命名空间共享网络配置。
在你的代码中，Pause 容器的 PID 用于获取网络命名空间路径（/proc/<pid>/ns/net），业务容器通过 oci.WithLinuxNamespace 加入该命名空间。


# 消息队列：CreatePod

# 上传 Kubelet

# 注册 API Server

监听消息队列

