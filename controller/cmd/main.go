package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Starting GeeGee Controller...")

	// 占位：启动 gRPC Server, DB 连接等

	// 优雅退出监听
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down GeeGee Controller...")
}
