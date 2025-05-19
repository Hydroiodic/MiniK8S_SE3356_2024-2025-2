# Flannel 安装

Flannel的安装包内包含一个`flanneld`可执行文件和一个脚本。

修改默认的`docker0`桥接网络

1. 向`etcd`中储存基础配置
2. 运行`flanneld`服务，自动生成`/run/flannel/subnet.env`
3. 运行`mk-docker-opts.sh`脚本可以生成`Docker`启动所需的选项，位于`/run/docker_opts`。

```sh
DOCKER_OPT_BIP="--bip=10.5.20.1/24"
DOCKER_OPT_IPMASQ="--ip-masq=false"
DOCKER_OPT_MTU="--mtu=1400"
DOCKER_OPTS=" --bip=10.5.20.1/24 --ip-masq=false --mtu=1400"
```

4. 重新加载`systemd`配置

```sh
sudo systemctl daemon-reload

sudo systemctl restart docker
```