package collector

import (
	"log"
	"time"
)

// MetricHandler 定义了采集到数据后的处理回调
type MetricHandler func(metrics NodeMetrics)

// Manager 管理所有的采集器生命周期
type Manager struct {
	stopChan chan struct{}
	handler  MetricHandler
}

func NewManager(handler MetricHandler) *Manager {
	return &Manager{
		stopChan: make(chan struct{}),
		handler:  handler,
	}
}

func (m *Manager) Start() {
	log.Println("Probe collectors starting...")

	go func() {
		ticker := time.NewTicker(1 * time.Second) // 1秒一次高频采集
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metrics := m.collectAll()
				// 回调推入 Ring Buffer，打破循环依赖
				if m.handler != nil {
					m.handler(metrics)
				}

			case <-m.stopChan:
				return
			}
		}
	}()
}

func (m *Manager) collectAll() NodeMetrics {
	var n NodeMetrics
	n.CPU = CollectCPU()
	n.Mem = CollectMem()
	n.Disk = CollectDisk()
	n.Net = CollectNet()
	n.KVM = CollectKVM()
	return n
}

func (m *Manager) Stop() {
	log.Println("Probe collectors stopping...")
	close(m.stopChan)
}
