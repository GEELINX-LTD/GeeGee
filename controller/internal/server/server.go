package server

import (
	"io"
	"log"

	pb "github.com/geelinx-ltd/geegee/api/proto"
	"github.com/geelinx-ltd/geegee/controller/internal/storage"
)

// GrpcServer 实现了 geegeepb.ProbeServiceServer 接口
type GrpcServer struct {
	pb.UnimplementedProbeServiceServer
	db    *storage.TSDB
	cache storage.Persister
}

func NewGrpcServer(db *storage.TSDB, cache storage.Persister) *GrpcServer {
	return &GrpcServer{
		db:    db,
		cache: cache,
	}
}

// ReportMetrics 接收并处理来自于 Node 端上报的高频汇算数据
func (s *GrpcServer) ReportMetrics(stream pb.ProbeService_ReportMetricsServer) error {
	log.Printf("New streaming connection established from a probe node.")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("Probe node stream closed by client.")
			return nil
		}
		if err != nil {
			log.Printf("Error receiving from stream: %v", err)
			return err
		}

		// 这里处理数据，例如打印或写入时序数据库
		pingCount := len(req.PingResults)
		var avgRtt float64
		if pingCount > 0 {
			avgRtt = req.PingResults[0].AvgRttMs // 仅做演示：打印第一个 Target 的平均延迟
		}

		log.Printf("Recv from Node [%s]: CPU Load1=%.2f, MEM Used=%.2f%%, NET Burst=%d, Pings=%d (Target1 Avg: %.2fms)",
			req.NodeId, req.Cpu.Load1, req.Mem.UsedPercent, req.Net.MicroburstEvents, pingCount, avgRtt)

		if s.cache != nil {
			s.cache.Ingest(req)
		}

		// 异步吸入 TSDB (避免阻塞 gRPC 接收主流)
		if s.db != nil {
			go func(r *pb.ReportRequest) {
				if err := s.db.Ingest(r); err != nil {
					log.Printf("TSDB Ingestion failed: %v", err)
				}
			}(req)
		}

		// 可选：发送下行指令心跳
		err = stream.Send(&pb.ReportResponse{
			Success: true,
			Message: "ok",
		})
		if err != nil {
			log.Printf("Error sending response to stream: %v", err)
			return err
		}
	}
}
