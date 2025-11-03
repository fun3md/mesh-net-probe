package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"

	"github.com/mesh-net-probe/probe/pkg/types"
)

// MeterProvider defines the interface for OpenTelemetry meter provider
type MeterProvider interface {
	Initialize(config *types.TelemetryConfig) error
	RecordMeasurement(measurement *types.MeasurementData) error
	RecordProbeMetrics(metrics *types.ProbeMetrics) error
	RecordConfigurationUpdate(configID string) error
	RecordError(errorType string, errorMsg string, probeID string)
	RecordProbeUptime(uptimeSeconds int64)
	Shutdown(ctx context.Context) error
}

// OTELMeterProvider implements OpenTelemetry metrics and tracing
type OTELMeterProvider struct {
	meter             metric.Meter
	tracer            trace.Tracer
	measurementCount  metric.Int64Counter
	measurementErrors metric.Int64Counter
	probeUptime       metric.Int64ObservableGauge
	rttHistogram      metric.Float64Histogram
}

// NewOTELMeterProvider creates a new OpenTelemetry meter provider
func NewOTELMeterProvider() *OTELMeterProvider {
	meter := otel.Meter("mesh-probe")
	tracer := otel.Tracer("mesh-probe")

	return &OTELMeterProvider{
		meter:  meter,
		tracer: tracer,
	}
}

// Initialize sets up the OpenTelemetry meter provider with the given configuration
func (p *OTELMeterProvider) Initialize(config *types.TelemetryConfig) error {
	// Create counters for measurement metrics
	measurementCount, err := p.meter.Int64Counter(
		"probe_measurements_total",
		metric.WithDescription("Total number of ICMP measurements performed"),
	)
	if err != nil {
		return fmt.Errorf("failed to create measurement counter: %w", err)
	}
	p.measurementCount = measurementCount

	// Create counter for measurement errors
	measurementErrors, err := p.meter.Int64Counter(
		"probe_measurement_errors_total",
		metric.WithDescription("Total number of failed ICMP measurements"),
	)
	if err != nil {
		return fmt.Errorf("failed to create error counter: %w", err)
	}
	p.measurementErrors = measurementErrors

	// Create gauge for probe uptime
	probeUptime, err := p.meter.Int64ObservableGauge(
		"probe_uptime_seconds",
		metric.WithDescription("Probe uptime in seconds"),
	)
	if err != nil {
		return fmt.Errorf("failed to create uptime gauge: %w", err)
	}
	p.probeUptime = probeUptime

	// Create histogram for round-trip time measurements
	rttHistogram, err := p.meter.Float64Histogram(
		"probe_rtt_microseconds",
		metric.WithDescription("ICMP round-trip time in microseconds"),
	)
	if err != nil {
		return fmt.Errorf("failed to create RTT histogram: %w", err)
	}
	p.rttHistogram = rttHistogram

	return nil
}

// RecordMeasurement records measurement data to OpenTelemetry
func (p *OTELMeterProvider) RecordMeasurement(measurement *types.MeasurementData) error {
	ctx := context.Background()

	// Record basic measurement count
	p.measurementCount.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("probe_id", measurement.ProbeID),
			attribute.String("target_id", measurement.Target.ID),
			attribute.String("target_address", measurement.Target.Address.String()),
			attribute.Bool("success", measurement.Success),
		),
	)

	// Record error if measurement failed
	if !measurement.Success {
		p.measurementErrors.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("probe_id", measurement.ProbeID),
				attribute.String("target_id", measurement.Target.ID),
				attribute.String("error_code", fmt.Sprintf("%d", measurement.ErrorCode)),
				attribute.String("error_message", measurement.ErrorMessage),
			),
		)
		return nil
	}

	// Record successful measurement metrics
	rttMicroseconds := float64(measurement.RTT.Nanoseconds()) / 1000.0
	p.rttHistogram.Record(ctx, rttMicroseconds,
		metric.WithAttributes(
			attribute.String("probe_id", measurement.ProbeID),
			attribute.String("target_id", measurement.Target.ID),
			attribute.String("target_address", measurement.Target.Address.String()),
			attribute.Int("ttl", measurement.TTL),
			attribute.Int("packet_size", measurement.PacketSize),
			attribute.String("source_ip", measurement.SourceIP.String()),
		),
	)

	return nil
}

// RecordProbeMetrics records probe runtime metrics
func (p *OTELMeterProvider) RecordProbeMetrics(metrics *types.ProbeMetrics) error {
	// Observable gauges are handled via callbacks during collection
	// This method is a placeholder for when we implement callback-based metrics
	return nil
}

// RecordConfigurationUpdate records configuration management events
func (p *OTELMeterProvider) RecordConfigurationUpdate(configID string) error {
	// This would typically be recorded as a trace span or event
	// Implementation depends on the specific OpenTelemetry setup
	return nil
}

// RecordProbeUptime records current probe uptime in seconds
func (p *OTELMeterProvider) RecordProbeUptime(uptimeSeconds int64) {
	// This would typically record the uptime value
	// For observable gauges, this would be handled via callbacks
}

// RecordError records error events for monitoring and debugging
func (p *OTELMeterProvider) RecordError(errorType string, errorMsg string, probeID string) {
	ctx := context.Background()
	
	p.measurementErrors.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("probe_id", probeID),
			attribute.String("error_type", errorType),
			attribute.String("error_message", errorMsg),
		),
	)
}

// Shutdown gracefully shuts down the meter provider
func (p *OTELMeterProvider) Shutdown(ctx context.Context) error {
	// OpenTelemetry SDK cleanup would happen here
	// This is a simplified implementation
	return nil
}

// TracerProvider provides distributed tracing functionality
type TracerProvider interface {
	StartMeasurementSpan(ctx context.Context, measurementID, probeID, targetID string) (context.Context, trace.Span)
	EndMeasurementSpan(ctx context.Context, span trace.Span, measurement *types.MeasurementData)
	StartConfigurationSpan(ctx context.Context, operation string) (context.Context, trace.Span)
	EndConfigurationSpan(ctx context.Context, span trace.Span, err error)
	StartMeshSpan(ctx context.Context, operation string, participants []string) (context.Context, trace.Span)
	EndMeshSpan(ctx context.Context, span trace.Span, err error)
}

// OTELTracerProvider implements OpenTelemetry tracing
type OTELTracerProvider struct {
	tracer trace.Tracer
}

// NewOTELTracerProvider creates a new OpenTelemetry tracer provider
func NewOTELTracerProvider() *OTELTracerProvider {
	return &OTELTracerProvider{
		tracer: otel.Tracer("mesh-probe"),
	}
}

// StartMeasurementSpan starts a trace span for an ICMP measurement
func (t *OTELTracerProvider) StartMeasurementSpan(ctx context.Context, measurementID, probeID, targetID string) (context.Context, trace.Span) {
	spanCtx, span := t.tracer.Start(ctx, "icmp.measurement",
		trace.WithAttributes(
			attribute.String("measurement.id", measurementID),
			attribute.String("probe.id", probeID),
			attribute.String("target.id", targetID),
		),
	)
	return spanCtx, span
}

// EndMeasurementSpan ends a measurement trace span with the result
func (t *OTELTracerProvider) EndMeasurementSpan(ctx context.Context, span trace.Span, measurement *types.MeasurementData) {
	if measurement == nil {
		span.SetStatus(codes.Error, "measurement data nil")
		span.End()
		return
	}

	// Set span attributes based on measurement result
	span.SetAttributes(
		attribute.Bool("measurement.success", measurement.Success),
		attribute.Int64("measurement.rtt_nanos", measurement.RTT.Nanoseconds()),
		attribute.Int("measurement.ttl", measurement.TTL),
	)

	if !measurement.Success {
		span.SetStatus(codes.Error, measurement.ErrorMessage)
		span.SetAttributes(
			attribute.String("measurement.error_code", fmt.Sprintf("%d", measurement.ErrorCode)),
		)
	} else {
		span.SetStatus(codes.Ok, "")
	}

	span.End()
}

// StartConfigurationSpan starts a trace span for configuration management operations
func (t *OTELTracerProvider) StartConfigurationSpan(ctx context.Context, operation string) (context.Context, trace.Span) {
	spanCtx, span := t.tracer.Start(ctx, "config."+operation,
		trace.WithAttributes(
			attribute.String("config.operation", operation),
		),
	)
	return spanCtx, span
}

// EndConfigurationSpan ends a configuration trace span
func (t *OTELTracerProvider) EndConfigurationSpan(ctx context.Context, span trace.Span, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}

// StartMeshSpan starts a trace span for mesh coordination operations
func (t *OTELTracerProvider) StartMeshSpan(ctx context.Context, operation string, participants []string) (context.Context, trace.Span) {
	spanCtx, span := t.tracer.Start(ctx, "mesh."+operation,
		trace.WithAttributes(
			attribute.String("mesh.operation", operation),
			attribute.Int("mesh.participants", len(participants)),
		),
	)
	return spanCtx, span
}

// EndMeshSpan ends a mesh trace span
func (t *OTELTracerProvider) EndMeshSpan(ctx context.Context, span trace.Span, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}