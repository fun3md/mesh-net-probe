package integration

import (
	"context"
	"net"
	"testing"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/mesh-net-probe/probe/internal/telemetry"
	"github.com/mesh-net-probe/probe/pkg/types"
)

// TestOpenTelemetryIntegration tests end-to-end OpenTelemetry integration
func TestOpenTelemetryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping OpenTelemetry integration test in short mode")
	}

	ctx := context.Background()

	// Initialize OpenTelemetry (default provider)
	otel.SetMeterProvider(nil)
	otel.SetTracerProvider(nil)

	// Create telemetry components
	meterProvider := telemetry.NewOTELMeterProvider()
	tracerProvider := telemetry.NewOTELTracerProvider()

	// Test meter provider initialization
	err := meterProvider.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize meter provider: %v", err)
	}

	t.Logf("OpenTelemetry meter provider initialized successfully")

	// Test measurement recording with tracing
	measurementSpanCtx, measurementSpan := tracerProvider.StartMeasurementSpan(
		ctx, "test-measurement-001", "test-probe-001", "test-target-001",
	)

	// Create test measurement data
	measurement := &types.MeasurementData{
		ID:          "test-001",
		ProbeID:     "test-probe-001",
		Target:      createTestTarget("test-target-001", "127.0.0.1"),
		RTT:         1500 * time.Microsecond,
		Success:     true,
		TTL:         64,
		PacketSize:  56,
		SourceIP:    net.ParseIP("127.0.0.1"),
		Timestamp:   time.Now(),
		ErrorCode:   types.ICMPErrNoError,
		ErrorMessage: "",
	}

	// Record measurement metrics
	err = meterProvider.RecordMeasurement(measurement)
	if err != nil {
		t.Errorf("Failed to record measurement metrics: %v", err)
	}

	// End measurement span
	tracerProvider.EndMeasurementSpan(measurementSpanCtx, measurementSpan, measurement)

	// Test error handling with tracing
	errorSpanCtx, errorSpan := tracerProvider.StartMeasurementSpan(
		ctx, "test-error-measurement", "test-probe-001", "test-target-002",
	)

	// Create failed measurement
	failedMeasurement := &types.MeasurementData{
		ID:          "test-error-001",
		ProbeID:       "test-probe-001",
		Target:        createTestTarget("test-target-002", "127.0.0.1"),
		RTT:           0,
		Success:       false,
		TTL:           0,
		PacketSize:    0,
		SourceIP:      nil,
		Timestamp:     time.Now(),
		ErrorCode:     types.ICMPErrTimeout,
		ErrorMessage:  "Connection timeout",
	}

	err = meterProvider.RecordMeasurement(failedMeasurement)
	if err != nil {
		t.Errorf("Failed to record failed measurement metrics: %v", err)
	}

	tracerProvider.EndMeasurementSpan(errorSpanCtx, errorSpan, failedMeasurement)

	// Test configuration update tracing
	configSpanCtx, configSpan := tracerProvider.StartConfigurationSpan(ctx, "update")
	tracerProvider.EndConfigurationSpan(configSpanCtx, configSpan, nil)

	// Test mesh operation tracing
	meshSpanCtx, meshSpan := tracerProvider.StartMeshSpan(ctx, "coordination", []string{"probe-1", "probe-2"})
	tracerProvider.EndMeshSpan(meshSpanCtx, meshSpan, nil)

	// Test error recording
	meterProvider.RecordError("network_error", "Connection refused", "test-probe-001")

	// Test probe uptime recording
	meterProvider.RecordProbeUptime(3600) // 1 hour

	t.Logf("OpenTelemetry integration test completed successfully")
}

// TestOpenTelemetryMetricsValidation validates OpenTelemetry metrics format and structure
func TestOpenTelemetryMetricsValidation(t *testing.T) {
	// Initialize with default provider
	otel.SetMeterProvider(nil)
	otel.SetTracerProvider(nil)

	meterProvider := telemetry.NewOTELMeterProvider()

	err := meterProvider.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize meter provider for validation: %v", err)
	}

	// Create test measurements with various scenarios
	testCases := []struct {
		name         string
		measurement  *types.MeasurementData
		expectError  bool
	}{
		{
			name: "successful measurement",
			measurement: &types.MeasurementData{
				ID:          "validation-001",
				ProbeID:     "validation-probe-001",
				Target:      createTestTarget("validation-target-001", "8.8.8.8"),
				RTT:         25000 * time.Microsecond,
				Success:     true,
				TTL:         64,
				PacketSize:  56,
				SourceIP:    net.ParseIP("192.168.1.100"),
				Timestamp:   time.Now(),
				ErrorCode:   types.ICMPErrNoError,
				ErrorMessage: "",
			},
			expectError: false,
		},
		{
			name: "failed measurement",
			measurement: &types.MeasurementData{
				ID:          "validation-002",
				ProbeID:       "validation-probe-001",
				Target:        createTestTarget("validation-target-002", "192.0.2.1"),
				RTT:           0,
				Success:       false,
				TTL:           0,
				PacketSize:    0,
				SourceIP:      nil,
				Timestamp:     time.Now(),
				ErrorCode:     types.ICMPErrUnreachable,
				ErrorMessage:  "Destination unreachable",
			},
			expectError: false,
		},
		{
			name: "zero RTT measurement",
			measurement: &types.MeasurementData{
				ID:          "validation-003",
				ProbeID:     "validation-probe-001",
				Target:      createTestTarget("validation-target-003", "127.0.0.1"),
				RTT:         0,
				Success:     true,
				TTL:         64,
				PacketSize:  56,
				SourceIP:    net.ParseIP("127.0.0.1"),
				Timestamp:   time.Now(),
				ErrorCode:   types.ICMPErrNoError,
				ErrorMessage: "",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := meterProvider.RecordMeasurement(tc.measurement)
			
			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s but got none", tc.name)
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.name, err)
			} else {
				t.Logf("Metrics validation passed for %s", tc.name)
			}
		})
	}

	t.Logf("OpenTelemetry metrics validation completed")
}

// TestOpenTelemetryTracingValidation validates OpenTelemetry tracing functionality
func TestOpenTelemetryTracingValidation(t *testing.T) {
	ctx := context.Background()

	// Initialize with default provider
	otel.SetMeterProvider(nil)
	otel.SetTracerProvider(nil)

	tracerProvider := telemetry.NewOTELTracerProvider()

	// Test measurement span lifecycle
	t.Run("measurement span lifecycle", func(t *testing.T) {
		measurementID := "trace-test-001"
		probeID := "trace-test-probe"
		targetID := "trace-test-target"

		// Start span
		spanCtx, span := tracerProvider.StartMeasurementSpan(ctx, measurementID, probeID, targetID)
		if spanCtx == nil || span == nil {
			t.Fatal("Failed to start measurement span")
		}

		// Create test measurement
		measurement := &types.MeasurementData{
			ID:          measurementID,
			ProbeID:     probeID,
			Target:      createTestTarget(targetID, "192.0.2.1"),
			RTT:         5000 * time.Microsecond,
			Success:     false,
			TTL:         64,
			PacketSize:  56,
			SourceIP:    net.ParseIP("192.168.1.1"),
			Timestamp:   time.Now(),
			ErrorCode:   types.ICMPErrNoRoute,
			ErrorMessage: "Network unreachable",
		}

		// End span with measurement result
		tracerProvider.EndMeasurementSpan(spanCtx, span, measurement)

		t.Logf("Measurement span lifecycle test completed")
	})

	// Test configuration span lifecycle
	t.Run("configuration span lifecycle", func(t *testing.T) {
		operation := "update_config"

		spanCtx, span := tracerProvider.StartConfigurationSpan(ctx, operation)
		if spanCtx == nil || span == nil {
			t.Fatal("Failed to start configuration span")
		}

		// End span with success
		tracerProvider.EndConfigurationSpan(spanCtx, span, nil)

		// Test with error
		testError := context.Canceled
		spanCtx2, span2 := tracerProvider.StartConfigurationSpan(ctx, operation)
		tracerProvider.EndConfigurationSpan(spanCtx2, span2, testError)

		t.Logf("Configuration span lifecycle test completed")
	})

	// Test mesh span lifecycle
	t.Run("mesh span lifecycle", func(t *testing.T) {
		operation := "synchronize"
		participants := []string{"probe-1", "probe-2", "probe-3"}

		spanCtx, span := tracerProvider.StartMeshSpan(ctx, operation, participants)
		if spanCtx == nil || span == nil {
			t.Fatal("Failed to start mesh span")
		}

		// End span with success
		tracerProvider.EndMeshSpan(spanCtx, span, nil)

		t.Logf("Mesh span lifecycle test completed")
	})

	t.Logf("OpenTelemetry tracing validation completed")
}

// TestOpenTelemetryEndToEndFlow tests the complete telemetry flow
func TestOpenTelemetryEndToEndFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end telemetry flow test in short mode")
	}

	ctx := context.Background()

	// Initialize with default provider
	otel.SetMeterProvider(nil)
	otel.SetTracerProvider(nil)

	// Create telemetry components
	meterProvider := telemetry.NewOTELMeterProvider()
	tracerProvider := telemetry.NewOTELTracerProvider()

	// Initialize meter provider
	err := meterProvider.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize meter provider: %v", err)
	}

	// Simulate a complete measurement workflow with telemetry
	target := createTestTarget("e2e-target", "8.8.8.8")
	probeID := "e2e-probe-001"

	// Start measurement operation span
	measurementSpanCtx, measurementSpan := tracerProvider.StartMeasurementSpan(
		ctx, "e2e-measurement-001", probeID, target.ID,
	)

	// Simulate probe processing time
	time.Sleep(10 * time.Millisecond)

	// Create successful measurement
	measurement := &types.MeasurementData{
		ID:          "e2e-001",
		ProbeID:     probeID,
		Target:      target,
		RTT:         15000 * time.Microsecond,
		Success:     true,
		TTL:         64,
		PacketSize:  56,
		SourceIP:    net.ParseIP("192.168.1.100"),
		Timestamp:   time.Now(),
		ErrorCode:   types.ICMPErrNoError,
		ErrorMessage: "",
	}

	// Record metrics during measurement
	err = meterProvider.RecordMeasurement(measurement)
	if err != nil {
		t.Errorf("Failed to record measurement metrics: %v", err)
	}

	// End measurement span
	tracerProvider.EndMeasurementSpan(measurementSpanCtx, measurementSpan, measurement)

	// Start configuration update span (simulating parallel operation)
	configSpanCtx, configSpan := tracerProvider.StartConfigurationSpan(ctx, "update_telemetry")
	
	// Simulate configuration processing
	time.Sleep(5 * time.Millisecond)
	
	// Record configuration update metrics
	err = meterProvider.RecordConfigurationUpdate("telemetry-config-001")
	if err != nil {
		t.Errorf("Failed to record configuration metrics: %v", err)
	}

	tracerProvider.EndConfigurationSpan(configSpanCtx, configSpan, nil)

	// Record overall probe health metrics
	meterProvider.RecordError("system", "normal operation", probeID)
	meterProvider.RecordProbeUptime(3600) // 1 hour uptime

	// Shutdown telemetry
	err = meterProvider.Shutdown(ctx)
	if err != nil {
		t.Errorf("Failed to shutdown meter provider: %v", err)
	}

	t.Logf("OpenTelemetry end-to-end flow test completed successfully")
}

// BenchmarkOpenTelemetryPerformance benchmarks OpenTelemetry performance
func BenchmarkOpenTelemetryMeasurementRecording(b *testing.B) {
	// Initialize with default provider
	otel.SetMeterProvider(nil)
	otel.SetTracerProvider(nil)

	meterProvider := telemetry.NewOTELMeterProvider()

	err := meterProvider.Initialize(nil)
	if err != nil {
		b.Fatalf("Failed to initialize meter provider: %v", err)
	}

	target := createTestTarget("benchmark-target", "127.0.0.1")
	measurement := &types.MeasurementData{
		ID:          "benchmark-001",
		ProbeID:     "benchmark-probe",
		Target:      target,
		RTT:         1000 * time.Microsecond,
		Success:     true,
		TTL:         64,
		PacketSize:  56,
		SourceIP:    net.ParseIP("127.0.0.1"),
		Timestamp:   time.Now(),
		ErrorCode:   types.ICMPErrNoError,
		ErrorMessage: "",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := meterProvider.RecordMeasurement(measurement)
		if err != nil {
			b.Fatalf("Failed to record measurement: %v", err)
		}
	}
}

// BenchmarkOpenTelemetrySpanCreation benchmarks span creation performance
func BenchmarkOpenTelemetrySpanCreation(b *testing.B) {
	ctx := context.Background()

	// Initialize with default provider
	otel.SetMeterProvider(nil)
	otel.SetTracerProvider(nil)

	tracerProvider := telemetry.NewOTELTracerProvider()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		spanCtx, span := tracerProvider.StartMeasurementSpan(ctx, "benchmark-measurement", "benchmark-probe", "benchmark-target")
		
		measurement := &types.MeasurementData{
			ID:          "benchmark-span-001",
			ProbeID:     "benchmark-probe",
			Target:      createTestTarget("benchmark-target", "127.0.0.1"),
			RTT:         1000 * time.Microsecond,
			Success:     true,
			TTL:         64,
			PacketSize:  56,
			SourceIP:    net.ParseIP("127.0.0.1"),
			Timestamp:   time.Now(),
			ErrorCode:   types.ICMPErrNoError,
			ErrorMessage: "",
		}
		
		tracerProvider.EndMeasurementSpan(spanCtx, span, measurement)
	}
}

// Helper functions

func createTestTarget(id, address string) types.NetworkTarget {
	return types.NetworkTarget{
		ID:      id,
		Address: net.ParseIP(address),
		Timeout: 5 * time.Second,
		Enabled: true,
	}
}