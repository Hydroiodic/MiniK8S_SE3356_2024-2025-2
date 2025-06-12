a. 演⽰利⽤配置⽂件创建ReplicaSet的配置⽂件与运⾏状况
    1、演⽰的ReplicaSet的配置⽂件中需⾄少包含：ReplicaSet的唯⼀标识符（name）、ReplicaSet对应的Pod、Replica的数⽬

    打开页面
    2、演⽰运⾏状况时需演⽰如何使⽤kubectl创建ReplicaSet，并利⽤kubectl指令展⽰创建的ReplicaSet与Pod，其中，创建的Pod应能应⽤多机部署调度⽅案
     ./kubectl.sh apply -f examples/api/replicaset.yaml 
     ./kubectl.sh get replicasets
     ./kubectl.sh get pods

    3、演⽰通过kubectl相关命令删除ReplicaSet的功能
    ./kubectl.sh delete replicaset my-replicaset    
     ./kubectl.sh get replicasets
     ./kubectl.sh get pods   
b. 演⽰将ReplicaSet绑定⾄Service的配置⽂件和运⾏状况
    1、演⽰的配置⽂件需包含如何配置合适的label和selector，将ReplicaSet绑定⾄Service上的配置
    打开页面
    2、演⽰运⾏状况时需⾃⾏设计场景，展⽰如何使⽤kubectl将Service与ReplicaSet映射⾄⼀起，从⽽使得访问Service的流量能够以⼀定其设置的负载均衡策略被分配到同⼀ReplicaSet内的位于多个节点的不同Pod中
    ./kubectl.sh apply -f examples/api/service-relicaset.yaml
    ./kubectl.sh apply -f examples/api/replicaset.yaml 

    ./kubectl.sh get replicasets
    ./kubectl.sh get pods
    ./kubectl.sh get services

    curl 192.168.1.6:30080
   

    3.1、此处需要介绍使⽤的负载策略的类型，如果使⽤多种，可以只展⽰⼀种，但是需要在⽂档中详细说明所有负载均衡策
     sudo ipvsadm -L -n
    3.2需要展现流量确实被均匀地分发到不同的Pod上（例如，可以让Pod返回⾃⼰的IP）
    curl 192.168.1.6:30080
    sudo ipvsadm -lcn | grep 192.168.1.6:30080
c、演⽰ReplicaSet中Pod停⽌运⾏时，ReplicaSet进⾏恢复的运⾏状况

    1、演⽰时需⾃⾏设计场景（如某⼀Pod中的某⼀容器由于内存⽤量超出资源上限⽽导致被终
    ⽌），展⽰正在运⾏的Pod的数⽬的变化过程
    ./kubectl.sh delete pod
    ./kubectl.sh get pods