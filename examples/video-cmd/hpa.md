
 a. 演⽰利⽤配置⽂件创建⽔平扩容（后⽂简写为HPA）配置，能够对Pod进⾏动态扩容
    ▪演⽰的HPA配置⽂件中需包含：HPA的唯⼀标识符（name），kind，扩容的⽬标workload,扩容的minReplicas和maxReplicas，以及扩容的metrics（⾄少两种，必须包含CPU利⽤率）。
   ./kubectl.sh apply -f examples/api/replicaset.yaml
   ./kubectl.sh apply -f examples/api/horizontal_pod_autoscaler.yaml
   ▪配置完成后能够通过 kubectl get 之类的指令查看到配置的HPA
    ./kubectl.sh get hpas       
   ./kubectl.sh get replicasets   
   ./kubectl.sh get pods
   ▪自⾏设计测试场景，使得HPA的⽬标Pod能够增加或减少负载，从⽽能够触发扩缩容条件
   ▪⾃⾏设计测试场景，使得HPA的⽬标Pod能够增加或减少负载，从⽽能够触发扩缩容条件
   ▪扩缩容所新创建的Pod应能够分布在不同节点中
b. 演⽰扩缩容策略：扩缩容时机
   ▪扩容：给HPA的⽬标Pod增加负载，当负载达到扩容策略metrics规定的值时，增加Pod数量
   直到maxReplicas
   ▪缩容：给HPA的⽬标Pod降低负载，当负载降低⾄规定值时，减少Pod数量直到
   minReplicas
   docker exec -it 265efcc1298e /bin/bash 
   运行apt update && apt install -y stress
   stress --cpu 2 --timeout 60s 

   ▪对能够⽀持的metrics，简单介绍minik8s是如何对metrics进⾏监控的，并且简单介绍
   minik8s如何通过监控的metrics执⾏扩缩容命令的
   cadvisor监控，计算平均值
c 
   •要求必须⽀持对于CPU的监控，并且⽀持的metric的数量⾄少为两种，需要说明额外⽀
   持的metric是哪些
   memory

c. 演⽰扩缩容策略：扩缩容速度
   ▪在演⽰扩缩容时机时，如果扩缩容策略包括了扩缩容速度，那么⼀同演⽰，即在部署的时候
   在配置⽂件中配置好扩容速度的标准，在扩容现象发⽣的时候简单衡量扩容的速度，说明扩
   缩容的速度符合规定
   固定5秒一个pod的速度

d. 演⽰扩缩容后访问⽬标Pod：
   ▪当Pod发⽣扩容后，演⽰通过Pod的Service的IP能够通过⼀定的负载均衡策略访问到扩容后
   的所有Pod，不可以只增加Pod的数量但是扩缩容后的Pod⽆法访问
   ./kubectl.sh apply -f examples/api/service-relicaset.yaml
   curl 192.168.1.6:30080
   sudo ipvsadm -L -n
   sudo ipvsadm -lcn | grep 192.168.1.6:30080