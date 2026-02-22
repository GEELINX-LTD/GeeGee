//go:build linux

package collector

import "log"

// CollectKVM 负责通过 libvirt 或者暴露接口读取真实宿主机里的 KVM 规格与目前占用情况
func CollectKVM() KVMMetrics {
	log.Println("[KVM Libvirt] CollectKVM called on Linux. Pending implementation.")
	return KVMMetrics{
		TotalVMs:       10, // dummy testing value
		ActiveVMs:      5,  // dummy testing value
		TotalAllocVcpu: 20, // dummy testing value
		TotalAllocMem:  32 * 1024 * 1024 * 1024,
	}
}
