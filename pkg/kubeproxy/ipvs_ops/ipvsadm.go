package ipvs_ops

import (
	"fmt"
	"log"
	"os/exec"
)

func IPVSADMAddVirtualService(ip string, port int, method string) error {
	output, err := exec.Command("ipvsadm", "-A", "-t", ip+":"+fmt.Sprint(port), "-s", method).
		Output()
	if err != nil {
		return err
	}

	log.Print(string(output))

	return nil
}

func IPVSADMAddRealServer(
	ip string,
	port int,
	endpointIP string,
	endpointPort int,
) error {
	output, err := exec.Command(
		"ipvsadm",
		"-a",
		"-t",
		ip+":"+fmt.Sprint(port),
		"-r",
		endpointIP+":"+fmt.Sprint(endpointPort),
		"-m",
	).
		Output()
	if err != nil {
		return err
	}

	log.Print(string(output))

	return nil
}

func IPVSADMDelVirtualService(ip string, port int) error {
	output, err := exec.Command("ipvsadm", "-D", "-t", ip+":"+fmt.Sprint(port)).
		Output()
	if err != nil {
		return err
	}

	log.Print(string(output))

	return nil
}

func IPVSADMDelVirtualServer(
	ip string,
	port int,
	endpointIP string,
	endpointPort int,
) error {
	output, err := exec.Command(
		"ipvsadm",
		"-d",
		"-t",
		ip+":"+fmt.Sprint(port),
		"-r",
		endpointIP+":"+fmt.Sprint(endpointPort),
	).
		Output()
	if err != nil {
		return err
	}

	log.Print(string(output))

	return nil
}
