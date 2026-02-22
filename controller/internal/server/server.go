package server

import (
	"io"
	"log"

	pb "github.com/geelinx-ltd/geegee/api/proto"
)

// GrpcServer 实现了 geegeepb.ProbeServiceServer 接口
type GrpcServer struct {
	pb.UnimplementedProbeServiceServer
}

func NewGrpcServer() *GrpcServer {
	return &GrpcServer{}
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
		log.Printf("Recv from Node [%s]: CPU Load1=%.2f, MEM Used=%.2f%%, NET BurstEvents=%d, KVM Total=%d",
			req.NodeId, req.Cpu.Load1, req.Mem.UsedPercent, req.Net.MicroburstEvents, req.Kvm.TotalVms)

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
