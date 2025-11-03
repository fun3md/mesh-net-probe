// Performance profiling configuration for measurement-critical code paths
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// PerformanceProfiler manages performance monitoring for critical measurement paths
type PerformanceProfiler struct {
	profilePath string
	enabled     bool
}

// NewPerformanceProfiler creates a new performance profiler
func NewPerformanceProfiler(profilePath string, enabled bool) *PerformanceProfiler {
	return &PerformanceProfiler{
		profilePath: profilePath,
		enabled:     enabled,
	}
}

// ProfileICMPMeasurement profiles ICMP measurement critical path
func (p *PerformanceProfiler) ProfileICMPMeasurement(ctx context.Context, measurementFunc func() error) error {
	if !p.enabled {
		return measurementFunc()
	}

	// Start CPU profiling
	cpuFile, err := os.Create(fmt.Sprintf("%s/cpu_prof_%d.prof", p.profilePath, time.Now().Unix()))
	if err != nil {
		return fmt.Errorf("failed to create CPU profile: %w", err)
	}
	defer cpuFile.Close()

	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}
	defer pprof.StopCPUProfile()

	// Record memory baseline
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Run measurement function
	err = measurementFunc()

	// Record memory after
	runtime.ReadMemStats(&m2)

	// Generate memory profile
	if err := p.generateMemoryProfile(&m1, &m2); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to generate memory profile: %v\n", err)
	}

	// Generate goroutine profile
	if err := p.generateGoroutineProfile(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to generate goroutine profile: %v\n", err)
	}

	return err
}

func (p *PerformanceProfiler) generateMemoryProfile(before, after *runtime.MemStats) error {
	heapFile, err := os.Create(fmt.Sprintf("%s/heap_prof_%d.prof", p.profilePath, time.Now().Unix()))
	if err != nil {
		return err
	}
	defer heapFile.Close()

	runtime.GC()

	if err := pprof.WriteHeapProfile(heapFile); err != nil {
		return err
	}

	fmt.Printf("Memory usage before: Alloc=%d MB, Sys=%d MB\n",
		before.Alloc/1024/1024, before.Sys/1024/1024)
	fmt.Printf("Memory usage after:  Alloc=%d MB, Sys=%d MB\n",
		after.Alloc/1024/1024, after.Sys/1024/1024)
	fmt.Printf("Memory delta: Alloc=%d MB, Sys=%d MB\n",
		(after.Alloc-before.Alloc)/1024/1024, (after.Sys-before.Sys)/1024/1024)

	return nil
}

func (p *PerformanceProfiler) generateGoroutineProfile() error {
	goroutineFile, err := os.Create(fmt.Sprintf("%s/goroutine_prof_%d.prof", p.profilePath, time.Now().Unix()))
	if err != nil {
		return err
	}
	defer goroutineFile.Close()

	if err := pprof.Lookup("goroutine").WriteTo(goroutineFile, 0); err != nil {
		return err
	}

	return nil
}

// ProfileLatencyMeasurement measures timing precision for ICMP measurements
func ProfileLatencyMeasurement(target string, samples int) {
	fmt.Printf("Profiling latency measurement for %s (%d samples)\n", target, samples)

	// Create timing profiler
	profiler := NewPerformanceProfiler("profiles", true)

	// Simulate ICMP measurement loop
	measurementFunc := func() error {
		for i := 0; i < samples; i++ {
			start := time.Now()

			// Simulate measurement work
			time.Sleep(10 * time.Millisecond)

			elapsed := time.Since(start)
			if i%10 == 0 {
				fmt.Printf("Sample %d: %v\n", i, elapsed)
			}
		}
		return nil
	}

	ctx := context.Background()
	if err := profiler.ProfileICMPMeasurement(ctx, measurementFunc); err != nil {
		fmt.Fprintf(os.Stderr, "Profiling failed: %v\n", err)
	}

	fmt.Println("Performance profiling completed. Check profiles/ directory for results.")
}

func main() {
	// Example usage
	if len(os.Args) > 1 {
		target := os.Args[1]
		samples := 100
		if len(os.Args) > 2 {
			if s, err := fmt.Sscanf(os.Args[2], "%d", &samples); err != nil || s != 1 {
				fmt.Printf("Invalid samples count: %s\n", os.Args[2])
				os.Exit(1)
			}
		}
		ProfileLatencyMeasurement(target, samples)
	} else {
		fmt.Println("Usage: profiler <target> [samples]")
		os.Exit(1)
	}
}
