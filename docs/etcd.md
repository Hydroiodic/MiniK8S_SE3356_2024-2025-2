你希望在 Ubuntu 上安装 etcd **v3.4** 版本（而不是最新的 v3.5.15）。以下是针对 etcd v3.4（以 v3.4.26 为例，这是 v3.4 系列的最后一个版本）的安装和配置步骤，以及解决之前提到的 `etcdctl` 命令（如 `put`、`get`、`version`）问题的方案。

### 为什么选择 v3.4？
- etcd v3.4 是一个较旧但稳定的版本，广泛用于早期的 Kubernetes 集群（例如 v1.18 至 v1.22）。
- v3.4 支持 v3 API，`etcdctl` 的 `put`、`get` 和 `version` 命令都可用，因此之前的“No help topic”问题很可能是因为你当前使用了更旧的版本（可能是 v2.x）。

### 安装 etcd v3.4 的步骤

#### 1. **卸载现有 etcd（如果有）**
为了避免版本冲突，先检查并清理现有的 `etcd` 和 `etcdctl`：
```bash
which etcd etcdctl
sudo rm -f $(which etcd) $(which etcdctl)
```

清理旧的数据目录（如果不需要保留）：
```bash
sudo rm -rf /var/lib/etcd
```

#### 2. **下载 etcd v3.4.26**
从 etcd 的 GitHub Releases 页面下载 v3.4.26：
```bash
wget https://github.com/etcd-io/etcd/releases/download/v3.4.26/etcd-v3.4.26-linux-amd64.tar.gz
```

#### 3. **解压并安装**
解压下载的文件并将二进制文件移动到系统路径：
```bash
tar -xvf etcd-v3.4.26-linux-amd64.tar.gz
sudo mv etcd-v3.4.26-linux-amd64/etcd* /usr/local/bin/
```

确保文件有执行权限：
```bash
sudo chmod +x /usr/local/bin/etcd /usr/local/bin/etcdctl
```

#### 4. **验证安装**
检查 etcd 和 etcdctl 的版本：
```bash
etcd --version
etcdctl version
```

预期输出类似：
```
etcd Version: 3.4.26
Git SHA: Not provided (use ./build)
Go Version: go1.19.9
Go OS/Arch: linux/amd64

etcdctl version: 3.4.26
API version: 3.4
```

如果仍然看到“No help topic for 'version'”，可能是 PATH 指向了旧版本的 `etcdctl`。检查：
```bash
find / -name etcdctl 2>/dev/null
```

如果有多个版本，删除旧版本或更新 PATH：
```bash
export PATH=/usr/local/bin:$PATH
```

#### 5. **配置 etcd 服务**
1. **创建数据目录**
   ```bash
   sudo mkdir -p /var/lib/etcd
   ```

2. **创建 systemd 服务文件**
   编辑 `/etc/systemd/system/etcd.service`：
   ```bash
   sudo nano /etc/systemd/system/etcd.service
   ```

   添加以下内容：
   ```ini
   [Unit]
   Description=etcd
   Documentation=https://github.com/etcd-io/etcd

   [Service]
   ExecStart=/usr/local/bin/etcd --data-dir=/var/lib/etcd
   Restart=always
   User=etcd

   [Install]
   WantedBy=multi-user.target
   ```

3. **创建 etcd 用户**
   ```bash
   sudo useradd --system --no-create-home etcd
   ```

4. **设置权限**
   ```bash
   sudo chown -R etcd:etcd /var/lib/etcd
   ```

#### 6. **启动 etcd 服务**
1. 重新加载 systemd 配置：
   ```bash
   sudo systemctl daemon-reload
   ```

2. 启动 etcd：
   ```bash
   sudo systemctl start etcd
   ```

3. 设置开机启动：
   ```bash
   sudo systemctl enable etcd
   ```

4. 检查服务状态：
   ```bash
   sudo systemctl status etcd
   ```

确保服务正在运行，没有错误。

#### 7. **测试 etcdctl 命令**
etcd v3.4 使用 v3 API 默认支持 `put` 和 `get`。测试你的命令：
```bash
ETCDCTL_API=3 etcdctl put testkey "Hello, etcd!"
ETCDCTL_API=3 etcdctl get testkey
```

预期输出：
```
testkey
Hello, etcd!
```

如果仍然出现“No help topic”或连接错误，尝试显式指定 endpoints：
```bash
ETCDCTL_API=3 etcdctl --endpoints=http://localhost:2379 put testkey "Hello, etcd!"
ETCDCTL_API=3 etcdctl --endpoints=http://localhost:2379 get testkey
```

#### 8. **Kubernetes 环境特殊处理**
你的环境提示（`k8s-0` 和 `MiniK8S_SE3356_2024-2025-2`）表明你可能在 MiniK8S 或自定义 Kubernetes 环境中。etcd v3.4 常用于较旧的 Kubernetes 版本。如果 etcd 是 Kubernetes 集群的一部分：

1. **检查 etcd pod**：
   ```bash
   kubectl get pods -n kube-system | grep etcd
   ```

2. **获取 etcd 服务地址**：
   ```bash
   kubectl get svc -n kube-system | grep etcd
   ```

3. **使用 etcdctl 连接**：
   如果 etcd 使用 TLS（Kubernetes 默认启用），需要证书：
   ```bash
   ETCDCTL_API=3 etcdctl --endpoints=https://<etcd-service-ip>:2379 \
     --cacert=/etc/kubernetes/pki/etcd/ca.crt \
     --cert=/etc/kubernetes/pki/etcd/server.crt \
     --key=/etc/kubernetes/pki/etcd/server.key \
     put testkey "Hello, etcd!"
   ```

   证书路径可能在 `/etc/kubernetes/pki/etcd` 或其他位置，具体取决于你的集群配置。

4. **进入 etcd 容器（可选）**：
   如果需要直接调试：
   ```bash
   kubectl exec -it -n kube-system etcd-<node-name> -- /bin/sh
   ```

   在容器内运行：
   ```bash
   ETCDCTL_API=3 etcdctl put testkey "Hello, etcd!"
   ```

#### 9. **解决常见问题**
- **“No help topic”错误**：
  - 确认 `etcdctl` 版本是 3.4.26。
  - 如果错误持续，可能是旧版本残留，清理所有 `etcdctl` 并重新安装：
    ```bash
    sudo find / -name etcdctl -exec rm -f {} \;
    sudo mv etcd-v3.4.26-linux-amd64/etcdctl /usr/local/bin/
    ```

- **连接错误**：
  - 确保 etcd 服务运行：
    ```bash
    sudo systemctl status etcd
    ```
  - 检查防火墙：
    ```bash
    sudo ufw allow 2379
    sudo ufw allow 2380
    ```

- **Key not found**：
  - 确认 `put` 命令成功执行。
  - 检查 etcd 数据目录 `/var/lib/etcd` 是否有写入权限。

#### 10. **验证最终结果**
运行以下命令确认一切正常：
```bash
etcdctl version
ETCDCTL_API=3 etcdctl --endpoints=http://localhost:2379 member list
ETCDCTL_API=3 etcdctl put testkey "Hello, etcd!"
ETCDCTL_API=3 etcdctl get test personally
```

如果在 Kubernetes 环境中，使用正确的 `--endpoints` 和证书。

### 额外说明
- **为什么 v3.4**：etcd v3.4 是 Kubernetes 1.18-1.22 的默认版本。如果你的 MiniK8S 项目需要与特定 Kubernetes 版本兼容，v3.4 是一个合理选择。但如果你没有版本限制，建议使用 v3.5 系列以获得更好的性能和修复。
- **备份数据**：在生产环境中，定期备份 `/var/lib/etcd`。
- **集群配置**：如果需要多节点 etcd 集群，请提供更多细节（例如节点数、IP 地址），我可以帮你配置 `--initial-cluster` 等参数。

### 下一步
如果问题仍未解决，请提供以下信息：
- `which etcdctl` 和 `etcdctl version` 的输出。
- `sudo systemctl status etcd` 的输出（如果本地运行 etcd）。
- 是否在 Kubernetes 环境中，etcd 是本地运行还是 pod 中。
- 运行 `ETCDCTL_API=3 etcdctl --endpoints=http://localhost:2379 member list` 的具体错误信息。

我可以继续帮你调试！