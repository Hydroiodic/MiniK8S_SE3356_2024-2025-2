#!/bin/bash

set -e

if [ -z "$NFS_SERVER" ]; then
    echo "Error: NFS_SERVER environment variable is not set."
    exit 1
fi

sudo apt update
sudo apt install -y nfs-common

sudo mkdir -p /nfs_share
sudo mount "${NFS_SERVER}:/nfs_share" /nfs_share

df -h