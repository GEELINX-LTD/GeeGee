package aggregator

import (
	"sync"
	"time"

	pb "github.com/geelinx-ltd/geegee/api/proto"
	"github.com/geelinx-ltd/geegee/node/internal/collector"
)

// RingBuffer 用于收集并暂存极高频的采集数据，然后在上报周期到来时将其汇算抽样
type RingBuffer struct {
	mu      sync.Mutex
	metrics []collector.NodeMetrics
	nodeID  string
}

func NewRingBuffer(nodeID string) *RingBuffer {
	return &RingBuffer{
		metrics: make([]collector.NodeMetrics, 0, 60),
		nodeID:  nodeID,
	}
}

// Push 放入最新的采集点
func (r *RingBuffer) Push(m collector.NodeMetrics) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics = append(r.metrics, m)
}

// Aggregate 计算并清空缓存区，生成用于网络传输的精简 pb 数据包
func (r *RingBuffer) Aggregate() *pb.ReportRequest {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.metrics) == 0 {
		return nil
	}

	// 简易聚合算法示例：这里我们取滑动窗口内各个指标的最新值（或平均值、极值）。
	// 目前对 CPU 的使用率采用计算平均值，其他取最终值(或累加)为示例方案
	latest := r.metrics[len(r.metrics)-1]

	var maxMicroburst uint64
	for _, m := range r.metrics {
		if m.Net.MicroburstEvents > maxMicroburst {
			maxMicroburst = m.Net.MicroburstEvents
		}
	}

	req := &pb.ReportRequest{
		NodeId:    r.nodeID,
		Timestamp: time.Now().UnixMilli(),
		Cpu: &pb.CPUSummary{
			ModelName: latest.CPU.ModelName,
			Cores:     int32(latest.CPU.Cores),
			Mhz:       latest.CPU.Mhz,
			UsagePerc: latest.CPU.UsagePerc,
			Load1:     latest.CPU.Load1,
			Load5:     latest.CPU.Load5,
			Load15:    latest.CPU.Load15,
		},
		Mem: &pb.MemSummary{
			Total:       latest.Mem.Total,
			Available:   latest.Mem.Available,
			Used:        latest.Mem.Used,
			UsedPercent: latest.Mem.UsedPercent,
			SwapTotal:   latest.Mem.SwapTotal,
			SwapFree:    latest.Mem.SwapFree,
		},
		Disk: &pb.DiskSummary{
			ReadBytes:      latest.Disk.ReadBytes,
			WriteBytes:     latest.Disk.WriteBytes,
			ReadCount:      latest.Disk.ReadCount,
			WriteCount:     latest.Disk.WriteCount,
			IopsInProgress: latest.Disk.IopsInProgress,
		},
		Net: &pb.NetSummary{
			BytesRecv:        latest.Net.BytesRecv,
			BytesSent:        latest.Net.BytesSent,
			PacketsRecv:      latest.Net.PacketsRecv,
			PacketsSent:      latest.Net.PacketsSent,
			MicroburstEvents: maxMicroburst,
		},
		Kvm: &pb.KVMSummary{
			TotalVms:       int32(latest.KVM.TotalVMs),
			ActiveVms:      int32(latest.KVM.ActiveVMs),
			TotalAllocVcpu: int32(latest.KVM.TotalAllocVcpu),
			TotalAllocMem:  latest.KVM.TotalAllocMem,
		},
	}

	// 聚合完毕，清空当前窗口的数据
	r.metrics = make([]collector.NodeMetrics, 0, 60)
	return req
}
