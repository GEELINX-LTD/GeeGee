package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/geelinx-ltd/geegee/node/internal/aggregator"
	"github.com/geelinx-ltd/geegee/node/internal/client"
	"github.com/geelinx-ltd/geegee/node/internal/collector"
)

func main() {
	log.Println("Starting GeeGee Node Probe...")

	// 1. 初始化 gRPC 客户端
	grpcClient := client.NewGrpcClient("localhost:50051")
	if err := grpcClient.Connect(); err != nil {
		log.Printf("Failed to connect to controller: %v\n", err)
	}
	defer grpcClient.Close()

	// 2. 初始化边缘计算 Ring Buffer
	ringBuf := aggregator.NewRingBuffer("test-node-windows")

	// 3. 启动定时上报协程 (每 5 秒上报一次)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			req := ringBuf.Aggregate()
			if req != nil {
				_ = grpcClient.SendMetrics(req)
			}
		}
	}()

	// 4. 初始化采集器并使用回调关联 Ring Buffer
	mgr := collector.NewManager(func(m collector.NodeMetrics) {
		ringBuf.Push(m)
	})
	mgr.Start()

	// 5. 优雅退出监听
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	mgr.Stop()
	log.Println("Shutting down GeeGee Node Probe...")
}
