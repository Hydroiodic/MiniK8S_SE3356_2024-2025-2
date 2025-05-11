package cadvisorutils

import (
	"fmt"

	utils "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/utils"
)

// 获取某一个container的CPU和MEMORY信息
func GetContainerCPUandMem(
	cAdvisorIP string,
	cAdvisorPort string,
	containerName string,
) (float64, float64, error) {
	targetUrl := fmt.Sprintf(
		"http://%s:%s/api/v1.3/docker/%s",
		cAdvisorIP,
		cAdvisorPort,
		containerName,
	)

	var containerInfo map[string]ContainerInfo

	responseStatus, err := utils.GetRequestWithParamsAndObject(
		targetUrl,
		nil,
		&containerInfo,
	)

	if err != nil {
		fmt.Println("Request failed with err:", err)
		return 0.0, 0.0, err
	}

	if responseStatus != 200 {
		fmt.Println("Request failed with status:", responseStatus)
		return 0.0, 0.0, err
	}

	for _, container := range containerInfo {
		totalCpuUtilization, cpuNums := calculateCpuUsage(container)

		var averageCpuUtilization = 0.0

		if cpuNums != 0 {
			averageCpuUtilization = totalCpuUtilization / cpuNums
		} else {
			fmt.Println("cpu occupied by the container is 0 ")
		}

		memUsePercentage := calculateMemoryUsage(container)

		fmt.Println("Average CPU Usage: ", averageCpuUtilization)
		fmt.Println("Memory Usage Percentage: ", memUsePercentage)

		return averageCpuUtilization, memUsePercentage, nil
	}

	return 0.0, 0.0, fmt.Errorf("no container info found")
}

func calculateCpuUsage(container ContainerInfo) (float64, float64) {
	var totalCpuUtilization, cpuNums = 0.0, 0.0

	for i := 1; i < len(container.Stats); i++ {
		cpuDuration := container.Stats[i].Timestamp.Sub(container.Stats[i-1].Timestamp).
			Seconds()
		cpuUsage := float64(
			container.Stats[i].Cpu.Usage.Total-container.Stats[i-1].Cpu.Usage.Total,
		) / 1e9

		if cpuDuration > 0 {
			cpuNums += 1
			totalCpuUtilization += (cpuUsage / cpuDuration)
		}
	}

	return totalCpuUtilization, cpuNums
}

func calculateMemoryUsage(container ContainerInfo) float64 {
	var totalMemoryUsage uint64 = 0

	// 避免除以零错误
	statCount := len(container.Stats)
	if statCount == 0 {
		fmt.Println("No stats available for the container.")
		return 0.0
	}

	// 计算总内存使用
	for _, stat := range container.Stats {
		totalMemoryUsage += stat.Memory.Usage
	}

	averageMemoryUsage := totalMemoryUsage / uint64(
		len(container.Stats),
	)

	totalMemCapacity, mem_err := utils.GetTotalMemory()
	if mem_err != nil {
		fmt.Println("error when try to get total Memory:", mem_err)
	}

	memUsePercentage := float64(
		averageMemoryUsage,
	) / float64(
		totalMemCapacity,
	)

	return memUsePercentage
}
