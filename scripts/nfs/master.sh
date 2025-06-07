#!/bin/bash

sudo apt update
sudo apt install nfs-kernel-server

sudo mkdir -p /nfs_share
sudo chown ubuntu:ubuntu /nfs_share  # Ubuntu/Debian
sudo chmod 777 /nfs_share

# no_root_squash allows root on client to have root access on the NFS share
echo "/nfs_share *(rw,sync,no_subtree_check,no_root_squash)" | sudo tee -a /etc/exports
sudo exportfs -a

sudo ufw allow from 192.168.1.0/24 to any port nfs

sudo systemctl restart nfs-kernel-server
echo "NFS server setup complete. Share available at /nfs_share"

echo "Current NFS exports:"
sudo exportfs -v