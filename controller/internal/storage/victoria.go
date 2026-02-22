package storage

import (
	"log"

	pb "github.com/geelinx-ltd/geegee/api/proto"
)

// TSDB 封装对时序存储库接口。当前我们仅保留打桩与结构，以便日后轻松填入 HTTP Client。
type TSDB struct {
	addr string
}

func NewTSDB(addr string) *TSDB {
	return &TSDB{
		addr: addr,
	}
}

// Ingest 用于将探针产生的 protobuf struct 转化为 Prometheus Text 协议存起留用
func (t *TSDB) Ingest(req *pb.ReportRequest) error {
	// 组装格式样例：
	// metric_name{label1="val1", label2="val2"} value timestamp
	// 例如：node_cpu_usage{node="test-node"} 42.1 1680000000000

	// 在没有实际部署 VictoriaMetrics 环境下，我们此处将接收到的结构抽象化，并不真实发包
	metricCount := 1 + 1 + len(req.PingResults) + 1 // CPU + MEM + PING*n + DISK 粗估项
	log.Printf("[TSDB Ingestion Stub] Prepared %d PromText lines for node [%s] to write locally/remotely.", metricCount, req.NodeId)

	// TODO: 使用 strings.Builder 拼接文本并使用 http.Post 发布到 t.addr (通常是 http://VM_IP:8428/api/v1/import/prometheus)

	return nil
}
