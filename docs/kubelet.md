# Containerd

Containerd Namespace

Containerd namespace 是 Containerd 运行时用来隔离容器资源的一种逻辑分组机制，类似于 Kubernetes 中的 namespace，但作用范围仅限于 Containerd 的上下文。
它用于在 Containerd 中隔离容器、镜像、快照等资源，防止不同工作负载之间的资源冲突。
Containerd namespace 是一个字符串标识符（如 default、k8s.io），由用户或上层系统（如 Kubernetes）指定。

在 Containerd 中，namespace 通过上下文（context.Context）传递，例如 namespaces.WithNamespace(ctx, "my-namespace")。
所有 Containerd 操作（如创建容器、拉取镜像）都在指定的 namespace 内执行，资源（如容器 ID、快照）被绑定到该 namespace。

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
在 Kubernetes 中，Containerd namespace 通常与 Kubernetes namespace 对齐，用于隔离不同命名空间的 Pod。例如，你的代码通过 pod.Metadata.Namespace 设置 Containerd namespace，确保 Pod 的容器操作在正确的隔离范围内。
同一个 Containerd namespace 可以包含多个 Pod，每个 Pod 有自己的网络命名空间。

在 Kubernetes 的 Pod 模型中，Pod 内的所有容器共享同一个网络命名空间，由 Pause 容器创建和持有。
Pause 容器的网络命名空间通过 CNI 插件配置网络（例如分配 Pod 的 IP 地址），Pod 内的其他容器通过加入这个网络命名空间共享网络配置。
在你的代码中，Pause 容器的 PID 用于获取网络命名空间路径（/proc/<pid>/ns/net），业务容器通过 oci.WithLinuxNamespace 加入该命名空间。

# Containerd API

Containerd 是一个高性能的容器运行时，其 API 提供了丰富的功能来管理容器、镜像、快照、任务等资源。Containerd 的 API 是基于 gRPC 的，定义在 `containerd/api` 模块中，主要通过 Go 语言的客户端库（如 `github.com/containerd/containerd`）与 Containerd 服务交互。你的代码中已经使用了部分 Containerd API，例如创建容器、拉取镜像、启动任务等。

为了回答你的问题，我将详细介绍 Containerd API 中涉及的核心参数，结合你的代码，聚焦于常用的 API 调用（如创建容器、任务管理等）及其参数，尽量简洁但全面。如果你有特定的 API 或场景需要深入分析，可以进一步说明。

---

### 1. **Containerd API 概述**
Containerd 的 API 按功能划分为多个服务，每个服务对应一组操作。例如：
- **ContainerService**：管理容器（创建、删除、查询等）。
- **ImageService**：管理镜像（拉取、导入、删除等）。
- **TaskService**：管理容器任务（启动、停止、删除等）。
- **SnapshotService**：管理快照（文件系统层）。
- **EventService**：订阅和处理事件。
- **NamespaceService**：管理命名空间。

每个 API 调用需要通过 gRPC 客户端（`containerd.Client`）执行，并传递以下关键参数：
- **Context**：包含命名空间、超时等上下文信息。
- **Request Parameters**：特定于每个 API 的参数，例如容器 ID、镜像名称、OCI 规范等。
- **Options**：可选的配置参数，通常以 `containerd.WithXXX` 形式提供。

你的代码主要涉及 `ContainerService`、`ImageService`、`TaskService` 和 `EventService`，下面我将按这些服务逐一分析常用 API 及其参数。

---

### 2. **核心 API 及其参数**
以下是 Containerd API 中常用的方法及其涉及的参数，结合你的代码中的使用场景进行说明。

#### 2.1 **ImageService（镜像管理）**
**常用方法**：
- `Pull`：拉取镜像。

**API 示例**（你的代码）：
```go
image, err := client.Pull(ctx, containerSpec.Image, containerd.WithPullUnpack)
```

**参数**：
- `ctx context.Context`：
  - 上下文，包含 Containerd 命名空间（通过 `namespaces.WithNamespace` 设置）和超时信息。
  - 示例：`ctx = namespaces.WithNamespace(ctx, pod.Metadata.Namespace)`。
- `ref string`：
  - 镜像引用，例如 `docker.io/library/nginx:latest` 或 `registry.k8s.io/pause:3.9`。
  - 示例：`containerSpec.Image`。
- `opts ...containerd.RemoteOpt`：
  - 可选参数，用于配置拉取行为，例如：
    - `containerd.WithPullUnpack`：拉取后解压镜像到快照（你的代码中使用）。
    - `containerd.WithSchema1Conversion`：处理旧版镜像格式。
    - `containerd.WithPullSnapshotter`：指定快照驱动（如 `overlayfs`）。
  - 示例：`containerd.WithPullUnpack`。

**返回值**：
- `image containerd.Image`：拉取的镜像对象。
- `error`：错误信息。

---

#### 2.2 **ContainerService（容器管理）**
**常用方法**：
- `NewContainer`：创建容器。
- `LoadContainer`：加载已有容器。
- `Delete`：删除容器。

**API 示例**（你的代码）：
```go
container, err := client.NewContainer(
    ctx,
    formatContainerName(podName, containerSpec.Name),
    containerd.WithImage(image),
    containerd.WithNewSnapshot(fmt.Sprintf("snapshot-%s", containerSpec.Name), image),
    containerd.WithNewSpec(
        oci.WithImageConfig(image),
        oci.WithEnv([]string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}),
        oci.WithProcessArgs(containerSpec.Command...),
    ),
)
```

**参数**：
- `ctx context.Context`：
  - 同上，包含命名空间和超时。
- `id string`：
  - 容器唯一标识符，必须在命名空间内唯一。
  - 示例：`formatContainerName(podName, containerSpec.Name)`（如 `my-pod-my-container`）。
- `opts ...containerd.NewContainerOpts`：
  - 配置容器创建的选项，例如：
    - `containerd.WithImage(image)`：指定容器使用的镜像。
    - `containerd.WithNewSnapshot(snapshotID, image)`：创建快照（文件系统层）。
      - `snapshotID`：快照的唯一标识符（如 `snapshot-my-container`）。
      - `image`：关联的镜像。
    - `containerd.WithNewSpec(specOpts...)`：设置 OCI 运行时规范（`oci.Spec`）。
      - `specOpts ...oci.SpecOpts`：OCI 规范的配置，例如：
        - `oci.WithImageConfig(image)`：从镜像配置容器（例如入口点、环境变量）。
        - `oci.WithEnv([]string)`：设置环境变量。
        - `oci.WithProcessArgs(args...)`：设置容器进程的命令和参数。
        - `oci.WithLinuxNamespace(ns, path)`：设置 Linux 命名空间（如网络命名空间）。
          - 示例：`oci.WithLinuxNamespace("network", "/proc/<pause_pid>/ns/net")`（你的修改建议中）。
        - `oci.WithHostNamespace(ns)`：使用宿主机的命名空间（如 PID、网络）。
  - 示例：`containerd.WithImage(image)`、`containerd.WithNewSnapshot` 等。

**返回值**：
- `container containerd.Container`：创建的容器对象。
- `error`：错误信息。

**其他方法**：
- `LoadContainer(ctx, id)`：
  - 参数：`ctx`（上下文）、`id`（容器 ID）。
  - 返回：`containerd.Container` 或错误。
  - 示例：`client.LoadContainer(ctx, containerID)`。
- `Delete(ctx)`：
  - 参数：`ctx`（上下文）。
  - 返回：`error`。
  - 示例：`container.Delete(ctx)`。

---

#### 2.3 **TaskService（任务管理）**
**常用方法**：
- `NewTask`：创建任务（容器运行实例）。
- `Start`：启动任务。
- `Kill`：发送信号终止任务。
- `Wait`：等待任务退出。
- `Delete`：删除任务。

**API 示例**（你的代码）：
```go
task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
err = task.Start(ctx)
task.Kill(ctx, syscall.SIGTERM)
_, err = task.Wait(ctx)
_, err = task.Delete(ctx, containerd.WithProcessKill)
```

**参数**：
- **NewTask**：
  - `ctx context.Context`：上下文。
  - `cioCreator cio.CIO`：配置容器 I/O（如标准输入输出）。
    - 示例：`cio.NewCreator(cio.WithStdio)`（将容器 I/O 连接到标准输入输出）。
    - 其他选项：`cio.WithStreams(stdin, stdout, stderr)`（自定义 I/O 流）。
  - `opts ...containerd.NewTaskOpts`：任务创建选项，例如：
    - `containerd.WithTaskCheckpoint`：设置检查点。
    - `containerd.WithRootFS`：自定义根文件系统。
  - 返回：`containerd.Task` 和 `error`。
- **Start**：
  - `ctx context.Context`：上下文。
  - 返回：`error`。
- **Kill**：
  - `ctx context.Context`：上下文。
  - `signal syscall.Signal`：发送的信号（如 `syscall.SIGTERM`、`syscall.SIGKILL`）。
  - `opts ...containerd.KillOpts`：可选参数，例如 `containerd.WithKillAll`（杀死所有进程）。
  - 返回：`error`。
- **Wait**：
  - `ctx context.Context`：上下文。
  - 返回：`<-chan containerd.ExitStatus`（退出状态通道）和 `error`。
- **Delete**：
  - `ctx context.Context`：上下文。
  - `opts ...containerd.DeleteOpts`：删除选项，例如：
    - `containerd.WithProcessKill`：强制杀死进程（你的代码中使用）。
  - 返回：`containerd.ExitStatus` 和 `error`。

**其他方法**：
- `Task(ctx, cio)`：加载任务。
  - 参数：`ctx`（上下文）、`cio`（I/O 配置，可为 `nil`）。
  - 返回：`containerd.Task` 和 `error`。
  - 示例：`container.Task(ctx, nil)`。

---

#### 2.4 **EventService（事件订阅）**
**常用方法**：
- `Subscribe`：订阅 Containerd 事件。

**API 示例**（你的代码）：
```go
eventsCh, errCh := eventService.Subscribe(ctx)
```

**参数**：
- `ctx context.Context`：上下文，包含命名空间。
- `filters ...string`：事件过滤器（可选），指定订阅的事件类型，例如：
  - `"/containers/create"`：容器创建事件。
  - `"/tasks/exit"`：任务退出事件。
  - 示例：`eventService.Subscribe(ctx, "/containers/create", "/tasks/exit")`。
- 返回：
  - `<-chan *events.Envelope`：事件通道，包含事件数据。
  - `<-chan error`：错误通道。

**事件类型**（你的代码注释中提到）：
- `/containers/create`：容器创建。
- `/snapshots/prepare`：快照准备。
- `/tasks/create`、`tasks/start`：任务创建和启动。
- `/tasks/exit`：任务退出。
- `/tasks/delete`：任务删除。
- `/snapshots/remove`：快照清理。
- `/containers/delete`：容器删除。

**事件处理**：
- 事件数据通过 `typeurl.UnmarshalAny(event.Event)` 解码为具体类型（如 `events.ContainerCreate`、`events.TaskExit`）。
- 示例：你的代码中打印事件类型和详情。

---

#### 2.5 **SnapshotService（快照管理）**
**常用方法**：
- `Remove`：删除快照（你的代码中注释掉的部分）。

**API 示例**（你的代码）：
```go
// snapshotName := formatSnapshotName(containerSpec.Name)
// client.SnapshotService("overlayfs").Remove(ctx, snapshotName)
```

**参数**：
- `ctx context.Context`：上下文。
- `key string`：快照 ID（如 `snapshot-my-container`）。
- `snapshotter string`：快照驱动（如 `overlayfs`、`btrfs`）。
- 返回：`error`。

**注意**：
- 快照是容器文件系统的底层存储，创建容器时通过 `containerd.WithNewSnapshot` 指定。
- 删除快照需谨慎，确保容器和任务已删除。

---

### 3. **与你的代码的关系**
你的代码使用了 Containerd API 的核心功能，以下是代码中涉及的 API 和参数的总结：

- **镜像拉取**（`client.Pull`）：
  - 参数：`ctx`（命名空间）、`containerSpec.Image`（镜像引用）、`containerd.WithPullUnpack`。
  - 用途：拉取容器镜像（如 `nginx` 或 `pause`）。
- **容器创建**（`client.NewContainer`）：
  - 参数：`ctx`、容器 ID、镜像、快照、OCI 规范（包括环境变量、命令、网络命名空间等）。
  - 用途：创建 Pause 容器和业务容器。
- **任务管理**（`container.NewTask`, `task.Start`, `task.Kill`, `task.Wait`, `task.Delete`）：
  - 参数：`ctx`、I/O 配置、信号、删除选项。
  - 用途：启动和停止容器进程。
- **容器删除**（`container.Delete`）：
  - 参数：`ctx`。
  - 用途：清理容器资源。
- **事件订阅**（`eventService.Subscribe`）：
  - 参数：`ctx`。
  - 用途：监控容器和任务的状态变化。

**未使用但建议的参数**：
- **CNI 集成**：需要调用 CNI 插件为 Pause 容器的网络命名空间配置网络（如 IP 分配）。
  - 参数：网络命名空间路径（`/proc/<pid>/ns/net`）、Pod 元数据（名称、命名空间）。
- **快照清理**：你的代码注释了快照删除逻辑，建议在 `DeleteContainer` 中启用 `SnapshotService.Remove`。
- **检查点和恢复**：如果需要支持容器迁移，可以使用 `containerd.WithTaskCheckpoint`。

---

### 4. **参数总结**
以下是 Containerd API 中常见参数的分类总结：
- **上下文参数**：
  - `ctx context.Context`：包含命名空间（`namespaces.WithNamespace`）、超时、取消信号。
- **标识参数**：
  - 容器 ID、快照 ID、镜像引用、任务 ID。
- **配置参数**：
  - 镜像：`containerd.WithImage`。
  - 快照：`containerd.WithNewSnapshot`。
  - OCI 规范：`containerd.WithNewSpec`（包括命令、环境变量、命名空间等）。
  - I/O：`cio.NewCreator`。
- **选项参数**：
  - 拉取选项：`containerd.WithPullUnpack`。
  - 任务选项：`containerd.WithProcessKill`。
  - 事件过滤器：`/containers/create` 等。
- **运行时参数**：
  - 信号：`syscall.SIGTERM`、`syscall.SIGKILL`。
  - 网络命名空间路径：`/proc/<pid>/ns/net`。

---

### 5. **建议和扩展**
- **参数验证**：在调用 API 前，验证参数（如容器 ID 是否合法，镜像引用是否有效）以提高健壮性。
- **错误处理**：Containerd API 的错误可能包含具体原因（如 `not found`），建议解析错误类型以提供更精确的日志。
  - 示例：检查 `strings.Contains(err.Error(), "not found")`（你的代码中已使用）。
- **CNI 集成**：为 Pause 容器配置网络需要额外参数（如 CNI 配置文件路径、Pod 元数据），建议参考 `libcni` 库。
- **事件处理**：为 `Subscribe` 添加过滤器（如 `"/tasks/exit"`），只处理特定事件以减少开销。
- **快照管理**：启用快照删除逻辑，确保资源清理完整。

---

### 6. **总结**
Containerd API 涉及的参数主要围绕上下文、资源标识、配置选项和运行时控制。你的代码已覆盖核心参数（如容器 ID、镜像、OCI 规范、I/O 配置），但可以扩展以支持 CNI 网络配置、快照清理和更细粒度的事件处理。如果需要针对某个特定 API（如 `NewTask` 的高级选项、`SnapshotService` 的使用）或场景（如高并发下的参数优化）深入探讨，请提供更多细节，我可以进一步定制解答！

```go
// LinuxNamespaceType is one of the Linux namespaces
type LinuxNamespaceType string

const (
	// PIDNamespace for isolating process IDs
	PIDNamespace LinuxNamespaceType = "pid"
	// NetworkNamespace for isolating network devices, stacks, ports, etc
	NetworkNamespace LinuxNamespaceType = "network"
	// MountNamespace for isolating mount points
	MountNamespace LinuxNamespaceType = "mount"
	// IPCNamespace for isolating System V IPC, POSIX message queues
	IPCNamespace LinuxNamespaceType = "ipc"
	// UTSNamespace for isolating hostname and NIS domain name
	UTSNamespace LinuxNamespaceType = "uts"
	// UserNamespace for isolating user and group IDs
	UserNamespace LinuxNamespaceType = "user"
	// CgroupNamespace for isolating cgroup hierarchies
	CgroupNamespace LinuxNamespaceType = "cgroup"
	// TimeNamespace for isolating the clocks
	TimeNamespace LinuxNamespaceType = "time"
)
```

Network Namespace 隔离网络
Mount Namespace 隔离挂载