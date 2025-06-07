# 演示 PV 的创建过程

1. 静态创建 PV （HostPath）

```shell
./kubectl.sh apply -f examples/pv/local_test/hostpath-pv.yaml
```

查看 PV 情况，观察到可用

2. 创建PVC, 展示自动绑定

```shell
./kubectl.sh apply -f examples/pv/local_test/local-pvc.yaml
```

3. 再次创建PVC,展示PV不存在情况下PVC的自动创建

```shell
./kubectl.sh apply -f examples/pv/nfs_test/autogen-pvc.yaml
```

查看 PV 和 PVC 情况

# 展示Pod和PVC的绑定

## HostPath 单机演示

1. 创建一个Pod, 其向PV中写入数据
创建另一个Pod,绑定相同的PVC,从其中读取数据（使用Nginx即可）

2. 删除两个Pod,重启具有读取功能的Pod，验证数据已经存在。

## NFS 多机演示

然后使用自动创建的NFS PV进行多机演示。
创建一个Pod,向PV中写入数据。
创建多个读取用的Pod，调度到各个节点上，保证各个节点均能成功读取。