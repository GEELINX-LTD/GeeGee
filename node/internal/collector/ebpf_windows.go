//go:build windows

package collector

import "log"

// CollectNet 在 Windows 环境下为了能够让编辑器编译通过所设置的桩点
// 直接返回空数据结构。实际上该功能严格依赖 Linux 的 eBPF
func CollectNet() NetMetrics {
	log.Println("[STUB] CollectNet called on Windows. Returning empty eBPF metrics.")
	return NetMetrics{}
}
