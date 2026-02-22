package storage

import (
	pb "github.com/geelinx-ltd/geegee/api/proto"
)

// MetricSnapshot 这是通用吐出给 API/外部图表的时序折线扁平结构
// (这里为了兼容将 MemoryCache 中的同名声明抽出转移至此)
type MetricSnapshot struct {
	Timestamp  int64   `json:"timestamp"`
	CPULoad1   float64 `json:"cpu_load1"`
	MemUsed    float64 `json:"mem_used_percent"`
	NetBurst   uint64  `json:"net_burst"`
	PingAvgRTT float64 `json:"ping_avg_rtt"`
}

type NodeStatus struct {
	NodeID      string           `json:"node_id"`
	LastSeen    int64            `json:"last_seen"` // Unix milli
	IsOnline    bool             `json:"is_online"`
	HistoryFlow []MetricSnapshot `json:"history"` // 图表缓冲数据
}

// Store 统一后端持久层行为定义，不管挂载内存、SQLite还是远端维多利亚系列，都走这里
type Persister interface {
	// 每次高刷包上门就收
	Ingest(req *pb.ReportRequest) error

	// API 层取卡片列表
	GetNodes() ([]NodeStatus, error)

	// API 层获取指定节点最近 N 个时间切片用于画图
	GetNodeHistory(nodeID string, limit int) ([]MetricSnapshot, error)
}
