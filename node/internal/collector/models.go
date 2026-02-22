package collector

// NodeMetrics 表示单个节点从探针收集上来的基础汇总数据
type NodeMetrics struct {
	CPU  CPUMetrics  `json:"cpu"`
	Mem  MemMetrics  `json:"mem"`
	Disk DiskMetrics `json:"disk"`
}

// CPUMetrics 包含 CPU 相关的信息
type CPUMetrics struct {
	ModelName string    `json:"model_name"`
	Cores     int       `json:"cores"`
	Mhz       float64   `json:"mhz"` // 频率
	UsagePerc []float64 `json:"usage_perc"`
	Load1     float64   `json:"load_1"`
	Load5     float64   `json:"load_5"`
	Load15    float64   `json:"load_15"`
}

// MemMetrics 包含内存相关的信息
type MemMetrics struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapFree    uint64  `json:"swap_free"`
}

// DiskMetrics 包含磁盘 IO 信息 (暂时聚合简化版)
type DiskMetrics struct {
	ReadBytes      uint64 `json:"read_bytes"`
	WriteBytes     uint64 `json:"write_bytes"`
	ReadCount      uint64 `json:"read_count"`
	WriteCount     uint64 `json:"write_count"`
	IopsInProgress uint64 `json:"iops_in_progress"`
}
