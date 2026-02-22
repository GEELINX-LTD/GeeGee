package collector

import (
	"log"
	"time"
)

// Manager 管理所有的采集器生命周期
type Manager struct {
	stopChan chan struct{}
}

func NewManager() *Manager {
	return &Manager{
		stopChan: make(chan struct{}),
	}
}

func (m *Manager) Start() {
	log.Println("Probe collectors starting...")

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metrics := m.collectAll()
				// 这里暂时仅仅打印输出，验证结果有效性
				log.Printf("METRICS [CPU]: %.2f%% | [MEM] Used: %.2f%% | [DISK] Read IOPS: %d, Write IOPS: %d\n",
					metrics.CPU.UsagePerc[0], metrics.Mem.UsedPercent, metrics.Disk.ReadCount, metrics.Disk.WriteCount)

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
	return n
}

func (m *Manager) Stop() {
	log.Println("Probe collectors stopping...")
	close(m.stopChan)
}
