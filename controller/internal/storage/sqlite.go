package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	pb "github.com/geelinx-ltd/geegee/api/proto"
	_ "modernc.org/sqlite" // 纯 Go SQLite 驱动，无 CGO 依赖
)

type SqliteStore struct {
	db            *sql.DB
	retentionDays int
}

// NewSqliteStore 挂载单文件数据库。并自动建表
func NewSqliteStore(dsn string, retentionDays int) (*SqliteStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// 强制启用 WAL 模式提高并发写入性能
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;`); err != nil {
		return nil, fmt.Errorf("failed to enable WAL: %w", err)
	}

	store := &SqliteStore{
		db:            db,
		retentionDays: retentionDays,
	}

	// 初始化核心表结构
	if err := store.initSchema(); err != nil {
		return nil, err
	}

	// 启动后台超期清理协程 (每小时清理一次旧数据)
	go store.cleanupRoutine()

	return store, nil
}

func (s *SqliteStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS nodes (
		node_id TEXT PRIMARY KEY,
		last_seen INTEGER
	);

	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		node_id TEXT,
		timestamp INTEGER,
		cpu_load1 REAL,
		mem_used REAL,
		net_burst INTEGER,
		ping_avg_rtt REAL
	);
	
	-- 聚合索引以加速前端点图渲染
	CREATE INDEX IF NOT EXISTS idx_metrics_node_time ON metrics(node_id, timestamp);
	CREATE INDEX IF NOT EXISTS idx_metrics_time ON metrics(timestamp);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *SqliteStore) Ingest(req *pb.ReportRequest) error {
	now := time.Now().UnixMilli()

	// 1. 更新 Nodes 库表状态 (采用 SQLite Upsert: INSERT ... ON CONFLICT)
	_, err := s.db.Exec(`
		INSERT INTO nodes (node_id, last_seen) 
		VALUES (?, ?) 
		ON CONFLICT(node_id) DO UPDATE SET last_seen=excluded.last_seen;
	`, req.NodeId, now)
	if err != nil {
		return err
	}

	// 2. 解析 Ping
	var avgRtt float64
	if len(req.PingResults) > 0 {
		avgRtt = req.PingResults[0].AvgRttMs
	}

	// 3. 落点库表
	_, err = s.db.Exec(`
		INSERT INTO metrics (node_id, timestamp, cpu_load1, mem_used, net_burst, ping_avg_rtt)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		req.NodeId, req.Timestamp,
		req.Cpu.Load1, req.Mem.UsedPercent,
		req.Net.MicroburstEvents, avgRtt)

	return err
}

func (s *SqliteStore) GetNodes() ([]NodeStatus, error) {
	rows, err := s.db.Query(`SELECT node_id, last_seen FROM nodes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []NodeStatus
	now := time.Now().UnixMilli()
	for rows.Next() {
		var n NodeStatus
		if err := rows.Scan(&n.NodeID, &n.LastSeen); err != nil {
			log.Printf("Sqlite scan node err: %v", err)
			continue
		}
		// 若最近 15 秒存活过则判定 Online
		n.IsOnline = (now - n.LastSeen) < 15000
		list = append(list, n)
	}
	return list, nil
}

func (s *SqliteStore) GetNodeHistory(nodeID string, limit int) ([]MetricSnapshot, error) {
	// 从数据库抽取属于他的过去 N 个流水
	// 若在大单体环境里时间跨度很长，这里我们可以按 ORDER BY DESC 取回再将其 Reverse
	// 但这只是最简单的拉取
	query := `
		SELECT timestamp, cpu_load1, mem_used, net_burst, ping_avg_rtt 
		FROM metrics 
		WHERE node_id = ? 
		ORDER BY timestamp DESC 
		LIMIT ?
	`
	rows, err := s.db.Query(query, nodeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MetricSnapshot
	for rows.Next() {
		var m MetricSnapshot
		if err := rows.Scan(&m.Timestamp, &m.CPULoad1, &m.MemUsed, &m.NetBurst, &m.PingAvgRTT); err != nil {
			continue
		}
		result = append(result, m)
	}

	// DB 里查出来是倒序的（最新时间在最前）。Echarts 需升序喂图
	reverse(result)
	return result, nil
}

// reverse 切片反转辅助函数
func reverse(s []MetricSnapshot) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// cleanupRoutine 自动蒸发老旧纪元
func (s *SqliteStore) cleanupRoutine() {
	if s.retentionDays <= 0 {
		return
	}
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		cutoff := time.Now().AddDate(0, 0, -s.retentionDays).UnixMilli()
		res, err := s.db.Exec(`DELETE FROM metrics WHERE timestamp < ?`, cutoff)
		if err == nil {
			affected, _ := res.RowsAffected()
			if affected > 0 {
				log.Printf("[SQLite Store] Cleaned up %d outdated metric rows (older than %d days)", affected, s.retentionDays)
			}
		} else {
			log.Printf("[SQLite Store] Cleanup error: %v", err)
		}
	}
}
