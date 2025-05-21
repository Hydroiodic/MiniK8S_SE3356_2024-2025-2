package ipvs_ops

import (
	"log"
	"os/exec"
)

func IPVSADMAddVirtualService(ip string, port int, method string) error {
	output, err := exec.Command("ipvsadm", "-A", "-t", ip+":"+string(port), "-s", method).
		Output()
	if err != nil {
		return err
	}

	log.Printf(string(output))
	return nil
}

func IPVSADMAddRealServer(
	ip string,
	port int,
	endpointIP string,
	endpointPort int,
) error {
	output, err := exec.Command("ipvsadm", "-a", "-t", ip+":"+string(port), "-r", endpointIP+":"+string(endpointPort), "-m").
		Output()
	if err != nil {
		return err
	}

	log.Printf(string(output))
	return nil
}

func IPVSADMDelVirtualService(ip string, port int) error {
	output, err := exec.Command("ipvsadm", "-D", "-t", ip+":"+string(port)).
		Output()
	if err != nil {
		return err
	}

	log.Printf(string(output))
	return nil
}

func IPVSADMDelVirtualServer(
	ip string,
	port int,
	endpointIP string,
	endpointPort int,
) error {
	output, err := exec.Command("ipvsadm", "-d", "-t", ip+":"+string(port), "-r", endpointIP+":"+string(endpointPort)).
		Output()
	if err != nil {
		return err
	}
	log.Printf(string(output))
	return nil
}
