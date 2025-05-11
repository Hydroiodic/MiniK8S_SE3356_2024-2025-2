package utils

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// GetCPUUsage returns the CPU usage percentage.
func GetCPUUsage() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}

	if len(percentages) == 0 {
		return 0, fmt.Errorf("no CPU usage data")
	}

	return percentages[0], nil
}

// GetMemoryUsage returns the memory usage percentage.
func GetMemoryUsage() (float64, error) {
	v, err := mem.VirtualMemory()

	if err != nil {
		return 0, err
	}

	return v.UsedPercent, nil
}

func GetTotalMemory() (float64, error) {
	v, err := mem.VirtualMemory()

	if err != nil {
		return 0, err
	}

	return float64(v.Total), nil
}
