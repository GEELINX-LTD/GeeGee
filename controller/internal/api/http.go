package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/geelinx-ltd/geegee/controller/internal/storage"
)

// HttpServer 构建 RESTful API 并暴露 /api 节点用于前端画图读取
type HttpServer struct {
	addr  string
	cache *storage.MemoryCache
}

func NewHttpServer(addr string, cache *storage.MemoryCache) *HttpServer {
	return &HttpServer{
		addr:  addr,
		cache: cache,
	}
}

func (s *HttpServer) Start() {
	mux := http.NewServeMux()

	// API 1: 列出当前环境已知所有节点的卡片信息
	mux.HandleFunc("/api/nodes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		nodes := s.cache.GetNodes()
		if err := json.NewEncoder(w).Encode(nodes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// API 2: 根据 Node ID 拉取该探针历史折线
	mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		nodeID := r.URL.Query().Get("node_id")
		if nodeID == "" {
			http.Error(w, "missing node_id", http.StatusBadRequest)
			return
		}

		history := s.cache.GetNodeHistory(nodeID)
		if err := json.NewEncoder(w).Encode(history); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Web Static Server: / 将作为前端网页托管根路径
	// 开发期间，我们先用一个极其简单的文字做打桩，下一个阶段直接构建静态页面。
	mux.Handle("/", http.FileServer(http.Dir("./web/static")))

	log.Printf("HTTP Dashboard API Server listening on %s", s.addr)
	if err := http.ListenAndServe(s.addr, mux); err != nil {
		log.Fatalf("HTTP Server failed: %v", err)
	}
}
