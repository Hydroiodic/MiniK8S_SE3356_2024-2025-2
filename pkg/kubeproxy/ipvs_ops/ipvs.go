package ipvs_ops

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/utils"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"

	"github.com/coreos/go-iptables/iptables"
	"github.com/moby/ipvs"
	"gopkg.in/yaml.v3"

	"github.com/vishvananda/netlink"
)

func init() {
	// 获取当前用户的主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Failed to get the user home directory:", err)
		return
	}

	IPTABLES_FILE_PATH = filepath.Join(
		homeDir,
		IPTABLES_FILE_PATH,
	)
	IPVS_FILE_PATH = filepath.Join(homeDir, IPVS_FILE_PATH)
	IPSET_FILE_PATH = filepath.Join(homeDir, IPSET_FILE_PATH)

	// 创建备份文件的目录
	_ = os.MkdirAll(filepath.Dir(IPTABLES_FILE_PATH), 0750)
}

func NewIpvsOps(clusterIPCIDR string) *IpvsOps {
	ops := &IpvsOps{
		ClusterIPCIDR: clusterIPCIDR,
	}
	ops.IptablesClient, _ = iptables.New()
	ops.IpvsClient, _ = ipvs.New("")

	return ops
}

func (ops *IpvsOps) Close() {
	ops.IpvsClient.Close()
	ops.IptablesClient = nil
	ops.IpvsClient = nil
}

func (ops *IpvsOps) Init() {
	// 创建ipvs模式需要的dummy网卡
	_ = createDummyInterface(KUBE_DUMMY_INTERFACE_NAME)

	// 创建表与链名的添加关系
	iptTable2Chains := map[string][]string{}
	iptTable2Chains["nat"] = []string{
		KUBE_SERVICE_CHAIN_NAME,
		KUBE_NODEPORT_CHAIN_NAME,
		KUBE_MARK_MASQ_CHAIN_NAME,
		KUBE_MARK_DROP_CHAIN_NAME,
		KUBE_POSTROUTING_CHAIN_NAME,
		KUBE_FIREWALL_CHAIN_NAME,
	}

	// 创建链，不需要检查是否存在，多次创建幂等
	for table, chains := range iptTable2Chains {
		for _, chain := range chains {
			_ = ops.IptablesClient.NewChain(table, chain)
		}
	}

	// 创建ipset集合，注意每个类型都不同
	// 第一个KUBE-CLUSTER-IP是ClusterIP:port的集合
	// 第二个KUBE-NODE-PORT-TCP是NodePort tcp的集合，为了简单我们只管tcp
	// 第三个KUBE-LOOP-BACK存放endpoints信息，
	// 直接创建出来，不做检查，应该保证命令行输入正确即可

	// ClusterIP:port
	_, err := exec.Command("ipset", "create", KUBE_CLUSTER_IP_SET_NAME, "hash:ip,port").
		Output()
	if err != nil {
		// log.Printf("Failed to execute command: %v", err)
		fmt.Printf(
			"Error in create ipset %s\n",
			KUBE_CLUSTER_IP_SET_NAME,
		)
	}

	// NodePort端口
	_, err = exec.Command("ipset", "create", KUBE_NODE_PORT_TCP_SET_NAME, "bitmap:port", "range", "0-65535").
		Output()
	if err != nil {
		// log.Printf("Failed to execute command: %v", err)
		fmt.Printf(
			"Error in create ipset %s\n",
			KUBE_NODE_PORT_TCP_SET_NAME,
		)
	}

	// 目标IP:目标端口:源IP
	// 存储每个 Endpoint 的三元组：PodIP:PodPort:PodIP。
	// 用于识别 hairpin 场景：当数据包的目标 IP:Port（dstIP:dstPort）是 Pod 自身，且源 IP（srcIP）也是该 Pod IP。
	// 将执行 MASQUERADE，将源 IP 改为节点 IP（如 192.168.1.1）。
	_, err = exec.Command("ipset", "create", KUBE_LOOP_BACK_SET_NAME, "hash:ip,port,ip").
		Output()
	if err != nil {
		fmt.Printf(
			"Error in create ipset %s\n",
			KUBE_LOOP_BACK_SET_NAME,
		)
	}

	/** 请求到达宿主机的网络栈（外部客户端访问ClusterIP:Port或者本机通过OUTPUT访问Service） */
	_ = ops.IptablesClient.AppendUnique(
		"nat",
		"PREROUTING",
		"-j",
		KUBE_SERVICE_CHAIN_NAME,
		"-m",
		"comment",
		"--comment",
		"mini-k8s service portals",
	)
	_ = ops.IptablesClient.AppendUnique(
		"nat",
		"OUTPUT",
		"-j",
		KUBE_SERVICE_CHAIN_NAME,
		"-m",
		"comment",
		"--comment",
		"mini-k8s service portals",
	)

	/** KUBE-SERVICES */
	// 非 ClusterIP 来源的 ClusterIP 流量，标记 SNAT
	// 跳转到 KUBE-MARK-MASQ 打标记
	// 保证数据返回时能正确从ClusterIP而不是PodIP回到客户端
	_ = ops.IptablesClient.AppendUnique(
		"nat",
		KUBE_SERVICE_CHAIN_NAME,
		"!",
		"-s",
		ops.ClusterIPCIDR,
		"-m",
		"set",
		"--match-set",
		KUBE_CLUSTER_IP_SET_NAME,
		"dst,dst",
		"-j",
		KUBE_MARK_MASQ_CHAIN_NAME,
		"-m",
		"comment",
		"--comment",
		"mini-k8s service cluster ip + port for masquerade purpose",
	)

	// 目标IP是本节点IP
	// 如本机Node（Port）访问
	// 跳转到 KUBE-NODEPORT
	_ = ops.IptablesClient.AppendUnique(
		"nat",
		KUBE_SERVICE_CHAIN_NAME,
		"-m",
		"addrtype",
		"--dst-type",
		"LOCAL",
		"-j",
		KUBE_NODEPORT_CHAIN_NAME,
	)

	// ClusterIP下的流量进入IPVS
	_ = ops.IptablesClient.AppendUnique(
		"nat",
		KUBE_SERVICE_CHAIN_NAME,
		"-m",
		"set",
		"--match-set",
		KUBE_CLUSTER_IP_SET_NAME,
		"dst,dst",
		"-j",
		"ACCEPT",
	)

	/** KUBE-NODE-PORT */
	// 如果符合dstPort也符合（Node）Port的访问
	// 那么跳转到KUBE_MARK_MASQ打上0x10000标记
	_ = ops.IptablesClient.AppendUnique(
		"nat",
		KUBE_NODEPORT_CHAIN_NAME,
		"-m",
		"set",
		"--match-set",
		KUBE_NODE_PORT_TCP_SET_NAME,
		"dst",
		"-j",
		KUBE_MARK_MASQ_CHAIN_NAME,
		"-m",
		"comment",
		"--comment",
		"mini-k8s nodeport TCP port for masquerade purpose",
	)

	/** MARK-MASQ */
	// 添加KUBE-MARK-MASQ链的规则，只需要打上0x10000标记
	_ = ops.IptablesClient.AppendUnique(
		"nat",
		KUBE_MARK_MASQ_CHAIN_NAME,
		"-j",
		"MARK",
		"--set-xmark",
		KUBE_MARK_MASQ_VALUE+"/"+KUBE_MARK_MASQ_VALUE,
	)

	// 添加POSTROUTING主链的规则，无条件跳转到KUBE-POSTROUTING链
	_ = ops.IptablesClient.AppendUnique(
		"nat",
		"POSTROUTING",
		"-j",
		KUBE_POSTROUTING_CHAIN_NAME,
		"-m",
		"comment",
		"--comment",
		"mini-k8s postrouting rules",
	)

	/** KUBE-POSTROUTING */
	// 添加KUBE-POSTROUTING链的规则
	// 此处已经由ipvs做好了DNAT
	// 对于所有0x10000标记的包，做SNAT；
	// 如果这个包没有被打上0x10000标记，不管它，直接返回即可
	// 准备发往EndPoint

	// 1. Hairpin 流量 属于回环，采取MASQUERADE进行SNAT
	_ = ops.IptablesClient.AppendUnique(
		"nat",
		KUBE_POSTROUTING_CHAIN_NAME,
		"-m",
		"set",
		"--match-set",
		KUBE_LOOP_BACK_SET_NAME,
		"dst,dst,src",
		"-j",
		"MASQUERADE",
		"-m",
		"comment",
		"--comment",
		"mini-k8s endpoints dst ip:port, source ip for solving hairpin purpose",
	)

	// 2. 无标记，返回主链进行下一条匹配
	_ = ops.IptablesClient.AppendUnique(
		"nat",
		KUBE_POSTROUTING_CHAIN_NAME,
		"-m",
		"mark",
		"!",
		"--mark",
		KUBE_MARK_MASQ_VALUE+"/"+KUBE_MARK_MASQ_VALUE,
		"-j",
		"RETURN",
	)

	// 在NodePort方式下，Kubernetes需要在IP包离开宿主机发往目的Pod时，对源IP进行SNAT处理。
	// 防止拥有Pod的节点直接返回给Client，而不是通过Client访问的NodeIP。
	// 对发往其他Node的网络包去除Tag，对源IP进行SNAT
	_ = ops.IptablesClient.AppendUnique(
		"nat",
		KUBE_POSTROUTING_CHAIN_NAME,
		"-j",
		"MARK",
		"--xor-mark",
		KUBE_MARK_MASQ_VALUE,
	)
	_ = ops.IptablesClient.AppendUnique(
		"nat",
		KUBE_POSTROUTING_CHAIN_NAME,
		"-j",
		"MASQUERADE",
		"-m",
		"comment",
		"--comment",
		"mini-k8s service traffic requiring SNAT",
	)
}

func (ops *IpvsOps) Clear() { // 只删除必要的部分！
	_ = ops.IptablesClient.ClearChain("nat", KUBE_SERVICE_CHAIN_NAME)
	_ = ops.IptablesClient.ClearChain("nat", KUBE_NODEPORT_CHAIN_NAME)
	_ = ops.IptablesClient.ClearChain("nat", KUBE_MARK_MASQ_CHAIN_NAME)
	_ = ops.IptablesClient.ClearChain("nat", KUBE_MARK_DROP_CHAIN_NAME)
	_ = ops.IptablesClient.ClearChain("nat", KUBE_POSTROUTING_CHAIN_NAME)

	// 清除所有ipset
	ipsetSets := []string{
		"mini-KUBE-CLUSTER-IP",
		"mini-KUBE-NODE-PORT-TCP",
		"mini-KUBE-LOOP-BACK",
	}

	for _, set := range ipsetSets {
		cmd := exec.Command("ipset", "flush", set)
		_ = cmd.Run()
		// 删除set
		cmd = exec.Command("ipset", "destroy", set)
		_ = cmd.Run()
	}

	// 清除所有ipvs规则
	_ = ops.IpvsClient.Flush()

	// 清除dummy网卡绑定的所有IP
	_ = clearAllIPsFromDummyInterface(KUBE_DUMMY_INTERFACE_NAME)
}

// 添加一个新的Service、配置相关的iptables, ipvs, ipset
func (ops *IpvsOps) AddService(svc *object.Service) {
	data, _ := yaml.Marshal(&svc)
	fmt.Printf("Add Service \n%s\n\n", string(data))

	// 将clusterIP绑定到dummy网卡
	_ = bindClusterIPToDummyInterface(
		KUBE_DUMMY_INTERFACE_NAME,
		svc.Status.ClusterIP,
	)

	// 添加ClusterIP:port到KUBE-CLUSTER-IP这个ipset
	for _, port := range svc.Spec.Ports {
		cmd := exec.Command(
			"ipset",
			"add",
			KUBE_CLUSTER_IP_SET_NAME,
			svc.Status.ClusterIP+",tcp:"+fmt.Sprint(port.Port),
		)
		fmt.Printf(cmd.String() + "\n")
		err := cmd.Run()

		if err != nil {
			log.Printf(
				"Failed to add clusterIP %s:%d to ipset %s: %v",
				svc.Status.ClusterIP,
				port.Port,
				KUBE_CLUSTER_IP_SET_NAME,
				err,
			)
		}
	}

	// 如果需要，添加NodePort到KUBE-NODE-PORT-TCP这个ipset
	if svc.Type == object.SERVICE_TYPE_NODEPORT_STR {
		for _, port := range svc.Spec.Ports {
			cmd := exec.Command(
				"ipset",
				"add",
				KUBE_NODE_PORT_TCP_SET_NAME,
				fmt.Sprint(port.NodePort),
			)
			err := cmd.Run()

			if err != nil {
				// log.Printf("Failed to add nodePort %d to ipset %s: %v", port.NodePort, KUBE_NODE_PORT_TCP_SET_NAME, err)
				fmt.Printf("")
			}
		}
	}

	// 添加Endpoints到KUBE-LOOP-BACK这个ipset
	for _, ep := range svc.Status.Endpoints {
		if ep.IP == "" {
			continue
		}

		cmd := exec.Command(
			"ipset",
			"add",
			KUBE_LOOP_BACK_SET_NAME,
			ep.IP+",tcp:"+fmt.Sprint(ep.Port)+","+ep.IP,
		)
		fmt.Printf(cmd.String() + "\n")

		err := cmd.Run()
		if err != nil {
			log.Printf(
				"Failed to add endpoint %s:%s:%s to ipset %s: %v",
				ep.IP,
				"tcp:"+fmt.Sprint(ep.Port),
				ep.IP,
				KUBE_LOOP_BACK_SET_NAME,
				err,
			)
		}
	}

	/**配置 ClusterIP 的 ipvs 规则 */
	clusterIP := svc.Status.ClusterIP
	ports := svc.Spec.Ports

	/** ClusterIP 的 IPVS 规则 */
	// 将每个端口绑定到相应的PodIP:PodPort
	for _, port := range ports {
		err := IPVSADMAddVirtualService(clusterIP, port.Port, "rr")
		if err != nil {
			continue
		}

		// 为每一个匹配选择器的Pod创建一个IPVS规则
		for _, endpoint := range svc.Status.Endpoints {
			if endpoint.IP == "" {
				log.Printf(
					"Endpoint IP is empty, skipping service %s:%d",
					clusterIP,
					port.Port,
				)

				continue
			}

			// 需要保证PodPort与TargetPort一致，才对应于这个前端虚服务添加IPVS目标
			if port.TargetPort != endpoint.Port {
				continue
			}

			err := IPVSADMAddRealServer(
				clusterIP,
				port.Port,
				endpoint.IP,
				endpoint.Port,
			)

			if err != nil {
				log.Panicf(
					"Failed to add IPVS destination for %s:%d: %v",
					endpoint.IP,
					endpoint.Port,
					err,
				)
			}
		}
	}

	/** NodePort 的 IPVS 规则 */
	if svc.Type == object.SERVICE_TYPE_NODEPORT_STR {
		nodeIP, _ := utils.GetNodeIP()

		for _, port := range ports { // 也许有一些服务没有暴露NodePort
			if port.NodePort == 0 {
				continue
			}

			err := IPVSADMAddVirtualService(nodeIP, port.NodePort, "rr")
			if err != nil {
				continue
			}

			// 为每一个匹配选择器的Pod创建一个IPVS规则
			for _, endpoint := range svc.Status.Endpoints {
				// 需要保证PodPort与TargetPort一致，才对应于这个前端虚服务添加IPVS目标
				if port.TargetPort != endpoint.Port {
					continue
				}

				err := IPVSADMAddRealServer(
					nodeIP,
					port.NodePort,
					endpoint.IP,
					endpoint.Port,
				)
				if err != nil {
					log.Printf(
						"Failed to add IPVS destination for %s:%d: %v",
						endpoint.IP,
						endpoint.Port,
						err,
					)
				}
			}
		}
	}
}

// 添加的逆过程
func (ops *IpvsOps) DelService(svc *object.Service) {
	data, _ := yaml.Marshal(&svc)
	fmt.Printf("Delete Service %s\n\n", string(data))

	// 解绑ClusterIP
	_ = unbindClusterIPFromDummyInterface(
		KUBE_DUMMY_INTERFACE_NAME,
		svc.Status.ClusterIP,
	)

	// 从KUBE-CLUSTER-IP这个ipset中删除ClusterIP:port
	for _, port := range svc.Spec.Ports {
		cmd := exec.Command(
			"ipset",
			"del",
			KUBE_CLUSTER_IP_SET_NAME,
			svc.Status.ClusterIP+",tcp:"+fmt.Sprint(port.Port),
		)
		fmt.Printf(cmd.String() + "\n")
		err := cmd.Run()

		if err != nil {
			log.Printf(
				"Failed to delete clusterIP %s:%d from ipset %s: %v",
				svc.Status.ClusterIP,
				port.Port,
				KUBE_CLUSTER_IP_SET_NAME,
				err,
			)
		}
	}

	// 如果需要，从KUBE-NODE-PORT-TCP这个ipset中删除NodePort
	if svc.Type == object.SERVICE_TYPE_NODEPORT_STR {
		for _, port := range svc.Spec.Ports {
			cmd := exec.Command(
				"ipset",
				"del",
				KUBE_NODE_PORT_TCP_SET_NAME,
				fmt.Sprint(port.NodePort),
			)

			err := cmd.Run()
			if err != nil {
				// log.Printf("Failed to delete nodePort %d from ipset %s: %v", port.NodePort, KUBE_NODE_PORT_TCP_SET_NAME, err)
				fmt.Printf("")
			}
		}
	}

	// 从KUBE-LOOP-BACK这个ipset中删除Endpoints
	for _, ep := range svc.Status.Endpoints {
		cmd := exec.Command(
			"ipset",
			"del",
			KUBE_LOOP_BACK_SET_NAME,
			ep.IP+",tcp:"+fmt.Sprint(ep.Port)+","+ep.IP,
		)
		fmt.Printf(cmd.String() + "\n")
		err := cmd.Run()

		if err != nil {
			log.Printf(
				"Failed to delete endpoint %s:%s:%s from ipset %s: %v",
				ep.IP,
				"tcp:"+fmt.Sprint(ep.Port),
				ep.IP,
				KUBE_LOOP_BACK_SET_NAME,
				err,
			)
		}
	}

	// 删除关于ClusterIP:port的DNAT规则，此处删除只需要指定ClusterIP:port，而无需对应的Endpoints
	clusterIP := svc.Status.ClusterIP
	ports := svc.Spec.Ports

	for _, port := range ports {
		svc_clusterip_addr := clusterIP + fmt.Sprintf(":%v", port.Port)
		_, err := exec.Command("ipvsadm", "-D", "-t", svc_clusterip_addr).
			Output()

		if err != nil {
			// log.Printf("Failed to delete IPVS service for %s:%d, reason: %v", clusterIP, port.Port, err)
			fmt.Printf("")
			continue
		}
	}

	// 删除关于NodePort的DNAT规则，对应到符合相应targetPort暴露的PodIP:PodPort
	if svc.Type == object.SERVICE_TYPE_NODEPORT_STR {
		nodeIP, _ := utils.GetNodeIP()

		for _, port := range ports {
			if port.NodePort == 0 {
				continue
			}

			err := IPVSADMDelVirtualService(nodeIP, port.NodePort)
			if err != nil {
				// log.Printf("Failed to delete IPVS service for %s:%d: %v", nodeIP, port.NodePort, err)
				fmt.Printf("")
				continue
			}
		}
	}
}

func compareEndpoints(
	oldEndpoints, newEndpoints []object.Endpoint,
) (added, removed []object.Endpoint) {
	oldMap := make(map[string]bool)
	newMap := make(map[string]bool)

	for _, ep := range oldEndpoints {
		oldMap[ep.IP+":"+fmt.Sprint(ep.Port)] = true
	}

	for _, ep := range newEndpoints {
		newMap[ep.IP+":"+fmt.Sprint(ep.Port)] = true

		if !oldMap[ep.IP+":"+fmt.Sprint(ep.Port)] {
			added = append(added, ep)
		}
	}

	for _, ep := range oldEndpoints {
		if !newMap[ep.IP+":"+fmt.Sprint(ep.Port)] {
			removed = append(removed, ep)
		}
	}

	return added, removed
}

// 处理EndPoint变更
func (ops *IpvsOps) UpdateServiceEps(oldSvc, newSvc *object.Service) {
	addedEndpoints, removedEndpoints := compareEndpoints(
		oldSvc.Status.Endpoints,
		newSvc.Status.Endpoints,
	)
	if len(addedEndpoints) == 0 && len(removedEndpoints) == 0 {
		return
	}

	fmt.Printf("Update current service\n")

	// 反向映射targetPort到ServicePort
	clusterIP := newSvc.Status.ClusterIP
	nodeIP, _ := utils.GetNodeIP()

	targetPortMap := make(map[int]object.ServicePort)
	for _, port := range newSvc.Spec.Ports {
		targetPortMap[port.TargetPort] = port
	}

	// endpoints只需要通过targetPort的对应反向索引到Cluster消息即可！
	for _, ep := range removedEndpoints {
		if ep.IP == "" {
			continue
		}

		// 从KUBE-LOOP-BACK这个ipset中删除Endpoint
		err := exec.Command(
			"ipset",
			"del",
			KUBE_LOOP_BACK_SET_NAME,
			ep.IP+",tcp:"+fmt.Sprint(ep.Port)+","+ep.IP,
		).Run()
		if err != nil {
			fmt.Printf("")
		}

		servicePort := targetPortMap[ep.Port]

		// 在ipvs删除ClusterIP:port关于这个ep的DNAT规则
		if servicePort != (object.ServicePort{}) {
			_ = IPVSADMDelVirtualServer(
				clusterIP,
				servicePort.Port,
				ep.IP,
				ep.Port,
			)

			// 如果它具有NodePort规则，一并删掉
			if servicePort.NodePort != 0 {
				_ = IPVSADMDelVirtualServer(
					nodeIP,
					servicePort.NodePort,
					ep.IP,
					ep.Port,
				)
			}
		}
	}

	// 添加新的endpoints
	for _, ep := range addedEndpoints {
		if ep.IP == "" {
			continue
		}

		// 添加到KUBE-LOOP-BACK这个ipset中
		cmd := exec.Command(
			"ipset",
			"add",
			KUBE_LOOP_BACK_SET_NAME,
			ep.IP+",tcp:"+fmt.Sprint(ep.Port)+","+ep.IP,
		)
		_ = cmd.Run()

		servicePort := targetPortMap[ep.Port]

		// 在ipvs添加ClusterIP:port关于这个ep的DNAT规则
		if servicePort != (object.ServicePort{}) {
			err := IPVSADMAddRealServer(
				clusterIP,
				servicePort.Port,
				ep.IP,
				ep.Port,
			)

			if err != nil {
				log.Printf(
					"Failed to add IPVS destination for %s:%d: %v",
					ep.IP,
					ep.Port,
					err,
				)

				continue
			}

			// 如果它具有NodePort，一并添加
			if servicePort.NodePort != 0 {
				err := IPVSADMAddRealServer(
					nodeIP,
					servicePort.NodePort,
					ep.IP,
					ep.Port,
				)
				if err != nil {
					log.Printf(
						"Failed to add IPVS destination for %s:%d: %v",
						ep.IP,
						ep.Port,
						err,
					)
				}
			}
		}
	}
}

// 执行以下命令行时需要sudo权限
func (ops *IpvsOps) SaveToFile(
	iptablesFilePath string,
	ipvsFilePath string,
	ipsetFilePath string,
) error {
	if iptablesFilePath != "" {
		// 保存iptables配置
		iptablesCmd := exec.Command(
			"sh",
			"-c",
			"iptables-save > "+iptablesFilePath,
		)
		err := iptablesCmd.Run()

		if err != nil {
			log.Printf("Failed to save iptables config: %v", err)
			return err
		}
	}

	if ipvsFilePath != "" {
		// 保存ipvs配置
		ipvsCmd := exec.Command("sh", "-c", "ipvsadm -S > "+ipvsFilePath)
		err := ipvsCmd.Run()

		if err != nil {
			log.Printf("Failed to save ipvs config: %v", err)
			return err
		}
	}

	if ipsetFilePath != "" {
		// 保存ipset配置
		ipsetCmd := exec.Command("sh", "-c", "ipset save > "+ipsetFilePath)
		err := ipsetCmd.Run()

		if err != nil {
			log.Printf("Failed to save ipset config: %v", err)
			return err
		}
	}

	return nil
}

func (ops *IpvsOps) RestoreFromFile(
	iptablesFilePath string,
	ipvsFilePath string,
	ipsetFilePath string,
) error {
	// 恢复iptables配置
	if ipsetFilePath != "" {
		iptablesCmd := exec.Command(
			"sh",
			"-c",
			"iptables-restore < "+iptablesFilePath,
		)
		err := iptablesCmd.Run()

		if err != nil {
			log.Printf("Failed to restore iptables config: %v", err)
			return err
		}
	}

	// 恢复ipvs配置
	if ipvsFilePath != "" {
		ipvsCmd := exec.Command("sh", "-c", "ipvsadm -R < "+ipvsFilePath)
		err := ipvsCmd.Run()

		if err != nil {
			log.Printf("Failed to restore ipvs config: %v", err)
			return err
		}
	}

	// 恢复ipset配置
	if ipsetFilePath != "" {
		ipsetCmd := exec.Command("sh", "-c", "ipset restore < "+ipsetFilePath)
		err := ipsetCmd.Run()

		if err != nil {
			log.Printf("Failed to restore ipset config: %v", err)
			return err
		}
	}

	return nil
}

// 创建ipvs模式需要的dummy网卡设备，需要root权限
func createDummyInterface(name string) error {
	if name == "" {
		return nil
	}

	if _, err := netlink.LinkByName(name); err == nil {
		// 如果dummy网卡已经存在，那么不需要再创建
		log.Printf(
			"Dummy interface %s already exists, no need to create",
			name,
		)

		return nil
	}

	// 创建一个新的dummy网卡
	dummy := &netlink.Dummy{
		LinkAttrs: netlink.LinkAttrs{
			Name: name,
		},
	}

	// 添加dummy网卡到系统
	err := netlink.LinkAdd(dummy)
	if err != nil {
		log.Printf("Failed to add dummy interface %s: %v", name, err)
		return err
	}

	// 启动dummy网卡
	err = netlink.LinkSetUp(dummy)
	if err != nil {
		log.Printf("Failed to set up dummy interface %s: %v", name, err)
		return err
	}

	return nil
}

// 绑定ClusterIP到dummy网卡
func bindClusterIPToDummyInterface(name string, clusterIP string) error {
	// 获取dummy网卡
	link, err := netlink.LinkByName(name)
	if err != nil {
		log.Printf("Failed to get dummy interface %s: %v", name, err)
		return err
	}

	// 获取dummy网卡的所有IP地址
	addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		log.Printf(
			"Failed to get IP addresses of dummy interface %s: %v",
			name,
			err,
		)

		return err
	}

	// 检查ClusterIP是否已经绑定到dummy网卡
	for _, addr := range addrs {
		if addr.IPNet.String() == clusterIP { // 如果已绑定，OK
			return nil
		}
	}

	// 绑定ClusterIP到dummy网卡
	fmt.Printf("clusterIP: %s\n", clusterIP)
	addr, err := netlink.ParseAddr(clusterIP + "/32")

	if err != nil {
		log.Printf("Failed to parse ClusterIP %s: %v", clusterIP, err)
		return err
	}

	err = netlink.AddrAdd(link, addr)
	if err != nil {
		log.Printf(
			"Failed to bind ClusterIP %s to dummy interface %s: %v",
			clusterIP,
			name,
			err,
		)

		return err
	}

	return nil
}

// 解绑
func unbindClusterIPFromDummyInterface(name string, clusterIP string) error {
	// 获取dummy网卡
	link, err := netlink.LinkByName(name)
	if err != nil {
		// log.Printf("Failed to get dummy interface %s: %v", name, err)
		return err
	}
	// 获取地址
	addr, err := netlink.ParseAddr(clusterIP + "/32")
	if err != nil {
		// log.Printf("Failed to parse ClusterIP %s: %v", clusterIP, err)
		return err
	}

	// 删除
	err = netlink.AddrDel(link, addr)
	if err != nil {
		// log.Printf("Failed to unbind ClusterIP %s from dummy interface %s: %v", clusterIP, name, err)
		return err
	}

	return nil
}

// 清理dummy网卡上的所有IP
func clearAllIPsFromDummyInterface(name string) error {
	// 获取dummy网卡
	link, err := netlink.LinkByName(name)
	if err != nil {
		// log.Printf("Failed to get dummy interface %s: %v", name, err)
		return err
	}

	// 获取dummy网卡的所有IP地址
	addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		// log.Printf("Failed to get IP addresses of dummy interface %s: %v", name, err)
		return err
	}

	// 遍历所有IP地址并移除
	for _, addr := range addrs {
		err = netlink.AddrDel(link, &addr)
		if err != nil {
			log.Printf(
				"Failed to remove IP %s from dummy interface %s: %v",
				addr.IPNet.String(),
				name,
				err,
			)

			return err
		}
	}

	return nil
}
