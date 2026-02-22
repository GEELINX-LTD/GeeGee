package collector

import (
	"log"

	"github.com/shirou/gopsutil/v4/mem"
)

func CollectMem() MemMetrics {
	var metrics MemMetrics

	vm, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Failed to get virtual memory: %v", err)
	} else {
		metrics.Total = vm.Total
		metrics.Available = vm.Available
		metrics.Used = vm.Used
		metrics.UsedPercent = vm.UsedPercent
	}

	sw, err := mem.SwapMemory()
	if err != nil {
		log.Printf("Failed to get swap memory: %v", err)
	} else {
		metrics.SwapTotal = sw.Total
		metrics.SwapFree = sw.Free
	}

	return metrics
}
