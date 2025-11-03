package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/mesh-net-probe/probe/pkg/types"
)

// Manager handles OpenTelemetry initialization and telemetry export
type Manager struct {
	config *types.TelemetryConfig
	tracer trace.Tracer
	meter  metric.Meter
}

// NewManager creates a new OpenTelemetry manager
func NewManager(config *types.TelemetryConfig) (*Manager, error) {
	if config == nil {
		config = &types.TelemetryConfig{
			EnableMetrics:   true,
			EnableTracing:   true,
			EnableLogging:   true,
			LogLevel:        "info",
			ExportInterval:  30 * time.Second,
			BufferSize:      1000,
		}
	}

	m := &Manager{
		config: config,
	}

	// Initialize tracer and meter
	m.tracer = otel.Tracer("mesh-probe")
	m.meter = otel.Meter("mesh-probe")

	return m, nil
}

// Start starts telemetry collection (placeholder for future implementation)
func (m *Manager) Start(ctx context.Context) error {
	// For now, this is a placeholder
	// Future implementation would set up exporters and processors
	return nil
}

// Stop stops telemetry collection (placeholder for future implementation)
func (m *Manager) Stop(ctx context.Context) error {
	// For now, this is a placeholder
	// Future implementation would flush data and shutdown
	return nil
}

// RecordMeasurement records ICMP measurement telemetry
func (m *Manager) RecordMeasurement(ctx context.Context, measurement *types.MeasurementData) error {
	if !m.config.EnableMetrics && !m.config.EnableTracing {
		return nil
	}

	// Create trace span for measurement
	ctx, span := m.tracer.Start(ctx, "icmp.measurement")
	defer span.End()

	// Record measurement attributes
	span.SetAttributes(
		attribute.String("probe.id", measurement.ProbeID),
		attribute.String("target.address", measurement.Target.Address.String()),
		attribute.Bool("measurement.success", measurement.Success),
		attribute.Int64("measurement.rtt_us", measurement.RTT.Microseconds()),
		attribute.Int("measurement.ttl", measurement.TTL),
		attribute.Int("measurement.packet_size", measurement.PacketSize),
		attribute.String("measurement.platform.os", measurement.PlatformData.OS),
		attribute.String("measurement.platform.arch", measurement.PlatformData.Architecture),
	)

	// Set error information if measurement failed
	if !measurement.Success {
		span.SetAttributes(
			attribute.String("measurement.error_message", measurement.ErrorMessage),
			attribute.String("measurement.error_code", string(measurement.ErrorCode)),
		)
	}

	// Record custom metrics (simplified implementation)
	if m.meter != nil && m.config.EnableMetrics {
		m.recordMeasurementMetrics(ctx, measurement)
	}

	return nil
}

// recordMeasurementMetrics records measurement-specific metrics
func (m *Manager) recordMeasurementMetrics(ctx context.Context, measurement *types.MeasurementData) {
	// Create counters for measurement success/failure
	measurementCounter, err := m.meter.Int64Counter(
		"icmp_measurements_total",
		metric.WithDescription("Total ICMP measurements performed"),
	)
	if err != nil {
		// Handle error silently for now
		return
	}

	measurementCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("target", measurement.Target.Address.String()),
			attribute.String("probe_id", measurement.ProbeID),
			attribute.Bool("success", measurement.Success),
		),
	)

	// Record RTT as a histogram if measurement was successful
	if measurement.Success && measurement.RTT > 0 {
		rttHistogram, err := m.meter.Int64Histogram(
			"icmp_rtt_microseconds",
			metric.WithDescription("ICMP round-trip time in microseconds"),
		)
		if err != nil {
			return
		}

		rttHistogram.Record(ctx, measurement.RTT.Microseconds(),
			metric.WithAttributes(
				attribute.String("target", measurement.Target.Address.String()),
				attribute.String("probe_id", measurement.ProbeID),
			),
		)
	}

	// Record packet loss
	if measurement.ErrorCode == types.ICMPErrTimeout {
		timeoutCounter, err := m.meter.Int64Counter(
			"icmp_timeouts_total",
			metric.WithDescription("Total ICMP timeouts"),
		)
		if err != nil {
			return
		}

		timeoutCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("target", measurement.Target.Address.String()),
				attribute.String("probe_id", measurement.ProbeID),
			),
		)
	}
}

// RecordProbeHealth records probe health telemetry
func (m *Manager) RecordProbeHealth(ctx context.Context, probe *types.ProbeInstance) error {
	if !m.config.EnableMetrics && !m.config.EnableTracing {
		return nil
	}

	// Create span for health check
	ctx, span := m.tracer.Start(ctx, "probe.health.check")
	defer span.End()

	// Record probe health attributes
	span.SetAttributes(
		attribute.String("probe.id", probe.ID),
		attribute.String("probe.state", string(probe.Status.State)),
		attribute.Float64("probe.health_score", probe.Status.HealthScore),
		attribute.Int("probe.active_targets", probe.Status.ActiveTargets),
		attribute.String("probe.os", probe.Platform.OS),
		attribute.String("probe.arch", probe.Platform.Arch),
		attribute.Bool("probe.container", probe.Platform.Container),
	)

	// Record probe metrics
	if m.meter != nil && m.config.EnableMetrics {
		// Measurements total
		measurementsCounter, err := m.meter.Int64Counter(
			"probe_measurements_total",
			metric.WithDescription("Total measurements performed by probe"),
		)
		if err == nil {
			measurementsCounter.Add(ctx, int64(probe.Metrics.MeasurementsTotal),
				metric.WithAttributes(
					attribute.String("probe_id", probe.ID),
					attribute.String("probe_state", string(probe.Status.State)),
				),
			)
		}

		// Success rate (simplified - just record as trace attribute)
		if probe.Metrics.MeasurementsTotal > 0 {
			successRate := float64(probe.Metrics.MeasurementsSuccess) / float64(probe.Metrics.MeasurementsTotal)
			span.SetAttributes(
				attribute.Float64("probe.success_rate", successRate),
			)
		}

		// CPU and memory usage (simplified - record as trace attributes)
		span.SetAttributes(
			attribute.Float64("probe.cpu_usage", probe.Metrics.CPUUsage),
			attribute.Int64("probe.memory_usage", int64(probe.Metrics.MemoryUsage)),
		)
	}

	return nil
}

// RecordConfigChange records configuration change telemetry
func (m *Manager) RecordConfigChange(ctx context.Context, config *types.Configuration, source string) error {
	if !m.config.EnableMetrics && !m.config.EnableTracing {
		return nil
	}

	// Create span for config change
	ctx, span := m.tracer.Start(ctx, "config.change")
	defer span.End()

	// Record config change attributes
	span.SetAttributes(
		attribute.String("config.id", config.ID),
		attribute.String("config.name", config.Name),
		attribute.Int("config.version", config.Version),
		attribute.String("config.source", source),
		attribute.Int("config.targets_count", len(config.Targets)),
	)

	if config.Network != nil {
		span.SetAttributes(
			attribute.String("config.network", config.Network.SourceIP.String()),
		)
	}

	// Record config metrics
	if m.meter != nil && m.config.EnableMetrics {
		configCounter, err := m.meter.Int64Counter(
			"config_changes_total",
			metric.WithDescription("Total configuration changes"),
		)
		if err == nil {
			configCounter.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("source", source),
					attribute.String("config_id", config.ID),
				),
			)
		}

		// Target count (record as trace attribute)
		span.SetAttributes(
			attribute.Int("config.targets_count", len(config.Targets)),
		)
	}

	return nil
}

// RecordMeshEvent records mesh networking events
func (m *Manager) RecordMeshEvent(ctx context.Context, eventType string, probeID string, details map[string]interface{}) error {
	if !m.config.EnableMetrics && !m.config.EnableTracing {
		return nil
	}

	// Create span for mesh event
	ctx, span := m.tracer.Start(ctx, "mesh.event")
	defer span.End()

	// Record mesh event attributes
	attrs := []attribute.KeyValue{
		attribute.String("mesh.event_type", eventType),
		attribute.String("probe.id", probeID),
	}

	// Add detail attributes
	for key, value := range details {
		switch v := value.(type) {
		case string:
			attrs = append(attrs, attribute.String(fmt.Sprintf("mesh.%s", key), v))
		case int, int32, int64:
			attrs = append(attrs, attribute.Int(fmt.Sprintf("mesh.%s", key), value.(int)))
		case float64:
			attrs = append(attrs, attribute.Float64(fmt.Sprintf("mesh.%s", key), v))
		case bool:
			attrs = append(attrs, attribute.Bool(fmt.Sprintf("mesh.%s", key), v))
		}
	}

	span.SetAttributes(attrs...)

	// Record mesh metrics
	if m.meter != nil && m.config.EnableMetrics {
		meshCounter, err := m.meter.Int64Counter(
			"mesh_events_total",
			metric.WithDescription("Total mesh networking events"),
		)
		if err == nil {
			meshCounter.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("event_type", eventType),
					attribute.String("probe_id", probeID),
				),
			)
		}
	}

	return nil
}

// GetTracer returns the OpenTelemetry tracer
func (m *Manager) GetTracer() trace.Tracer {
	return m.tracer
}

// GetMeter returns the OpenTelemetry meter
func (m *Manager) GetMeter() metric.Meter {
	return m.meter
}

// IsEnabled returns whether telemetry is enabled
func (m *Manager) IsEnabled() bool {
	return m.config.EnableMetrics || m.config.EnableTracing || m.config.EnableLogging
}