<!-- 使用IP-IP模式

给FORWARD链配上`FLANNEL-FWD`链
```
sudo iptables -I FORWARD -j FLANNEL-FWD
```

https://docs.docker.com/engine/network/packet-filtering-firewalls/#docker-and-iptables-chains

Docker的问题？

```sh
$ sudo iptables -L FORWARD -v            
Chain FORWARD (policy DROP 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination         
   20  1836 FLANNEL-FWD  all  --  any    any     anywhere             anywhere            
  269 17928 DOCKER-USER  all  --  any    any     anywhere             anywhere            
  249 16092 DOCKER-FORWARD  all  --  any    any     anywhere             anywhere            
    0     0 FLANNEL-FWD  all  --  any    any     anywhere             anywhere             /* flanneld forward */

# ubuntu @ k8s-1 in ~/liu/MiniK8S_SE3356_2024-2025-2 on git:feat/multiple_node x [16:14:41] 
$ sudo iptables -L FLANNEL-FWD   -v
Chain FLANNEL-FWD (2 references)
 pkts bytes target     prot opt in     out     source               destination         
   50  4590 ACCEPT     all  --  any    any     101.6.0.0/16         anywhere             /* flanneld forward */
    0     0 ACCEPT     all  --  any    any     anywhere             101.6.0.0/16         /* flanneld forward */
``` -->

仍旧使用 VXLAN

内核需要配置以允许转发。