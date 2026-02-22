package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/geelinx-ltd/geegee/node/internal/collector"
)

func main() {
	log.Println("Starting GeeGee Node Probe...")

	// 初始化与启动采集任务
	mgr := collector.NewManager()
	mgr.Start()

	// 优雅退出监听
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	mgr.Stop()
	log.Println("Shutting down GeeGee Node Probe...")
}
