#!/bin/bash

sudo apt update
sudo apt install nfs-kernel-server

sudo mkdir -p /nfs_share
sudo chown nobody:nogroup /nfs_share  # Ubuntu/Debian
sudo chown nfsnobody:nfsnobody /nfs_share  # CentOS/RHEL
sudo chmod 755 /nfs_share

echo "/nfs_share *(rw,sync,no_subtree_check)" | sudo tee -a /etc/exports
sudo exportfs -a

sudo ufw allow from 192.168.1.0/24 to any port nfs

sudo systemctl restart nfs-kernel-server
echo "NFS server setup complete. Share available at /nfs_share"

echo "Current NFS exports:"
sudo exportfs -v