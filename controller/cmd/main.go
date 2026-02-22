package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/geelinx-ltd/geegee/api/proto"
	"github.com/geelinx-ltd/geegee/controller/config"
	"github.com/geelinx-ltd/geegee/controller/internal/api"
	"github.com/geelinx-ltd/geegee/controller/internal/server"
	"github.com/geelinx-ltd/geegee/controller/internal/storage"
	"google.golang.org/grpc"
)

func main() {
	log.Println("Starting GeeGee Controller...")

	// 0. 加载外部配置
	config.LoadConfig("config.yaml")
	cfg := config.Cfg

	// 监听端口
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 1. 初始化统一持久化接口工厂
	var persister storage.Persister

	if cfg.Storage.Type == "sqlite" {
		log.Println("Using SQLite Storage Engine...")
		sqliteDB, err := storage.NewSqliteStore(cfg.Storage.Sqlite.Dsn, cfg.Storage.RetentionDays)
		if err != nil {
			log.Fatalf("Failed to open SQLite: %v", err)
		}
		persister = sqliteDB

	} else if cfg.Storage.Type == "victoria" {
		log.Println("Using VictoriaMetrics Storage Engine... (Routing back to original MemoryCache layout + TSDB Push stub)")
		// 为了给 Victoria 用户同样的前端体验，保留内存环路
		memCache := storage.NewMemoryCache(300)
		persister = memCache
	} else {
		log.Println("Unknown storage type, fallback to Memory-Only.")
		persister = storage.NewMemoryCache(300)
	}

	// 2. 实例化 API 服务供大屏调用
	httpApi := api.NewHttpServer(cfg.Http.Port, persister)
	go httpApi.Start()

	// 3. 实例化 gRPC 接收端
	grpcServer := grpc.NewServer()
	// 旧版的 NewGrpcServer 目前只接受单一的 persister 或者 (db, memCache)
	// 我们已经抽象化了存储接口，因此同级重构 `server.NewGrpcServer`
	probeServer := server.NewGrpcServer(nil, persister)

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
