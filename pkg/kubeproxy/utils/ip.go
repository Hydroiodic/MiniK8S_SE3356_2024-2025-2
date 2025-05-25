package utils

import (
	"fmt"
	"net"
	"strings"
)

// 获取本机的物理网卡IP，区别方式是物理网卡可以连接到外部网络，故建立连接后反过来获取本地IP即可
func GetNodeIP() (string, error) {
	// 创建一个到外部网络的连接
	conn, err := net.Dial("udp", "www.baidu.com:80")
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	defer func() {
		if cerr := conn.Close(); cerr != nil {
			fmt.Println("error closing connection:", cerr)
		}
	}()

	// 获取连接的本地地址
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	// 打印本地IP地址
	fmt.Println(localAddr.IP)

	return strings.TrimSpace(localAddr.IP.String()), nil
}

func GetEnInterfaceIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if strings.HasPrefix(iface.Name, "en") && iface.Flags&net.FlagUp != 0 &&
			iface.Flags&net.FlagLoopback == 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				if ip != nil && ip.To4() != nil {
					fmt.Println("Found en* interface IP:", ip.String())
					return ip.String(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("no en* interface IP found")
}
