package prober

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Target struct {
	IP         string
	Port       int
	TargetType string // "tcpping"
}

type PingResult struct {
	Target     Target
	MinRTTMs   float64
	MaxRTTMs   float64
	AvgRTTMs   float64
	PacketLoss float64 // 0.0 - 1.0
}

// Prober 负责发起对外探测并统计结果
type Prober struct {
	targets []Target
	mu      sync.RWMutex
}

func NewProber() *Prober {
	return &Prober{
		targets: []Target{
			{IP: "8.8.8.8", Port: 53, TargetType: "tcpping"},    // Google DNS
			{IP: "1.1.1.1", Port: 80, TargetType: "tcpping"},    // Cloudflare
			{IP: "223.5.5.5", Port: 443, TargetType: "tcpping"}, // Aliyun DNS
		},
	}
}

// UpdateTargets 由主控端下发新的探测列表
func (p *Prober) UpdateTargets(targets []Target) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.targets = targets
}

// RunPingCycle 并发地对所有 Target 进行测试，每个 target 测指定次数（如 3 次）
func (p *Prober) RunPingCycle(count int, timeout time.Duration) []PingResult {
	p.mu.RLock()
	targets := make([]Target, len(p.targets))
	copy(targets, p.targets)
	p.mu.RUnlock()

	var wg sync.WaitGroup
	results := make([]PingResult, len(targets))

	for i, t := range targets {
		wg.Add(1)
		go func(idx int, target Target) {
			defer wg.Done()
			results[idx] = performTCPPing(target, count, timeout)
		}(i, t)
	}

	wg.Wait()
	return results
}

// performTCPPing 对指定的一个目标执行数次连通测算
func performTCPPing(t Target, count int, timeout time.Duration) PingResult {
	var totalRtt float64
	var minRtt float64 = -1
	var maxRtt float64 = -1
	var failedCount int

	addr := net.JoinHostPort(t.IP, fmt.Sprintf("%d", t.Port))

	for i := 0; i < count; i++ {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", addr, timeout)
		rtt := float64(time.Since(start).Milliseconds()) // 粗略转为毫秒

		if err != nil {
			failedCount++
		} else {
			conn.Close()
			totalRtt += rtt
			if minRtt == -1 || rtt < minRtt {
				minRtt = rtt
			}
			if rtt > maxRtt {
				maxRtt = rtt
			}
		}

		// 避免短时间密集发包被防护拦截，通常每次探测间隔一点点
		time.Sleep(50 * time.Millisecond)
	}

	var avgRtt float64
	successCount := count - failedCount
	if successCount > 0 {
		avgRtt = totalRtt / float64(successCount)
	}
	if minRtt == -1 {
		minRtt = 0
	}
	if maxRtt == -1 {
		maxRtt = 0
	}

	return PingResult{
		Target:     t,
		MinRTTMs:   minRtt,
		MaxRTTMs:   maxRtt,
		AvgRTTMs:   avgRtt,
		PacketLoss: float64(failedCount) / float64(count),
	}
}
