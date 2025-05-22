package dns

import (
	"fmt"

	"os"
	"strconv"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

const WorkDir = "/home/ubuntu/wang/MiniK8S_SE3356_2024-2025-2"

func WriteNginxConf(dns object.DNS) {
	ci := apiserver.NewAPIClient("http://localhost:8080")
	filePath := WorkDir + "/assets/nginxconf/" + dns.Spec.Host + ".conf"
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

	if err != nil {
		fmt.Println("open file failed")
		fmt.Println(err.Error())

		return
	}

	defer file.Close()

	_, _ = file.WriteString("server {\n")
	_, _ = file.WriteString("    listen 80;\n")
	_, _ = file.WriteString("    server_name " + dns.Spec.Host + ";\n")

	for _, p := range dns.Spec.Paths {
		s, err := ci.GetServiceByName(p.ServiceName)
		if err != nil {
			fmt.Println("invalid service" + p.ServiceName)
		}

		_, _ = file.WriteString("    location " + p.Path + " {\n")
		_, _ = file.WriteString(
			"        proxy_pass http://" + s.Status.ClusterIP + ":" + strconv.Itoa(
				p.ServicePort,
			) + "/;\n",
		)
		_, _ = file.WriteString("    }\n")
	}

	_, _ = file.WriteString("}\n")
}

func DeleteNginxConf(dns object.DNS) {
	fmt.Println("deleteNginxConf")

	filePath := WorkDir + "/assets/nginxconf/" + dns.Spec.Host + ".conf"
	err := os.Remove(filePath)

	if err != nil {
		fmt.Println("delete file failed")
		fmt.Println(err.Error())

		return
	}
}
