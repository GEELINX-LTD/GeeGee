package storage

import (
	"sync"
	"time"

	pb "github.com/geelinx-ltd/geegee/api/proto"
)

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

func (m *MemoryCache) Ingest(req *pb.ReportRequest) error {
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
	return nil
}

// GetNodes 返回所有已知节点当前状态
func (m *MemoryCache) GetNodes() ([]NodeStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now().UnixMilli()
	list := make([]NodeStatus, 0, len(m.nodes))

	for _, n := range m.nodes {
		// 若 15 秒未上报则当做掉线
		n.IsOnline = (now - n.LastSeen) < 15000
		list = append(list, *n)
	}
	return list, nil
}

// GetNodeHistory 获取指定节点的最新流水线
func (m *MemoryCache) GetNodeHistory(nodeID string, limit int) ([]MetricSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if node, ok := m.nodes[nodeID]; ok {
		// 对于内存版，不严格限制截断由于它自带长度管控，原样切片拷贝即可
		cp := make([]MetricSnapshot, len(node.HistoryFlow))
		copy(cp, node.HistoryFlow)
		return cp, nil
	}
	return nil, nil
}
