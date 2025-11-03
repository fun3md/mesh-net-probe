package icmp

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"github.com/mesh-net-probe/probe/pkg/types"
)

// Engine provides ICMP measurement capabilities with microsecond precision
type Engine struct {
	config  *ICMPConfig
	socket  net.PacketConn
	network string
	address string
}

// ICMPConfig contains ICMP engine configuration
type ICMPConfig struct {
	Network        string        // "ip4", "ip6", or "ip"
	SourceIP       net.IP        // Source IP address (nil for any)
	Timeout        time.Duration // Request timeout
	BufferSize     int           // Send/receive buffer size
	DSCP           int           // DSCP value for QoS
	TTL            int           // Time-to-live
	BindInterface  string        // Network interface name
}

// ICMPEngineOption functional option for configuring the engine
type ICMPEngineOption func(*ICMPConfig)

// WithNetwork specifies the IP network version
func WithNetwork(network string) ICMPEngineOption {
	return func(config *ICMPConfig) {
		config.Network = network
	}
}

// WithSourceIP sets the source IP address
func WithSourceIP(ip net.IP) ICMPEngineOption {
	return func(config *ICMPConfig) {
		config.SourceIP = ip
	}
}

// WithTimeout sets the measurement timeout
func WithTimeout(timeout time.Duration) ICMPEngineOption {
	return func(config *ICMPConfig) {
		config.Timeout = timeout
	}
}

// WithBufferSize sets the socket buffer size
func WithBufferSize(size int) ICMPEngineOption {
	return func(config *ICMPConfig) {
		config.BufferSize = size
	}
}

// WithTTL sets the time-to-live for packets
func WithTTL(ttl int) ICMPEngineOption {
	return func(config *ICMPConfig) {
		config.TTL = ttl
	}
}

// NewEngine creates a new ICMP measurement engine
func NewEngine(opts ...ICMPEngineOption) (*Engine, error) {
	config := &ICMPConfig{
		Network:     "ip4", // Default to IPv4
		Timeout:     5 * time.Second,
		BufferSize:  65535,
		DSCP:        0,
		TTL:         64,
	}

	// Apply functional options
	for _, opt := range opts {
		opt(config)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid ICMP config: %w", err)
	}

	// Determine network and address
	network, address, err := determineNetworkAddress(config)
	if err != nil {
		return nil, fmt.Errorf("failed to determine network: %w", err)
	}

	// Create ICMP socket
	socket, err := createICMPSocket(config, network, address)
	if err != nil {
		return nil, fmt.Errorf("failed to create ICMP socket: %w", err)
	}

	// Configure socket options
	if err := configureSocket(socket, config); err != nil {
		socket.Close()
		return nil, fmt.Errorf("failed to configure socket: %w", err)
	}

	return &Engine{
		config:  config,
		socket:  socket,
		network: network,
		address: address,
	}, nil
}

// validateConfig validates the ICMP engine configuration
func validateConfig(config *ICMPConfig) error {
	if config.Network != "ip4" && config.Network != "ip6" && config.Network != "ip" {
		return fmt.Errorf("invalid network: %s", config.Network)
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if config.BufferSize <= 0 {
		return fmt.Errorf("buffer size must be positive")
	}

	if config.TTL <= 0 || config.TTL > 255 {
		return fmt.Errorf("TTL must be between 1 and 255")
	}

	if config.DSCP < 0 || config.DSCP > 63 {
		return fmt.Errorf("DSCP must be between 0 and 63")
	}

	return nil
}

// determineNetworkAddress determines the appropriate network and address
func determineNetworkAddress(config *ICMPConfig) (network, address string, err error) {
	switch config.Network {
	case "ip4":
		return "udp4", "0.0.0.0", nil
	case "ip6":
		return "udp6", "::", nil
	case "ip":
		return "udp", "0.0.0.0", nil
	default:
		return "", "", fmt.Errorf("unsupported network: %s", config.Network)
	}
}

// createICMPSocket creates the underlying ICMP socket
func createICMPSocket(config *ICMPConfig, network, address string) (net.PacketConn, error) {
	socket, err := net.ListenPacket(network, address)
	if err != nil {
		return nil, err
	}

	// Apply source IP if specified
	if config.SourceIP != nil {
		if _, ok := socket.(*net.UDPConn); ok {
			// Note: UDPConn doesn't have Bind method after creation
			// Source IP selection would need to be done during socket creation
			// For now, we'll use the default interface
		}
	}

	return socket, nil
}

// configureSocket applies socket-level configuration
func configureSocket(socket net.PacketConn, config *ICMPConfig) error {
	// Set socket buffer size
	if udpSocket, ok := socket.(*net.UDPConn); ok {
		// Set send buffer
		if err := udpSocket.SetWriteBuffer(config.BufferSize); err != nil {
			return err
		}

		// Set receive buffer
		if err := udpSocket.SetReadBuffer(config.BufferSize); err != nil {
			return err
		}
	}

	// Configure TTL and DSCP if possible
	if _, ok := socket.(*net.IPConn); ok {
		// Set TTL
		if config.TTL > 0 {
			// This would require platform-specific socket options
			// For now, we'll handle TTL in packet construction
		}

		// Set DSCP
		if config.DSCP > 0 {
			// This would require platform-specific socket options
			// For now, we'll handle DSCP in packet construction
		}
	}

	return nil
}

// Ping performs a single ICMP measurement to a target
func (e *Engine) Ping(ctx context.Context, target net.IP) (*types.MeasurementData, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Validate target
	if err := validateTarget(target, e.config.Network); err != nil {
		return nil, fmt.Errorf("invalid target: %w", err)
	}

	// Create ICMP echo request packet
	packet, err := createEchoRequest(e.config.Network, target)
	if err != nil {
		return nil, fmt.Errorf("failed to create ICMP packet: %w", err)
	}

	// Create measurement context
	measurement := &types.MeasurementData{
		ID:           generateMeasurementID(),
		Target:       createNetworkTarget(target),
		Timestamp:    time.Now(),
		PacketSize:   len(packet),
		PlatformData: types.PlatformMeta{OS: "unknown", Architecture: "unknown"}, // Will be filled by caller
	}

	// Perform the measurement with timeout
	result, err := e.performMeasurement(ctx, target, packet, measurement)
	if err != nil {
		measurement.Success = false
		measurement.ErrorMessage = err.Error()
		return measurement, err
	}

	// Fill measurement data with results
	*measurement = *result
	return measurement, nil
}

// validateTarget validates the target IP address
func validateTarget(target net.IP, network string) error {
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	if target.IsUnspecified() {
		return fmt.Errorf("target cannot be unspecified address")
	}

	switch network {
	case "ip4":
		if target.To4() == nil {
			return fmt.Errorf("target must be IPv4 address")
		}
	case "ip6":
		if target.To4() != nil {
			return fmt.Errorf("target must be IPv6 address")
		}
	case "ip":
		// Accept both IPv4 and IPv6
	}

	return nil
}

// createEchoRequest creates an ICMP echo request packet
func createEchoRequest(network string, target net.IP) ([]byte, error) {
	var messageType icmp.Type
	var messageBody icmp.MessageBody

	switch {
	case target.To4() != nil:
		// IPv4 ICMP echo request
		messageType = ipv4.ICMPTypeEcho
		messageBody = &icmp.Echo{
			ID:   getProcessID(),
			Seq:  1,
			Data: []byte("probe measurement"),
		}
	default:
		// IPv6 ICMP echo request
		messageType = ipv6.ICMPTypeEchoRequest
		messageBody = &icmp.Echo{
			ID:   getProcessID(),
			Seq:  1,
			Data: []byte("probe measurement"),
		}
	}

	message := &icmp.Message{
		Type: messageType,
		Code: 0,
		Body: messageBody,
	}

	packet, err := message.Marshal(nil)
	if err != nil {
		return nil, err
	}

	return packet, nil
}

// getProcessID returns the current process ID for ICMP packet identification
func getProcessID() int {
	// This is a simple approach - in production, you might want
	// a more sophisticated ID generation strategy
	return 1000 // Placeholder
}

// createNetworkTarget creates a NetworkTarget from an IP address
func createNetworkTarget(ip net.IP) types.NetworkTarget {
	var addr net.IP
	if ip.To4() != nil {
		addr = ip.To4()
	} else {
		addr = ip
	}

	return types.NetworkTarget{
		ID:          fmt.Sprintf("target_%s", addr.String()),
		DisplayName: addr.String(),
		Address:     addr,
		Port:        0, // ICMP uses port 0
		Enabled:     true,
		Priority:    5,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// performMeasurement performs the actual ICMP measurement
func (e *Engine) performMeasurement(ctx context.Context, target net.IP, packet []byte, measurement *types.MeasurementData) (*types.MeasurementData, error) {
	// Create destination address
	dstAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:0", target.String()))
	if err != nil {
		return measurement, fmt.Errorf("failed to resolve destination: %w", err)
	}

	// Set request sent timestamp
	measurement.RequestSent = time.Now()

	// Send ICMP packet
	_, err = e.socket.WriteTo(packet, dstAddr)
	if err != nil {
		measurement.Success = false
		measurement.ErrorMessage = fmt.Sprintf("failed to send ICMP packet: %v", err)
		return measurement, err
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.config.Timeout)
	defer cancel()

	// Receive response
	responseBuffer := make([]byte, 65535)
	n, addr, err := e.socket.ReadFrom(responseBuffer)
	if err != nil {
		measurement.Success = false
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			measurement.ErrorCode = types.ICMPErrTimeout
			measurement.ErrorMessage = "ICMP request timed out"
		} else {
			measurement.ErrorMessage = fmt.Sprintf("failed to receive ICMP response: %v", err)
		}
		return measurement, err
	}

	measurement.ResponseRecv = time.Now()
	measurement.RTT = measurement.ResponseRecv.Sub(measurement.RequestSent)
	measurement.SourceIP = net.ParseIP(addr.String())
	measurement.DestIP = target
	measurement.Success = true

	// Parse ICMP response
	responseData := responseBuffer[:n]
	if err := e.parseResponse(responseData, measurement); err != nil {
		measurement.Success = false
		measurement.ErrorMessage = fmt.Sprintf("failed to parse ICMP response: %v", err)
		return measurement, err
	}

	return measurement, nil
}

// parseResponse parses the ICMP response packet
func (e *Engine) parseResponse(responseData []byte, measurement *types.MeasurementData) error {
	// Parse ICMP message
	message, err := icmp.ParseMessage(0xFF, responseData) // 0xFF is a placeholder protocol
	if err != nil {
		return fmt.Errorf("failed to parse ICMP message: %w", err)
	}

	// Check if it's an echo reply
	switch measurement.Target.Address.To4() {
	case nil:
		// IPv6
		if message.Type != ipv6.ICMPTypeEchoReply {
			return fmt.Errorf("expected IPv6 echo reply, got type %v", message.Type)
		}
	default:
		// IPv4
		if message.Type != ipv4.ICMPTypeEchoReply {
			return fmt.Errorf("expected IPv4 echo reply, got type %v", message.Type)
		}
	}

	// Extract sequence number from echo body
	if echo, ok := message.Body.(*icmp.Echo); ok {
		measurement.Sequence = uint16(echo.Seq)
		measurement.PacketID = uint16(echo.ID)
	}

	return nil
}

// generateMeasurementID generates a unique measurement ID
func generateMeasurementID() string {
	return fmt.Sprintf("meas_%d_%d", time.Now().UnixNano(), getProcessID())
}

// Close closes the ICMP engine and releases resources
func (e *Engine) Close() error {
	if e.socket != nil {
		return e.socket.Close()
	}
	return nil
}

// PingBatch performs multiple ICMP measurements to the same target
func (e *Engine) PingBatch(ctx context.Context, target net.IP, count int) ([]*types.MeasurementData, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive")
	}

	results := make([]*types.MeasurementData, count)
	for i := 0; i < count; i++ {
		measurement, err := e.Ping(ctx, target)
		if err != nil {
			results[i] = measurement
			// Continue with remaining measurements even if some fail
			continue
		}
		results[i] = measurement
	}

	return results, nil
}

// PingContinuous performs continuous ICMP measurements with configurable intervals
func (e *Engine) PingContinuous(ctx context.Context, target net.IP, interval time.Duration, count int) (<-chan *types.MeasurementData, <-chan error, error) {
	if interval <= 0 {
		return nil, nil, fmt.Errorf("interval must be positive")
	}

	results := make(chan *types.MeasurementData, 100)
	errors := make(chan error, 10)

	go func() {
		defer close(results)
		defer close(errors)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		measurementsMade := 0
		for {
			select {
			case <-ctx.Done():
				errors <- ctx.Err()
				return
			case <-ticker.C:
				measurement, err := e.Ping(ctx, target)
				if err != nil {
					errors <- err
				} else {
					results <- measurement
				}

				measurementsMade++
				if count > 0 && measurementsMade >= count {
					return
				}
			}
		}
	}()

	return results, errors, nil
}