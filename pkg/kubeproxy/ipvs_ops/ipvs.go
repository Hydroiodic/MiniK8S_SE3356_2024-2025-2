package ipvs_ops

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/utils"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"

	"github.com/coreos/go-iptables/iptables"
	"github.com/moby/ipvs"
	"gopkg.in/yaml.v3"

	"github.com/vishvananda/netlink"
)

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
}

func (ops *IpvsOps) Clear() { // 只删除必要的部分！	// 清除dummy网卡绑定的所有IP
	_ = clearAllIPsFromDummyInterface(KUBE_DUMMY_INTERFACE_NAME)

	// 清除所有的IPVS规则
	output, err := exec.Command(
		"ipvsadm",
		"--clear",
	).Output()
	if err != nil {
		log.Print(string(output))
	}
}

// 添加一个新的Service、配置相关的ipvs
func (ops *IpvsOps) AddService(svc *object.Service) {
	data, _ := yaml.Marshal(&svc)
	fmt.Printf("Add Service \n%s\n\n", string(data))

	// 将clusterIP绑定到dummy网卡
	_ = bindClusterIPToDummyInterface(
		KUBE_DUMMY_INTERFACE_NAME,
		svc.Status.ClusterIP,
	)

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

		log.Printf(
			"Removed IP %s from dummy interface %s",
			addr.IPNet.String(),
			name,
		)
	}

	return nil
}
