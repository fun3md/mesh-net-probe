package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"golang.org/x/net/icmp"
)

// PrecisionBenchmark provides microsecond-level timing benchmarks
type PrecisionBenchmark struct {
	targets  []string
	duration time.Duration
	samples  int
}

// BenchmarkResult represents timing measurement results
type BenchmarkResult struct {
	Target       string
	MinLatency   time.Duration
	MaxLatency   time.Duration
	AvgLatency   time.Duration
	StdDeviation time.Duration
	PacketLoss   float64
	TotalSamples int
}

// RunPrecisionBenchmark executes high-resolution timing measurements
func (p *PrecisionBenchmark) RunPrecisionBenchmark(ctx context.Context) ([]BenchmarkResult, error) {
	var results []BenchmarkResult

	for _, target := range p.targets {
		fmt.Printf("Benchmarking target: %s\n", target)
		result, err := p.benchmarkSingleTarget(ctx, target)
		if err != nil {
			return nil, fmt.Errorf("benchmark failed for target %s: %w", target, err)
		}
		results = append(results, result)
	}

	return results, nil
}

func (p *PrecisionBenchmark) benchmarkSingleTarget(ctx context.Context, target string) (BenchmarkResult, error) {
	var latencies []time.Duration
	var successfulPackets, totalPackets int

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return BenchmarkResult{}, err
	}
	defer conn.Close()

	startTime := time.Now()
	timeout := time.After(p.duration)

	for i := 0; i < p.samples; i++ {
		select {
		case <-ctx.Done():
			return BenchmarkResult{}, ctx.Err()
		case <-timeout:
			goto end
		default:
		}

		// Send ICMP packet and measure timing
		latency, err := p.measureSinglePing(ctx, conn, target)
		if err != nil {
			totalPackets++
			continue
		}

		latencies = append(latencies, latency)
		successfulPackets++
		totalPackets++

		// Small delay between pings to avoid overwhelming network
		time.Sleep(10 * time.Millisecond)
	}

end:
	if len(latencies) == 0 {
		return BenchmarkResult{
			Target:       target,
			MinLatency:   0,
			MaxLatency:   0,
			AvgLatency:   0,
			StdDeviation: 0,
			PacketLoss:   100.0,
			TotalSamples: totalPackets,
		}, nil
	}

	// Calculate statistics
	var sum time.Duration
	minLatency, maxLatency := latencies[0], latencies[0]

	for _, lat := range latencies {
		sum += lat
		if lat < minLatency {
			minLatency = lat
		}
		if lat > maxLatency {
			maxLatency = lat
		}
	}

	avgLatency := sum / time.Duration(len(latencies))

	// Calculate standard deviation
	var varianceSum float64
	for _, lat := range latencies {
		diff := float64(lat - avgLatency)
		varianceSum += diff * diff
	}
	stdDev := time.Duration(math.Sqrt(varianceSum / float64(len(latencies))))

	packetLoss := float64(totalPackets-successfulPackets) / float64(totalPackets) * 100.0

	return BenchmarkResult{
		Target:       target,
		MinLatency:   minLatency,
		MaxLatency:   maxLatency,
		AvgLatency:   avgLatency,
		StdDeviation: stdDev,
		PacketLoss:   packetLoss,
		TotalSamples: totalPackets,
	}, nil
}

func (p *PrecisionBenchmark) measureSinglePing(ctx context.Context, conn *icmp.PacketConn, target string) (time.Duration, error) {
	start := time.Now()

	// Create ICMP echo request
	msg := &icmp.Message{
		Type: icmp.TypeEchoRequest,
		Code: 0,
		Body: &icmp.Echo{
			ID:   1234,
			Seq:  1,
			Data: []byte("mesh-probe-precision-test"),
		},
	}

	// Marshal the message
	data, err := msg.Marshal(nil)
	if err != nil {
		return 0, err
	}

	// Send packet with timeout
	sendCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := conn.WriteTo(data, &target); err != nil {
		return 0, err
	}

	// Wait for response with timeout
	buffer := make([]byte, 1500)
	err = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return 0, err
	}

	n, _, err := conn.ReadFrom(buffer)
	if err != nil {
		return 0, err
	}

	// Parse response
	resp, err := icmp.ParseMessage(1, buffer[:n])
	if err != nil {
		return 0, err
	}

	if resp.Type != icmp.TypeEchoReply {
		return 0, fmt.Errorf("unexpected ICMP response type: %v", resp.Type)
	}

	elapsed := time.Since(start)
	return elapsed, nil
}

// FormatResults provides formatted output for benchmark results
func FormatResults(results []BenchmarkResult) string {
	var output string
	output += fmt.Sprintf("Precision Benchmark Results (%d targets)\n", len(results))
	output += fmt.Sprintf("%s\n", strings.Repeat("=", 60))

	for _, result := range results {
		output += fmt.Sprintf("Target: %s\n", result.Target)
		output += fmt.Sprintf("  Min Latency:    %v\n", result.MinLatency)
		output += fmt.Sprintf("  Max Latency:    %v\n", result.MaxLatency)
		output += fmt.Sprintf("  Avg Latency:    %v\n", result.AvgLatency)
		output += fmt.Sprintf("  Std Deviation:  %v\n", result.StdDeviation)
		output += fmt.Sprintf("  Packet Loss:    %.1f%%\n", result.PacketLoss)
		output += fmt.Sprintf("  Total Samples:  %d\n", result.TotalSamples)
		output += fmt.Sprintf("%s\n", strings.Repeat("-", 40))
	}

	return output
}
