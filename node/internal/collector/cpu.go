package collector

import (
	"log"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
)

func CollectCPU() CPUMetrics {
	var metrics CPUMetrics

	// 获取基本信息
	info, err := cpu.Info()
	if err != nil {
		log.Printf("Failed to get CPU info: %v", err)
	} else if len(info) > 0 {
		metrics.ModelName = info[0].ModelName
		metrics.Cores = int(info[0].Cores)
		metrics.Mhz = info[0].Mhz
	}

	// 获取使用率，采样 0 表示不等待，获取从上次调用到现在的速率
	usage, err := cpu.Percent(0, true)
	if err != nil {
		log.Printf("Failed to get CPU usage: %v", err)
	} else {
		metrics.UsagePerc = usage
	}

	// 获取 Load
	l, err := load.Avg()
	if err != nil {
		log.Printf("Failed to get CPU load: %v", err)
	} else {
		metrics.Load1 = l.Load1
		metrics.Load5 = l.Load5
		metrics.Load15 = l.Load15
	}

	return metrics
}
