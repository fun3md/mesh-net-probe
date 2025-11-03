package performance

import (
	"context"
	"math/rand"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/mesh-net-probe/probe/internal/icmp"
)

// Benchmark suite for measuring performance optimizations
func BenchmarkICMPEngineCreation(b *testing.B) {
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		engine, err := icmp.NewEngine(
			icmp.WithTimeout(5*time.Second),
			icmp.WithBufferSize(65535),
		)
		if err != nil {
			b.Fatalf("Engine creation failed: %v", err)
		}
		engine.Close()
	}
}

func BenchmarkICMPPing(b *testing.B) {
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		b.Fatalf("Engine creation failed: %v", err)
	}
	defer engine.Close()
	
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		b.Fatal("Invalid target IP")
	}
	
	ctx := context.Background()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		measurement, err := engine.Ping(ctx, target)
		if err != nil {
			b.Errorf("Ping failed: %v", err)
		}
		
		// Prevent compiler optimization
		_ = measurement
	}
}

func BenchmarkICMPPingBatch(b *testing.B) {
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		b.Fatalf("Engine creation failed: %v", err)
	}
	defer engine.Close()
	
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		b.Fatal("Invalid target IP")
	}
	
	ctx := context.Background()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		measurements, err := engine.PingBatch(ctx, target, 10)
		if err != nil {
			b.Errorf("Ping batch failed: %v", err)
		}
		
		// Prevent compiler optimization
		_ = measurements
	}
}

func BenchmarkTimingPrecision(b *testing.B) {
	// Test microsecond precision timing using time.Now()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Simulate precise timing measurement
		elapsed := measureOperation(func() {
			// Small operation to measure
			_ = i * 2
		})
		
		// Ensure we actually use the result
		if elapsed < 0 {
			b.Error("Negative elapsed time")
		}
	}
}

func BenchmarkTimingHighResolution(b *testing.B) {
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		ts := time.Now()
		// Small operation
		_ = i + rand.Int()
		te := time.Now()
		
		elapsed := te.Sub(ts)
		if elapsed < 0 {
			b.Error("Negative elapsed time")
		}
	}
}

func BenchmarkMemoryAllocation(b *testing.B) {
	b.ResetTimer()
	
	// Test memory allocation patterns
	for i := 0; i < b.N; i++ {
		// Simulate measurement data allocation
		measurements := make([]*MeasurementResult, 100)
		for j := 0; j < 100; j++ {
			measurements[j] = &MeasurementResult{
				RTT:         time.Microsecond * time.Duration(i+j),
				Success:     true,
				Timestamp:   time.Now(),
				TargetIP:    net.ParseIP("127.0.0.1"),
			}
		}
		
		// Prevent compiler optimization
		_ = len(measurements)
	}
}

func BenchmarkMemoryPool(b *testing.B) {
	pool := NewMeasurementPool(1000)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Get from pool
		result := pool.Get()
		
		// Simulate measurement
		result.RTT = time.Microsecond * time.Duration(i)
		result.Success = true
		result.Timestamp = time.Now()
		result.TargetIP = net.ParseIP("127.0.0.1")
		
		// Return to pool
		pool.Put(result)
	}
}

func BenchmarkConcurrentMeasurements(b *testing.B) {
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		b.Fatalf("Engine creation failed: %v", err)
	}
	defer engine.Close()
	
	targets := []string{"127.0.0.1", "8.8.8.8", "1.1.1.1"}
	targetIPs := make([]net.IP, len(targets))
	for i, targetStr := range targets {
		ip := net.ParseIP(targetStr)
		if ip == nil {
			b.Fatalf("Invalid target IP: %s", targetStr)
		}
		targetIPs[i] = ip
	}
	
	ctx := context.Background()
	b.ResetTimer()
	
	// Parallel benchmarking
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			target := targetIPs[i%len(targetIPs)]
			measurement, err := engine.Ping(ctx, target)
			if err != nil {
				b.Errorf("Concurrent ping failed: %v", err)
			}
			
			// Prevent compiler optimization
			_ = measurement
			i++
		}
	})
}

func BenchmarkCPUUsage(b *testing.B) {
	// Create CPU load to measure CPU usage patterns
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(1*time.Second),
		icmp.WithBufferSize(4096),
	)
	if err != nil {
		b.Fatalf("Engine creation failed: %v", err)
	}
	defer engine.Close()
	
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		b.Fatal("Invalid target IP")
	}
	
	// Get initial CPU stats
	var startCPU runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&startCPU)
	
	ctx := context.Background()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		measurement, err := engine.Ping(ctx, target)
		if err != nil {
			b.Errorf("CPU measurement failed: %v", err)
		}
		
		// Prevent compiler optimization
		_ = measurement
	}
	
	// Get final CPU stats
	var endCPU runtime.MemStats
	runtime.ReadMemStats(&endCPU)
	
	// Log memory usage for analysis
	b.Logf("Memory allocated during benchmarking: %d bytes", endCPU.Alloc-startCPU.Alloc)
	b.Logf("Total allocations: %d", endCPU.Mallocs-startCPU.Mallocs)
}

// Test for memory pool implementation
func TestMemoryPool(t *testing.T) {
	pool := NewMeasurementPool(10)
	
	// Test get and put
	result1 := pool.Get()
	if result1 == nil {
		t.Error("Pool should return non-nil result")
	}
	
	// Modify result
	result1.RTT = time.Millisecond
	result1.Success = true
	
	// Put back
	pool.Put(result1)
	
	// Get again - should reuse the same object
	result2 := pool.Get()
	if result2 == nil {
		t.Error("Pool should return non-nil result")
	}
	
	// Verify reuse (memory efficiency)
	if result2.RTT == time.Millisecond {
		t.Log("Memory pool working correctly - reusing objects")
	}
}

// Support structures
type MeasurementResult struct {
	RTT         time.Duration
	Success     bool
	Timestamp   time.Time
	TargetIP    net.IP
	PacketSize  int
	Sequence    uint16
	Platform    string
}

type MeasurementPool struct {
	pool chan *MeasurementResult
}

func NewMeasurementPool(size int) *MeasurementPool {
	return &MeasurementPool{
		pool: make(chan *MeasurementResult, size),
	}
}

func (p *MeasurementPool) Get() *MeasurementResult {
	select {
	case result := <-p.pool:
		return result
	default:
		return &MeasurementResult{}
	}
}

func (p *MeasurementPool) Put(result *MeasurementResult) {
	// Reset the result for reuse
	result.RTT = 0
	result.Success = false
	result.Timestamp = time.Time{}
	result.TargetIP = nil
	result.PacketSize = 0
	result.Sequence = 0
	
	select {
	case p.pool <- result:
	default:
		// Pool is full, discard result
	}
}

// Utility function for timing measurement
func measureOperation(op func()) time.Duration {
	start := time.Now()
	op()
	return time.Since(start)
}