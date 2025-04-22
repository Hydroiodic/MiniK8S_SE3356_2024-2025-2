tar -xzvf pkg/flannel/flannel-v0.26.7-linux-amd64.tar.gz -C pkg/flannel
cp pkg/flannel/{flanneld,mk-docker-opts.sh} pkg/flannel/bin

export ETCDCTL_API=3
etcdctl put /coreos.com/network/config '{ "Network": "10.244.0.0/16", "Backend": {"Type": "vxlan"}}'

FLANNEL_START_FILE="/etc/systemd/system/flannel.service"
echo "[Unit]" > $FLANNEL_START_FILE
echo "Description=flannel_start" >> $FLANNEL_START_FILE
echo "" >> $FLANNEL_START_FILE
echo "[Service]" >> $FLANNEL_START_FILE
echo "ExecStart=/home/wlx/Desktop/MiniK8S_SE3356_2024-2025-2/pkg/flannel/flanneld" >> $FLANNEL_START_FILE
echo "User=root" >> $FLANNEL_START_FILE

systemctl daemon-reload
systemctl start flannel