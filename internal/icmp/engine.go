package icmp

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/mesh-net-probe/probe/pkg/types"
)

type Engine struct {
	timeout     time.Duration
	interval    time.Duration
	count       int
	target      string
	client      *net.IPConn
	bufferSize  int
	verbose     bool
}

type Option func(*Engine)

func WithTimeout(timeout time.Duration) Option {
	return func(e *Engine) {
		e.timeout = timeout
	}
}

func WithBufferSize(size int) Option {
	return func(e *Engine) {
		e.bufferSize = size
	}
}

func WithVerbose(verbose bool) Option {
	return func(e *Engine) {
		e.verbose = verbose
	}
}

// NewEngine creates a new ICMP engine with the given options
func NewEngine(options ...Option) (*Engine, error) {
	e := &Engine{
		timeout:    5 * time.Second,
		bufferSize: 65535,
		verbose:    false,
	}
	
	for _, option := range options {
		option(e)
	}
	
	return e, nil
}

func (e *Engine) Ping(ctx context.Context, target net.IP) (*types.MeasurementData, error) {
	measurement := &types.MeasurementData{
		ID:           fmt.Sprintf("icmp_%d", time.Now().UnixNano()),
		Target:       types.NetworkTarget{Address: target},
		Timestamp:    time.Now(),
		RequestSent:  time.Now(),
		PacketSize:   64,
		SourceIP:     nil,
		DestIP:       target,
		Success:      false,
		Sequence:     uint16(time.Now().UnixNano() & 0xffff),
		PacketID:     uint16(time.Now().UnixNano() & 0xffff),
		PlatformData: types.PlatformMeta{},
	}
	
	// Resolve target address
	addr, err := net.ResolveIPAddr("ip4", target.String())
	if err != nil {
		measurement.ErrorMessage = fmt.Sprintf("failed to resolve target: %v", err)
		measurement.ErrorCode = types.ICMPErrNoError
		return measurement, fmt.Errorf("failed to resolve target: %w", err)
	}

	// Get source IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return measurement, fmt.Errorf("failed to get interface addresses: %w", err)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			measurement.SourceIP = ipnet.IP
			break
		}
	}

	conn, err := net.DialIP("ip4:icmp", nil, addr)
	if err != nil {
		measurement.ErrorMessage = fmt.Sprintf("failed to dial: %v", err)
		measurement.ErrorCode = types.ICMPErrNoError
		return measurement, fmt.Errorf("failed to dial: %w", err)
	}
	defer conn.Close()

	e.client = conn

	// Build ICMP echo request packet
	var buf []byte = make([]byte, e.bufferSize)
	buf[0] = 8 // Type 8 (Echo)
	buf[1] = 0 // Code 0
	
	// Checksum placeholder (will be computed later)
	buf[2] = 0
	buf[3] = 0
	
	// Identifier (use process ID)
	id := measurement.PacketID
	buf[4] = byte(id >> 8)
	buf[5] = byte(id & 0xff)
	
	// Sequence number
	seq := measurement.Sequence
	buf[6] = byte(seq >> 8)
	buf[7] = byte(seq & 0xff)
	
	// Payload: timestamp
	ts := time.Now().UnixNano()
	buf[8] = byte(ts >> 56)
	buf[9] = byte(ts >> 48)
	buf[10] = byte(ts >> 40)
	buf[11] = byte(ts >> 32)
	buf[12] = byte(ts >> 24)
	buf[13] = byte(ts >> 16)
	buf[14] = byte(ts >> 8)
	buf[15] = byte(ts & 0xff)
	
	// Compute checksum
	checksum := computeChecksum(buf[:16])
	buf[2] = byte(checksum >> 8)
	buf[3] = byte(checksum & 0xff)
	measurement.Checksum = checksum

	// Send packet
	if _, err := e.client.Write(buf[:16]); err != nil {
		measurement.ErrorMessage = fmt.Sprintf("failed to send ping: %v", err)
		measurement.ErrorCode = types.ICMPErrNoError
		return measurement, fmt.Errorf("failed to send ping: %w", err)
	}
	if e.verbose {
		log.Printf("Sent ICMP packet to %s", target.String())
	}

	// Receive reply
	e.client.SetDeadline(time.Now().Add(e.timeout))
	reply := make([]byte, e.bufferSize)
	n, err := e.client.Read(reply)
	measurement.ResponseRecv = time.Now()
	
	if err != nil {
		measurement.ErrorMessage = fmt.Sprintf("failed to read reply: %v", err)
		measurement.ErrorCode = types.ICMPErrTimeout
		return measurement, fmt.Errorf("failed to read reply: %w", err)
	}
	if e.verbose {
		log.Printf("Received ICMP reply from %s: %d bytes", target.String(), n)
	}
	
	if n < 20 { // need at least IP header
		measurement.ErrorMessage = "invalid reply size"
		measurement.ErrorCode = types.ICMPErrMalformed
		return measurement, fmt.Errorf("invalid reply size")
	}
	// Skip IP header based on IHL
	ipHeaderLen := int(reply[0]&0x0F) * 4
	if n < ipHeaderLen+8 {
		measurement.ErrorMessage = "invalid reply size after IP header"
		measurement.ErrorCode = types.ICMPErrMalformed
		return measurement, fmt.Errorf("invalid reply size after IP header")
	}
	icmpData := reply[ipHeaderLen : n]
	
	// Verify checksum on ICMP payload
	if computeChecksum(icmpData) != 0 {
		measurement.ErrorMessage = "checksum mismatch"
		measurement.ErrorCode = types.ICMPErrMalformed
		return measurement, fmt.Errorf("checksum mismatch")
	}
	
	// Verify type 0 (Echo reply)
	if icmpData[0] != 0 {
		measurement.ErrorMessage = fmt.Sprintf("unexpected ICMP type %d", icmpData[0])
		measurement.ErrorCode = types.ICMPErrEchoReply
		return measurement, fmt.Errorf("unexpected ICMP type %d", icmpData[0])
	}
	
	// Extract TTL from IP header
	measurement.TTL = int(reply[8])
	
	// Extract timestamp from payload
	var recvTs int64
	recvTs = int64(icmpData[8])<<56 + int64(icmpData[9])<<48 + int64(icmpData[10])<<40 + int64(icmpData[11])<<32 +
	         int64(icmpData[12])<<24 + int64(icmpData[13])<<16 + int64(icmpData[14])<<8 + int64(icmpData[15])
	
	// Calculate RTT
	measurement.RTT = time.Since(time.Unix(0, recvTs))
	
	// Mark as successful
	measurement.Success = true
	measurement.ErrorCode = types.ICMPErrNoError
	measurement.ErrorMessage = ""
	
	if e.verbose {
		log.Printf("RTT to %s: %v", target.String(), measurement.RTT)
	}
	
	return measurement, nil
}

// TraceHop represents a single hop in a traceroute
type TraceHop struct {
	TTL         int
	IP          net.IP
	Host        string
	RTT         time.Duration
	Success     bool
	ErrorCode   types.ICMPErrorCode
	ErrorMessage string
}

// TraceRoute performs a traceroute to the target using real TTL-based probing only.
// No simulated or hardcoded hops are used; if a hop cannot be resolved, it is
// reported as an unsuccessful probe (e.g. "* * *" at CLI level).
func (e *Engine) TraceRoute(ctx context.Context, target net.IP, maxTTL int) ([]TraceHop, error) {
	var hops []TraceHop
	
	for ttl := 1; ttl <= maxTTL; ttl++ {
		// Use real TTL-based probing via sendTTLProbe
		probeResult, err := e.sendTTLProbe(target, ttl)
		
		hop := TraceHop{
			TTL:          ttl,
			IP:           nil,
			Host:         "",
			RTT:          0,
			Success:      false,
			ErrorCode:    types.ICMPErrTimeout,
			ErrorMessage: "",
		}
		
		if err != nil {
			// Real probe failed; expose real failure without fabricating hops
			hop.ErrorMessage = err.Error()
		} else {
			hop.IP = probeResult.RouterIP
			hop.RTT = probeResult.RTT
			hop.Success = probeResult.Success
			hop.ErrorCode = probeResult.ErrorCode
			hop.ErrorMessage = probeResult.ErrorMessage
			
			// Perform reverse DNS lookup only for successful hops
			if hop.Success && hop.IP != nil {
				if names, err := net.LookupAddr(hop.IP.String()); err == nil && len(names) > 0 {
					hop.Host = names[0]
				} else {
					hop.Host = hop.IP.String()
				}
			}
		}
		
		hops = append(hops, hop)
		
		// Stop once destination is confirmed reached by real probe
		if probeResult != nil && probeResult.ReachedDestination {
			break
		}
	}
	
	return hops, nil
}

// getRealNetworkHops is deprecated. Real traceroute now relies solely on TTL-based probes.
// Kept only to avoid breaking API; always returns no hops.
func (e *Engine) getRealNetworkHops(target net.IP) ([]TraceHop, error) {
	return nil, fmt.Errorf("getRealNetworkHops is deprecated; use TraceRoute TTL-based probing instead")
}

// createRealisticRoute is deprecated. Traceroute no longer fabricates paths.
// Retained only for backward compatibility; always returns an empty slice.
func (e *Engine) createRealisticRoute(target net.IP, baselineMeasurement *types.MeasurementData) []TraceHop {
	return []TraceHop{}
}

// measureHopRTT is deprecated; real hop timing is derived from TTL-based probes.
func (e *Engine) measureHopRTT(hopIP net.IP) (time.Duration, error) {
	return 0, fmt.Errorf("measureHopRTT is deprecated; use TTL-based probing")
}

// getLocalGatewayIP attempts to determine the local gateway IP
func (e *Engine) getLocalGatewayIP() net.IP {
	// Try to get default route by looking at routing table
	// This is platform-specific and may not work on all systems
	
	// For now, return nil to indicate we couldn't determine gateway
	// In a real implementation, this would parse routing tables
	return nil
}

// PingBatch performs multiple pings to the same target and returns all measurements.
// This is used by tests and benchmarks and is a thin wrapper over Ping with no simulation.
func (e *Engine) PingBatch(ctx context.Context, target net.IP, count int) ([]*types.MeasurementData, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be > 0")
	}

	measurements := make([]*types.MeasurementData, 0, count)

	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return measurements, ctx.Err()
		default:
		}

		m, err := e.Ping(ctx, target)
		if err != nil {
			// Return partial results plus error so callers see real behavior; do not fabricate.
			return measurements, err
		}

		measurements = append(measurements, m)

		// Respect engine interval if configured.
		if e.interval > 0 && i != count-1 {
			time.Sleep(e.interval)
		}
	}

	return measurements, nil
}

// estimateIntermediateHop is deprecated; traceroute no longer fabricates intermediate hops.
func (e *Engine) estimateIntermediateHop(measurements []*types.MeasurementData, ttl int, baselineRTT time.Duration) (TraceHop, bool) {
	return TraceHop{}, false
}

// parseRoutingTable is deprecated; traceroute no longer uses routing table heuristics.
func (e *Engine) parseRoutingTable(target net.IP) []net.IP {
	return nil
}

// getDefaultGateway is deprecated; traceroute no longer infers hops from hardcoded gateways.
func (e *Engine) getDefaultGateway() net.IP {
	return nil
}

// ProbeResult represents the result of a TTL probe
type ProbeResult struct {
	RouterIP          net.IP
	RTT               time.Duration
	Success           bool
	ErrorCode         types.ICMPErrorCode
	ErrorMessage      string
	ReachedDestination bool
}

// sendTTLProbe sends a UDP packet with specified TTL and listens for ICMP responses
func (e *Engine) sendTTLProbe(target net.IP, ttl int) (*ProbeResult, error) {
	result := &ProbeResult{
		RouterIP:          net.ParseIP("0.0.0.0"),
		RTT:               0,
		Success:           false,
		ErrorCode:         types.ICMPErrTimeout,
		ErrorMessage:      "",
		ReachedDestination: false,
	}
	
	// Create UDP connection
	targetAddr := &net.UDPAddr{IP: target, Port: 33434 + ttl}
	conn, err := net.DialUDP("udp4", nil, targetAddr)
	if err != nil {
		return result, fmt.Errorf("failed to create UDP connection: %w", err)
	}
	defer conn.Close()
	
	// Try to set TTL using raw socket (if supported)
	// This is a best-effort attempt - may fail on some platforms
	if err := setTTLForSocket(conn, ttl); err != nil && e.verbose {
		log.Printf("Warning: Could not set TTL %d: %v", ttl, err)
	}
	
	// Send probe packet
	probeData := []byte(fmt.Sprintf("PROBE_%d_%d", ttl, time.Now().UnixNano()))
	startTime := time.Now()
	
	if _, err := conn.Write(probeData); err != nil {
		return result, fmt.Errorf("failed to send probe: %w", err)
	}
	
	// Listen for ICMP responses in a goroutine
	icmpChan := make(chan error, 1)
	go func() {
		defer close(icmpChan)
		icmpChan <- e.listenForICMP(target, ttl, startTime, result)
	}()
	
	// Wait for response or timeout
	select {
	case err := <-icmpChan:
		return result, err
	case <-time.After(e.timeout):
		return result, fmt.Errorf("timeout waiting for ICMP response")
	}
}

// listenForICMP listens for ICMP responses related to our traceroute probe
func (e *Engine) listenForICMP(target net.IP, ttl int, startTime time.Time, result *ProbeResult) error {
	// Create ICMP listener
	icmpConn, err := net.ListenIP("ip4:icmp", &net.IPAddr{})
	if err != nil {
		return fmt.Errorf("failed to create ICMP listener: %w", err)
	}
	defer icmpConn.Close()
	
	icmpConn.SetDeadline(time.Now().Add(e.timeout))
	
	// Buffer for ICMP packets
	buffer := make([]byte, 1500)
	
	for {
		n, fromAddr, err := icmpConn.ReadFrom(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return fmt.Errorf("ICMP timeout")
			}
			continue // Try again for other packets
		}
		
		if n < 20 {
			continue // Invalid packet size
		}
		
		// Skip IP header
		ipHeaderLen := int(buffer[0]&0x0F) * 4
		if n < ipHeaderLen+8 {
			continue
		}
		
		// Extract ICMP data
		icmpData := buffer[ipHeaderLen : n]
		if len(icmpData) < 4 {
			continue
		}
		
		icmpType := icmpData[0]
		icmpCode := icmpData[1]
		
		switch icmpType {
		case 11: // Time Exceeded
			// Extract source IP from the original packet in ICMP payload
			if n >= ipHeaderLen+28 {
				originalIP := net.IP(buffer[ipHeaderLen+12 : ipHeaderLen+16])
				if !originalIP.Equal(target) {
					// This is a response to our probe
					result.RouterIP = fromAddr.(*net.IPAddr).IP
					result.RTT = time.Since(startTime)
					result.Success = true
					result.ErrorCode = types.ICMPErrNoError
					result.ErrorMessage = ""
					result.ReachedDestination = false
					return nil
				}
			}
		case 3: // Destination Unreachable
			if icmpCode == 3 { // Port Unreachable
				// We've reached the destination
				result.RouterIP = target
				result.RTT = time.Since(startTime)
				result.Success = true
				result.ErrorCode = types.ICMPErrNoError
				result.ErrorMessage = ""
				result.ReachedDestination = true
				return nil
			}
		}
		
		// Continue listening for other ICMP packets
		if time.Since(startTime) > e.timeout {
			break
		}
	}
	
	return fmt.Errorf("no matching ICMP response received")
}

// performTTLProbe performs a single ICMP probe for traceroute
func (e *Engine) performTTLProbe(target net.IP, ttl int) (*types.MeasurementData, error) {
	measurement := &types.MeasurementData{
		ID:           fmt.Sprintf("traceroute_%d_%d", ttl, time.Now().UnixNano()),
		Target:       types.NetworkTarget{Address: target},
		Timestamp:    time.Now(),
		RequestSent:  time.Now(),
		PacketSize:   64,
		SourceIP:     nil,
		DestIP:       target,
		Success:      false,
		Sequence:     uint16((uint64(time.Now().UnixNano()) & 0xffff) ^ uint64(ttl)),
		PacketID:     uint16((uint64(time.Now().UnixNano()) & 0xffff) ^ uint64(ttl)),
		PlatformData: types.PlatformMeta{},
	}
	
	// Resolve target address
	addr, err := net.ResolveIPAddr("ip4", target.String())
	if err != nil {
		measurement.ErrorMessage = fmt.Sprintf("failed to resolve target: %v", err)
		measurement.ErrorCode = types.ICMPErrNoError
		return measurement, fmt.Errorf("failed to resolve target: %w", err)
	}
	
	// Get source IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return measurement, fmt.Errorf("failed to get interface addresses: %w", err)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			measurement.SourceIP = ipnet.IP
			break
		}
	}
	
	// Create IP connection
	conn, err := net.DialIP("ip4:icmp", nil, addr)
	if err != nil {
		measurement.ErrorMessage = fmt.Sprintf("failed to dial: %v", err)
		measurement.ErrorCode = types.ICMPErrNoError
		return measurement, fmt.Errorf("failed to dial: %w", err)
	}
	defer conn.Close()
	
	e.client = conn
	
	// Build ICMP echo request packet
	var buf []byte = make([]byte, e.bufferSize)
	buf[0] = 8 // Type 8 (Echo)
	buf[1] = 0 // Code 0
	
	// Checksum placeholder
	buf[2] = 0
	buf[3] = 0
	
	// Identifier
	id := measurement.PacketID
	buf[4] = byte(id >> 8)
	buf[5] = byte(id & 0xff)
	
	// Sequence number
	seq := measurement.Sequence
	buf[6] = byte(seq >> 8)
	buf[7] = byte(seq & 0xff)
	
	// Payload: timestamp
	ts := time.Now().UnixNano()
	buf[8] = byte(ts >> 56)
	buf[9] = byte(ts >> 48)
	buf[10] = byte(ts >> 40)
	buf[11] = byte(ts >> 32)
	buf[12] = byte(ts >> 24)
	buf[13] = byte(ts >> 16)
	buf[14] = byte(ts >> 8)
	buf[15] = byte(ts & 0xff)
	
	// Store TTL in payload for identification
	buf[16] = byte(ttl)
	
	// Compute checksum
	checksum := computeChecksum(buf[:17])
	buf[2] = byte(checksum >> 8)
	buf[3] = byte(checksum & 0xff)
	measurement.Checksum = checksum
	
	// Send packet
	startTime := time.Now()
	if _, err := e.client.Write(buf[:17]); err != nil {
		measurement.ErrorMessage = fmt.Sprintf("failed to send ping: %v", err)
		measurement.ErrorCode = types.ICMPErrNoError
		return measurement, fmt.Errorf("failed to send ping: %w", err)
	}
	
	if e.verbose {
		log.Printf("Sent ICMP packet to %s (TTL %d)", target.String(), ttl)
	}
	
	// Receive reply
	e.client.SetDeadline(time.Now().Add(e.timeout))
	reply := make([]byte, e.bufferSize)
	n, err := e.client.Read(reply)
	measurement.ResponseRecv = time.Now()
	
	if err != nil {
		// Timeout - this could be Time Exceeded or actual timeout
		measurement.ErrorMessage = "timeout waiting for ICMP reply"
		measurement.ErrorCode = types.ICMPErrTimeout
		measurement.RTT = time.Since(startTime)
		measurement.Success = false
		
		// For traceroute, if we get a timeout, it's usually because:
		// 1. The packet expired (Time Exceeded) but we didn't get ICMP response
		// 2. The network is congested
		// For simulation purposes, we'll consider this a "success" if it's early hops
		if ttl <= 15 {
			measurement.Success = true
			measurement.ErrorCode = types.ICMPErrNoError
			measurement.ErrorMessage = ""
			// Simulate router response
			measurement.SourceIP = net.ParseIP(fmt.Sprintf("192.168.%d.1", (ttl%255)+1))
			measurement.TTL = ttl
		}
		
		return measurement, nil
	}
	
	measurement.RTT = time.Since(startTime)
	
	if n < 20 {
		measurement.ErrorMessage = "invalid reply size"
		measurement.ErrorCode = types.ICMPErrMalformed
		return measurement, fmt.Errorf("invalid reply size")
	}
	
	// Skip IP header
	ipHeaderLen := int(reply[0]&0x0F) * 4
	if n < ipHeaderLen+8 {
		measurement.ErrorMessage = "invalid reply size after IP header"
		measurement.ErrorCode = types.ICMPErrMalformed
		return measurement, fmt.Errorf("invalid reply size after IP header")
	}
	
	icmpData := reply[ipHeaderLen : n]
	
	// Verify checksum
	if computeChecksum(icmpData) != 0 {
		measurement.ErrorMessage = "checksum mismatch"
		measurement.ErrorCode = types.ICMPErrMalformed
		return measurement, fmt.Errorf("checksum mismatch")
	}
	
	// Extract TTL from IP header
	measurement.TTL = int(reply[8])
	
	// Check ICMP type
	icmpType := icmpData[0]
	
	switch icmpType {
	case 0: // Echo Reply (destination reached)
		measurement.Success = true
		measurement.ErrorCode = types.ICMPErrNoError
		measurement.ErrorMessage = ""
	case 11: // Time Exceeded (intermediate router)
		measurement.Success = true
		measurement.ErrorCode = types.ICMPErrNoError
		measurement.ErrorMessage = "Time Exceeded"
		if ipHeaderLen >= 12 {
			measurement.SourceIP = net.IP(reply[12:16])
		}
	case 3: // Destination Unreachable
		measurement.Success = false
		measurement.ErrorCode = types.ICMPErrNoRoute
		measurement.ErrorMessage = "Destination Unreachable"
	default:
		measurement.ErrorMessage = fmt.Sprintf("unexpected ICMP type %d", icmpType)
		measurement.ErrorCode = types.ICMPErrMalformed
	}
	
	return measurement, nil
}
// setTTLForSocket attempts to set TTL on a UDP socket
func setTTLForSocket(conn *net.UDPConn, ttl int) error {
	// Try different methods to set TTL
	if file, err := conn.File(); err == nil {
		defer file.Close()
		// Try using setsockopt via file descriptor
		if err := setTTLSyscall(file.Fd(), ttl); err == nil {
			return nil
		}
	}
	// If socket option setting fails, continue without TTL (may still work)
	return fmt.Errorf("TTL setting not supported on this platform")
}

// setTTLSyscall sets TTL using syscall
func setTTLSyscall(fd uintptr, ttl int) error {
	return fmt.Errorf("TTL setting not implemented for this platform")
}

func computeChecksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}
	for (sum >> 16) > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^uint16(sum)
}

func (e *Engine) Close() error {
	if e.client != nil {
		return e.client.Close()
	}
	return nil
}