package storage

import (
	"sync"
	"time"

	pb "github.com/geelinx-ltd/geegee/api/proto"
)

// MetricSnapshot 存储前端展示必要的关键维度
type MetricSnapshot struct {
	Timestamp int64   `json:"timestamp"`
	CPULoad1  float64 `json:"cpu_load1"`
	MemUsed   float64 `json:"mem_used_percent"`
	NetBurst  uint64  `json:"net_burst"`

	// 这里保存平均 TCP Rtt (演示取第一个 Target)
	PingAvgRTT float64 `json:"ping_avg_rtt"`
}

type NodeStatus struct {
	NodeID      string           `json:"node_id"`
	LastSeen    int64            `json:"last_seen"` // Unix milli
	IsOnline    bool             `json:"is_online"`
	HistoryFlow []MetricSnapshot `json:"history"` // 环形缓冲，最多 N 条
}

// MemoryCache 提供给前端直接可用的时序缓冲环
type MemoryCache struct {
	mu    sync.RWMutex
	nodes map[string]*NodeStatus
	limit int
}

func NewMemoryCache(limit int) *MemoryCache {
	return &MemoryCache{
		nodes: make(map[string]*NodeStatus),
		limit: limit,
	}
}

func (m *MemoryCache) Ingest(req *pb.ReportRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, exists := m.nodes[req.NodeId]
	if !exists {
		node = &NodeStatus{
			NodeID:      req.NodeId,
			HistoryFlow: make([]MetricSnapshot, 0, m.limit),
		}
		m.nodes[req.NodeId] = node
	}

	node.LastSeen = time.Now().UnixMilli()
	node.IsOnline = true

	var avgRtt float64
	if len(req.PingResults) > 0 {
		avgRtt = req.PingResults[0].AvgRttMs
	}

	snap := MetricSnapshot{
		Timestamp:  req.Timestamp,
		CPULoad1:   req.Cpu.Load1,
		MemUsed:    req.Mem.UsedPercent,
		NetBurst:   req.Net.MicroburstEvents,
		PingAvgRTT: avgRtt,
	}

	// 环式追加
	node.HistoryFlow = append(node.HistoryFlow, snap)
	if len(node.HistoryFlow) > m.limit {
		// 移除最老的一条
		node.HistoryFlow = node.HistoryFlow[1:]
	}
}

// GetNodes 返回所有已知节点当前状态
func (m *MemoryCache) GetNodes() []NodeStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now().UnixMilli()
	list := make([]NodeStatus, 0, len(m.nodes))

	for _, n := range m.nodes {
		// 若 15 秒未上报则当做掉线
		n.IsOnline = (now - n.LastSeen) < 15000
		list = append(list, *n)
	}
	return list
}

// GetNodeHistory 获取指定节点的最新流水线
func (m *MemoryCache) GetNodeHistory(nodeID string) []MetricSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if node, ok := m.nodes[nodeID]; ok {
		// 拷贝切片防止数据并发修改引发前端 JSON 序列化崩溃
		cp := make([]MetricSnapshot, len(node.HistoryFlow))
		copy(cp, node.HistoryFlow)
		return cp
	}
	return nil
}
