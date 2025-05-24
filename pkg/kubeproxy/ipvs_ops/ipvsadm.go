package ipvs_ops

import (
	"fmt"
	"log"
	"os/exec"
)

func IPVSADMAddVirtualService(ip string, port int, method string) error {
	// TCP
	output, err := exec.Command("ipvsadm", "-A", "-t", ip+":"+fmt.Sprint(port), "-s", method).
		Output()
	if err != nil {
		log.Print(string(output))
		return err
	}

	// UDP
	output, err = exec.Command("ipvsadm", "-A", "-u", ip+":"+fmt.Sprint(port), "-s", method).
		Output()
	if err != nil {
		log.Print(string(output))
		return err
	}

	log.Printf("Added virtual service %s:%d with method %s\n", ip, port, method)

	return nil
}

func IPVSADMAddRealServer(
	ip string,
	port int,
	endpointIP string,
	endpointPort int,
) error {
	// TCP
	output, err := exec.Command(
		"ipvsadm",
		"-a",
		"-t",
		ip+":"+fmt.Sprint(port),
		"-r",
		endpointIP+":"+fmt.Sprint(endpointPort),
		"-m",
	).Output()

	if err != nil {
		log.Print(string(output))
		return err
	}

	// UDP
	output, err = exec.Command(
		"ipvsadm",
		"-a",
		"-u",
		ip+":"+fmt.Sprint(port),
		"-r",
		endpointIP+":"+fmt.Sprint(endpointPort),
		"-m",
	).Output()
	if err != nil {
		log.Print(string(output))
		return err
	}

	log.Printf("Added real server %s:%d for virtual service %s:%d\n",
		endpointIP, endpointPort, ip, port)

	return nil
}

func IPVSADMDelVirtualService(ip string, port int) error {
	// Delete TCP virtual service
	output, err := exec.Command("ipvsadm", "-D", "-t", ip+":"+fmt.Sprint(port)).
		Output()
	if err != nil {
		log.Print(string(output))
		return err
	}

	// Delete UDP virtual service
	output, err = exec.Command("ipvsadm", "-D", "-u", ip+":"+fmt.Sprint(port)).
		Output()
	if err != nil {
		log.Print(string(output))
		return err
	}

	log.Printf("Deleted virtual service %s:%d\n", ip, port)

	return nil
}

func IPVSADMDelVirtualServer(
	ip string,
	port int,
	endpointIP string,
	endpointPort int,
) error {
	// Delete TCP real server
	output, err := exec.Command(
		"ipvsadm",
		"-d",
		"-t",
		ip+":"+fmt.Sprint(port),
		"-r",
		endpointIP+":"+fmt.Sprint(endpointPort),
	).Output()
	if err != nil {
		log.Print(string(output))
		return err
	}

	// Delete UDP real server
	output, err = exec.Command(
		"ipvsadm",
		"-d",
		"-u",
		ip+":"+fmt.Sprint(port),
		"-r",
		endpointIP+":"+fmt.Sprint(endpointPort),
	).Output()
	if err != nil {
		log.Print(string(output))
		return err
	}

	log.Printf("Deleted real server %s:%d for virtual service %s:%d\n",
		endpointIP, endpointPort, ip, port)

	return nil
}
