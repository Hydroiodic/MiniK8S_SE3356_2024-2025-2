#!/bin/bash

# 这个脚本用于安装flannel网络插件，需要合适地配置etcd服务的地址，master节点和worker节点的行为也不同
ETCD_ENDPOINTS="192.168.1.6:2379"

if systemctl is-active --quiet flanneld; then
    echo "Flannel is already running. No need to install."
else
    echo "Flannel is not running. Starting installation..."

    # br_netfilter 模块是 Flannel 的关键依赖，必须加载以支持桥接网络的 iptables 规则。
    sudo modprobe br_netfilter
    echo "br_netfilter" | sudo tee /etc/modules-load.d/br_netfilter.conf

    # 下载flannel安装包，解压并复制到 /usr/local/bin/目录下（这个目录已经在PATH里，方便在任何地方启动可执行文件），添加脚本执行权限
    # wget https://github.com/flannel-io/flannel/releases/download/v0.26.7/flannel-v0.26.7-linux-amd64.tar.gz
    mkdir ./flannel_install
    tar -xzvf flannel-v0.26.7-linux-amd64.tar.gz -C "./flannel_install"
    cd ./flannel_install

    sudo cp ./flanneld /usr/local/bin/
    sudo cp ./mk-docker-opts.sh /usr/local/bin/
    sudo chmod 755 /usr/local/bin/mk-docker-opts.sh

    # 创建flanneld的systemd服务文件，用于开机启动
    sudo tee /etc/systemd/system/flanneld.service << EOF
    [Unit]
    Description=Flannel
    After=network.target
    After=network-online.target
    Wants=network-online.target

    [Service]
    # 设置服务启动的命令，注意应该与具体的etcd地址保持一致！
    ExecStart=/usr/local/bin/flanneld --ip-masq --kube-subnet-mgr=false --etcd-endpoints=${ETCD_ENDPOINTS}
    Restart=on-failure

    [Install]
    WantedBy=multi-user.target
EOF

    # 创建这个服务单元文件后，重新加载，启动flanneld服务并设置为开机启动
    sudo systemctl daemon-reload
    sudo systemctl start flanneld
    sudo systemctl enable flanneld

    sudo /usr/local/bin/mk-docker-opts.sh # 在flannel正常运行时，再运行这个tar包自带的脚本，它会提取flannel的实时状态，并生成下面的、用于配置docker的环境变量文件

    # 以下是完整删除旧网络的脚本
    # 获取所有连接到flannel网络的容器的ID
    container_ids=$(docker network inspect -f '{{range $k, $v := .Containers}}{{$v.Name}} {{end}}' flannel)

    # 对于每个容器ID，停止并删除对应的容器
    for id in $container_ids
    do
        docker rm -f $id
    done

    # 删除flannel网络
    docker network rm flannel

    source /run/flannel/subnet.env # 加载这个环境变量文件到当前shell
    # 将FLANNEL_SUBNET=10.5.x.1/24 改为 FLANNEL_SUBNET=10.5.x.0/24
    FLANNEL_SUBNET=$(echo "$FLANNEL_SUBNET" | sed 's|\(\([0-9]\+\.\)\{3\}\)[0-9]\+/24|\10/24|')
    export FLANNEL_SUBNET
    
    # 创建新的网络：flannel
    # 网桥名称为mini-cni0
    docker network create --attachable=true --subnet=${FLANNEL_SUBNET} -o "com.docker.network.driver.mtu"=${FLANNEL_MTU} -o "com.docker.network.bridge.name"="mini-cni0" flannel # 为docker创建一个flannel网络，它使用了这个flannel分配的子网；让这个网络在主机上的设备名称为mini-cni0

    # 查看本节点持有的flannel网段，默认一个节点上可分配256个IP；
    docker network inspect flannel | grep Subnet | awk -F '\"' '{print $4}'

    sudo systemctl daemon-reload
    sudo systemctl restart docker

    # 添加iptables规则，允许flannel网络的流量转发
    # 贼逆天，非得要不可。不然跨容器通信就不通！
    sudo iptables -I FORWARD -j FLANNEL-FWD

    ip addr show mini-cni0 # 查看flannel网络的网桥设备
fi
systemctl status flanneld