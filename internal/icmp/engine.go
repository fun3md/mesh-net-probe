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