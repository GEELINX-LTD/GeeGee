//go:build windows

package collector

import "log"

// CollectKVM 在 Windows 环境下的空桩点。实际依赖于 libvirt-go，纯 Linux 发行版环境。
func CollectKVM() KVMMetrics {
	log.Println("[STUB] CollectKVM called on Windows. Returning empty KVM metrics.")
	return KVMMetrics{}
}
