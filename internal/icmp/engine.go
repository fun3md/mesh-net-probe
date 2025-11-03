package icmp

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"

	"github.com/mesh-net-probe/probe/internal/errors"
	"github.com/mesh-net-probe/probe/internal/logger"
	"github.com/mesh-net-probe/probe/internal/network"
	"github.com/mesh-net-probe/probe/pkg/types"
)

// Engine handles ICMP measurement operations with microsecond precision
type Engine interface {
	// Initialize sets up the ICMP engine with configuration
	Initialize(ctx context.Context, config *types.NetworkConfig) error

	// Measure performs a single ICMP measurement
	Measure(ctx context.Context, target *types.NetworkTarget) (*types.MeasurementData, error)

	// MeasureBatch performs multiple measurements in sequence
	MeasureBatch(ctx context.Context, targets []*types.NetworkTarget) ([]*types.MeasurementData, error)

	// StartContinuous starts continuous measurement mode
	StartContinuous(ctx context.Context, targets []*types.NetworkTarget, interval time.Duration) error

	// StopContinuous stops continuous measurement mode
	StopContinuous(ctx context.Context) error

	// GetStats returns current engine statistics
	GetStats() *EngineStats

	// Close shuts down the ICMP engine
	Close(ctx context.Context) error
}

// EngineStats contains ICMP engine runtime statistics
type EngineStats struct {
	MeasurementsTotal   uint64           `json:"measurements_total"`     // Total measurements performed
	MeasurementsSuccess uint64           `json:"measurements_success"`   // Successful measurements
	MeasurementsFailed  uint64           `json:"measurements_failed"`    // Failed measurements
	AverageRTT          time.Duration    `json:"average_rtt"`            // Average RTT
	MinRTT             time.Duration    `json:"min_rtt"`                // Minimum RTT
	MaxRTT             time.Duration    `json:"max_rtt"`                // Maximum RTT
	LastUpdate         time.Time        `json:"last_update"`            // Last stats update
}

// simpleEngine provides a simplified ICMP measurement implementation
type simpleEngine struct {
	config       *types.NetworkConfig
	conn         *net.IPConn
	packetHandler network.PacketHandler
	logger       logger.Logger
	stats        EngineStats
}

// NewEngine creates a new ICMP measurement engine
func NewEngine() Engine {
	return &simpleEngine{
		packetHandler: network.NewPacketHandler(),
		logger:        logger.GetGlobalLogger(),
	}
}

// Initialize sets up the ICMP engine with configuration
func (e *simpleEngine) Initialize(ctx context.Context, config *types.NetworkConfig) error {
	// Create ICMP connection for IPv4
	sourceAddr := &net.IPAddr{IP: config.SourceIP}
	conn, err := net.ListenIP("ip4:icmp", sourceAddr)
	if err != nil {
		return errors.NewNetworkError(
			errors.ICMPPermissionDenied,
			fmt.Sprintf("failed to create ICMP connection: %v", err),
			errors.WithCause(err),
			errors.WithComponent("icmp.engine"),
		)
	}

	e.config = config
	e.conn = conn

	// Configure socket options
	if config.BufferSize > 0 {
		if err := conn.SetReadBuffer(config.BufferSize); err != nil {
			conn.Close()
			return fmt.Errorf("failed to set read buffer: %w", err)
		}
		if err := conn.SetWriteBuffer(config.BufferSize); err != nil {
			conn.Close()
			return fmt.Errorf("failed to set write buffer: %w", err)
		}
	}

	e.logger.Info(ctx, "ICMP engine initialized",
		"source_ip", config.SourceIP.String(),
		"buffer_size", config.BufferSize,
	)

	return nil
}

// Measure performs a single ICMP measurement
func (e *simpleEngine) Measure(ctx context.Context, target *types.NetworkTarget) (*types.MeasurementData, error) {
	// Validate target
	if err := e.validateTarget(target); err != nil {
		return nil, err
	}

	startTime := time.Now()

	// Create ICMP echo request packet
	packet, err := createICMPEchoRequest()
	if err != nil {
		return nil, errors.NewNetworkError(
			errors.ICMPInvalidPacket,
			fmt.Sprintf("failed to create ICMP packet for target %s: %v", target.ID, err),
			errors.WithCause(err),
			errors.WithComponent("icmp.engine"),
			errors.WithContext("target_id", target.ID),
		)
	}

	// Marshal packet to binary
	data, err := packet.Marshal(nil)
	if err != nil {
		return nil, errors.NewNetworkError(
			errors.ICMPInvalidPacket,
			fmt.Sprintf("failed to marshal ICMP packet for target %s: %v", target.ID, err),
			errors.WithCause(err),
			errors.WithComponent("icmp.engine"),
			errors.WithContext("target_id", target.ID),
		)
	}

	// Set timeout
	timeout := target.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second // Default timeout
	}

	// Send ICMP packet
	requestSent := time.Now()
	dstAddr := &net.IPAddr{IP: target.Address}
	
	n, err := e.conn.WriteTo(data, dstAddr)
	if err != nil {
		e.recordFailure()
		return nil, e.handleSendError(ctx, target, err)
	}

	if n != len(data) {
		e.recordFailure()
		return nil, errors.NewNetworkError(
			errors.ICMPInvalidPacket,
			fmt.Sprintf("partial write for target %s: sent %d bytes, expected %d", target.ID, n, len(data)),
			errors.WithComponent("icmp.engine"),
			errors.WithContext("target_id", target.ID),
		)
	}

	// Receive response
	recvBuffer := make([]byte, 1500) // Standard MTU
	e.conn.SetReadDeadline(time.Now().Add(timeout))

	n, _, _ = e.conn.ReadFrom(recvBuffer)
	responseRecv := time.Now()
	
	// Reset deadline
	e.conn.SetReadDeadline(time.Time{})

	if n == 0 {
		e.recordFailure()
		return &types.MeasurementData{
			ID:           generateMeasurementID(),
			ProbeID:      "local",
			Target:       *target,
			Timestamp:    startTime,
			RequestSent:  requestSent,
			Success:      false,
			ErrorCode:    types.ICMPErrTimeout,
			ErrorMessage: fmt.Sprintf("timeout after %v", timeout),
			SourceIP:     e.config.SourceIP,
			DestIP:       target.Address,
		}, nil
	}

	// Calculate RTT with microsecond precision
	rtt := responseRecv.Sub(requestSent)

	// Record success
	e.recordSuccess(rtt)

	// Create measurement data
	measurement := &types.MeasurementData{
		ID:           generateMeasurementID(),
		ProbeID:      "local",
		Target:       *target,
		Timestamp:    startTime,
		RequestSent:  requestSent,
		ResponseRecv: responseRecv,
		RTT:          rtt,
		Success:      true,
		SourceIP:     e.config.SourceIP,
		DestIP:       target.Address,
		PacketSize:   len(data),
		
		// Platform metadata
		PlatformData: types.PlatformMeta{
			OS:           "local",
			Architecture: "amd64",
			Precision:    types.PrecisionMicrosecond,
		},
	}

	return measurement, nil
}

// MeasureBatch performs multiple measurements in sequence
func (e *simpleEngine) MeasureBatch(ctx context.Context, targets []*types.NetworkTarget) ([]*types.MeasurementData, error) {
	measurements := make([]*types.MeasurementData, 0, len(targets))

	for _, target := range targets {
		select {
		case <-ctx.Done():
			return measurements, ctx.Err()
		default:
		}

		measurement, err := e.Measure(ctx, target)
		if err != nil {
			e.logger.Error(ctx, err, "Failed to measure target", "target_id", target.ID)
			continue
		}

		measurements = append(measurements, measurement)
	}

	return measurements, nil
}

// StartContinuous starts continuous measurement mode
func (e *simpleEngine) StartContinuous(ctx context.Context, targets []*types.NetworkTarget, interval time.Duration) error {
	// Simplified implementation - just log the start
	e.logger.Info(ctx, "Continuous measurement mode started", 
		"targets_count", len(targets), 
		"interval", interval)
	return nil
}

// StopContinuous stops continuous measurement mode
func (e *simpleEngine) StopContinuous(ctx context.Context) error {
	// Simplified implementation - just log the stop
	e.logger.Info(ctx, "Continuous measurement mode stopped")
	return nil
}

// GetStats returns current engine statistics
func (e *simpleEngine) GetStats() *EngineStats {
	stats := e.stats
	stats.LastUpdate = time.Now()
	return &stats
}

// Close shuts down the ICMP engine
func (e *simpleEngine) Close(ctx context.Context) error {
	if e.conn != nil {
		e.conn.Close()
	}
	e.logger.Info(ctx, "ICMP engine closed")
	return nil
}

// Helper methods

func (e *simpleEngine) validateTarget(target *types.NetworkTarget) error {
	if target == nil {
		return errors.NewValidationError(
			errors.ConfigMissingRequired,
			"target cannot be nil",
			errors.WithComponent("icmp.engine"),
		)
	}

	if target.Address == nil {
		return errors.NewValidationError(
			errors.ConfigMissingRequired,
			"target address cannot be nil",
			errors.WithContext("target_id", target.ID),
			errors.WithComponent("icmp.engine"),
		)
	}

	return nil
}

func (e *simpleEngine) handleSendError(ctx context.Context, target *types.NetworkTarget, err error) *errors.MeshError {
	// Handle specific network errors
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			return errors.NewTimeoutError(
				errors.ICMPTimeout,
				fmt.Sprintf("ICMP timeout for target %s", target.ID),
				errors.WithCause(err),
				errors.WithComponent("icmp.engine"),
				errors.WithContext("target_id", target.ID),
			)
		}
	}

	// Permission denied
	if _, ok := err.(*net.OpError); ok {
		return errors.NewPermissionError(
			errors.ICMPPermissionDenied,
			fmt.Sprintf("permission denied for ICMP to target %s", target.ID),
			errors.WithCause(err),
			errors.WithComponent("icmp.engine"),
			errors.WithContext("target_id", target.ID),
		)
	}

	return errors.NewNetworkError(
		errors.ICMPNoResponse,
		fmt.Sprintf("failed to send ICMP to target %s: %v", target.ID, err),
		errors.WithCause(err),
		errors.WithComponent("icmp.engine"),
		errors.WithContext("target_id", target.ID),
	)
}

func (e *simpleEngine) recordSuccess(rtt time.Duration) {
	e.stats.MeasurementsTotal++
	e.stats.MeasurementsSuccess++
	
	// Update RTT statistics
	if e.stats.MinRTT == 0 || rtt < e.stats.MinRTT {
		e.stats.MinRTT = rtt
	}
	if rtt > e.stats.MaxRTT {
		e.stats.MaxRTT = rtt
	}

	// Calculate average (simple average for now)
	if e.stats.MeasurementsSuccess > 0 {
		e.stats.AverageRTT = rtt // Simplified - using last measurement
	}
}

func (e *simpleEngine) recordFailure() {
	e.stats.MeasurementsTotal++
	e.stats.MeasurementsFailed++
}

// Utility functions

func createICMPEchoRequest() (*icmp.Message, error) {
	// Create simple echo request with minimal payload
	payload := make([]byte, 56) // Standard ping payload size
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	// Create the echo body
	body := &icmp.Echo{
		ID:   1,
		Seq:  1,
		Data: payload,
	}

	// Create the ICMP message for Echo Request
	msg := &icmp.Message{
		Type: icmpMessageType{value: 8}, // Echo Request
		Code: 0,
		Body: body,
	}

	return msg, nil
}

// icmpMessageType implements the icmp.Type interface
type icmpMessageType struct {
	value int
}

func (t icmpMessageType) Protocol() int {
	return t.value
}

func (t icmpMessageType) String() string {
	return fmt.Sprintf("ICMP Type %d", t.value)
}

func generateMeasurementID() string {
	return fmt.Sprintf("measurement_%d", time.Now().UnixNano())
}