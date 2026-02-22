package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/geelinx-ltd/geegee/api/proto"
	"github.com/geelinx-ltd/geegee/controller/internal/server"
	"github.com/geelinx-ltd/geegee/controller/internal/storage"
	"google.golang.org/grpc"
)

func main() {
	log.Println("Starting GeeGee Controller...")

	// 监听端口
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 实例化后端 TSDB 及 gRPC 服务
	// 此处后续可根据配置文件调整写入地址
	tsdb := storage.NewTSDB("http://localhost:8428/api/v1/import/prometheus")
	grpcServer := grpc.NewServer()
	probeServer := server.NewGrpcServer(tsdb)

	// 注册服务
	pb.RegisterProbeServiceServer(grpcServer, probeServer)

	go func() {
		log.Println("gRPC Server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// 优雅退出监听
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down GeeGee Controller...")
	grpcServer.GracefulStop()
}
