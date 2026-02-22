//go:build linux

package collector

import "log"

// CollectNet 在 Linux 下负责加载并读取 eBPF 采集到的网卡流量以及微秒级发包突变
// 真实实现会在单独接入 cilium/ebpf 后编写，这里先搭框架
func CollectNet() NetMetrics {
	log.Println("[eBPF] CollectNet called on Linux. Pending implementation.")
	return NetMetrics{
		BytesRecv:        0,
		BytesSent:        0,
		PacketsRecv:      0,
		PacketsSent:      0,
		MicroburstEvents: 1, // dummy testing value
	}
}
