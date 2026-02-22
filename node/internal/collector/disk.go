package collector

import (
	"log"

	"github.com/shirou/gopsutil/v4/disk"
)

func CollectDisk() DiskMetrics {
	var metrics DiskMetrics

	// 测试：获取所有的磁盘 IO 计数并简单的进行聚合
	// 实际生产环境中需要对每个设备进行过滤和存储（例如只返回系统盘或特定的块设备）
	counters, err := disk.IOCounters()
	if err != nil {
		log.Printf("Failed to get disk io counters: %v", err)
	} else {
		for _, stat := range counters {
			metrics.ReadBytes += stat.ReadBytes
			metrics.WriteBytes += stat.WriteBytes
			metrics.ReadCount += stat.ReadCount
			metrics.WriteCount += stat.WriteCount
			metrics.IopsInProgress += stat.IopsInProgress
		}
	}

	return metrics
}
