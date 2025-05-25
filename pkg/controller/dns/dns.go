package dns

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/mqtemplate"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"gopkg.in/yaml.v3"
)

type DnsController struct {
	NginxServiceIp    string
	HostList          []string
	NginxServiceName  string
	NginxPodName      string
	NginxPodNamespace string
	ci                *apiserver.APIClient
}

func (dc *DnsController) Init() {
	dc.ci = apiserver.NewAPIClient("")
	//创建一个nginx pod
	var nginxPod object.Pod

	data, err := os.ReadFile(WorkDir + "/examples/api/nginx-pod.yaml")

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal(data, &nginxPod)

	if err != nil {
		fmt.Println(err)
	}

	err = dc.ci.CreatePod(&nginxPod)

	if err != nil {
		fmt.Println(err)
	}

	dc.NginxPodName = "nginx_pod"
	dc.NginxPodNamespace = "default"
	//创建一个nginx service
	var nginxService object.Service

	s_data, err := os.ReadFile(WorkDir + "/examples/api/nginx-service.yaml")

	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal(s_data, &nginxService)

	if err != nil {
		fmt.Println(err)
	}

	err = dc.ci.CreateService(&nginxService)

	if err != nil {
		fmt.Println(err)
	}

	dc.NginxServiceName = nginxService.Metadata.Name
	dc.NginxServiceIp = nginxService.Status.ClusterIP
	//handle messgae published from api server
	_ = mqtemplate.ConsumeMessageOnQueue(
		mqtemplate.DnsCreatePod,
		dc.handleCreateDns,
	)
	_ = mqtemplate.ConsumeMessageOnQueue(
		mqtemplate.DnsDeletePod,
		dc.handleDeleteDns,
	)
}

func (dc *DnsController) handleCreateDns(msg map[string]interface{}) error {

	var dns object.DNS

	req, err := json.Marshal(msg)

	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}

	err = json.Unmarshal(req, &dns)

	if err != nil {
		fmt.Println("unmarshal request body failed")
		return err
	}

	//根据发来的dns消息，修改host映射
	dc.HostList = append(dc.HostList, dc.NginxServiceIp+" "+dns.Spec.Host)

	WriteNginxConf(dns)

	//重新生成nginx pod, 可能无法重启
	var nginxPod object.Pod

	data, err := os.ReadFile(WorkDir + "/examples/api/nginx-pod.yaml")

	if err != nil {
		fmt.Println(err)
		return err
	}

	err = yaml.Unmarshal(data, &nginxPod)

	if err != nil {
		fmt.Println(err)
		return err
	}

	err = dc.ci.DeletePod(&nginxPod)

	if err != nil {
		fmt.Println(err)
		return err
	}
	//等待2s
	time.Sleep(2 * time.Second)

	err = dc.ci.CreatePod(&nginxPod)

	if err != nil {
		fmt.Println(err)
		return err
	}

	// 直接找到该docker容器，让其执行nginx -s reload
	exec.Command(
		"bash",
		"-c",
		"docker",
		"exec",
		"nginx_pod",
		"sh",
		"-c",
		"nginx -s reload",
	)

	err = dc.ci.CreateDns(&dns)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	return nil
}

func (dc *DnsController) handleDeleteDns(msg map[string]interface{}) error {

	var dns object.DNS

	req, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}

	err = json.Unmarshal(req, &dns)
	if err != nil {
		fmt.Println("unmarshal request body failed")
		return err
	}
	//根据发来的dns消息，修改host映射, 删除dns对应的条目
	for i, host := range dc.HostList {
		if host == dc.NginxServiceIp+" "+dns.Spec.Host {
			dc.HostList = append(dc.HostList[:i], dc.HostList[i+1:]...)
			break
		}
	}

	DeleteNginxConf(dns)
	//重新生成nginx pod
	//重新生成nginx pod, 可能无法重启
	var nginxPod object.Pod

	data, err := os.ReadFile(WorkDir + "/examples/api/nginx-pod.yaml")

	if err != nil {
		fmt.Println(err)
		return err
	}

	err = yaml.Unmarshal(data, &nginxPod)

	if err != nil {
		fmt.Println(err)
		return err
	}

	err = dc.ci.DeletePod(&nginxPod)

	if err != nil {
		fmt.Println(err)
		return err
	}
	//等待2s
	time.Sleep(2 * time.Second)

	err = dc.ci.CreatePod(&nginxPod)

	if err != nil {
		fmt.Println(err)
		return err
	}

	err = dc.ci.DeleteDns(&dns)

	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}

	return nil
}
