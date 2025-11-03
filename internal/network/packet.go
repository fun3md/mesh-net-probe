package network

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
)

// PacketHandler handles ICMP packet parsing and validation
type PacketHandler interface {
	// ParsePacket parses a binary ICMP packet
	ParsePacket(data []byte) (*icmp.Message, error)

	// ValidatePacket performs validation on an ICMP message
	ValidatePacket(msg *icmp.Message) error

	// CalculateChecksum calculates the ICMP checksum for a packet
	CalculateChecksum(msg *icmp.Message) (uint16, error)

	// GetPacketType returns numeric packet type for comparison
	GetPacketType(msg *icmp.Message) int

	// IsEchoRequest checks if message is an echo request by type value
	IsEchoRequest(msg *icmp.Message) bool

	// IsEchoReply checks if message is an echo reply by type value
	IsEchoReply(msg *icmp.Message) bool

	// Matches checks if echo reply matches echo request
	Matches(request, reply *icmp.Message) bool

	// ExtractPacketInfo extracts metadata from a packet
	ExtractPacketInfo(msg *icmp.Message, sourceIP, destIP net.IP) *PacketInfo
}

// defaultPacketHandler implements the PacketHandler interface
type defaultPacketHandler struct{}

// NewPacketHandler creates a new packet handler
func NewPacketHandler() PacketHandler {
	return &defaultPacketHandler{}
}

// ParsePacket parses a binary ICMP packet
func (h *defaultPacketHandler) ParsePacket(data []byte) (*icmp.Message, error) {
	// Parse ICMP message (protocol number 1 = ICMP)
	msg, err := icmp.ParseMessage(1, data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ICMP packet: %w", err)
	}

	return msg, nil
}

// ValidatePacket performs validation on an ICMP message
func (h *defaultPacketHandler) ValidatePacket(msg *icmp.Message) error {
	if msg == nil {
		return fmt.Errorf("ICMP message is nil")
	}

	// Validate message type (0 = Echo Reply, 8 = Echo Request)
	typ := h.GetPacketType(msg)
	if typ != 0 && typ != 8 {
		return fmt.Errorf("unsupported ICMP message type: %d", typ)
	}

	// Validate message code
	if msg.Code != 0 {
		return fmt.Errorf("unsupported ICMP message code: %v", msg.Code)
	}

	// Validate body type
	switch msg.Body.(type) {
	case *icmp.Echo, *icmp.DstUnreach, *icmp.TimeExceeded, nil:
		// Valid body types
	default:
		return fmt.Errorf("unsupported ICMP body type: %T", msg.Body)
	}

	return nil
}

// GetPacketType returns numeric packet type for comparison
func (h *defaultPacketHandler) GetPacketType(msg *icmp.Message) int {
	if msg == nil {
		return -1
	}
	// Use the Protocol() method to get the numeric type
	return msg.Type.Protocol()
}

// CalculateChecksum calculates the ICMP checksum for a packet
func (h *defaultPacketHandler) CalculateChecksum(msg *icmp.Message) (uint16, error) {
	if msg == nil {
		return 0, fmt.Errorf("message is nil")
	}

	// Marshal the message to get the binary data
	data, err := msg.Marshal(nil)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal message: %w", err)
	}

	return checksum(data), nil
}

// IsEchoRequest checks if message is an echo request by type value
func (h *defaultPacketHandler) IsEchoRequest(msg *icmp.Message) bool {
	return msg != nil && msg.Type.Protocol() == 8 // 8 = Echo Request
}

// IsEchoReply checks if message is an echo reply by type value
func (h *defaultPacketHandler) IsEchoReply(msg *icmp.Message) bool {
	return msg != nil && msg.Type.Protocol() == 0 // 0 = Echo Reply
}

// Matches checks if echo reply matches echo request
func (h *defaultPacketHandler) Matches(request, reply *icmp.Message) bool {
	if !h.IsEchoRequest(request) || !h.IsEchoReply(reply) {
		return false
	}

	requestEcho, ok1 := request.Body.(*icmp.Echo)
	replyEcho, ok2 := reply.Body.(*icmp.Echo)

	return ok1 && ok2 && 
		requestEcho.ID == replyEcho.ID && 
		requestEcho.Seq == replyEcho.Seq
}

// ExtractPacketInfo extracts metadata from a packet
func (h *defaultPacketHandler) ExtractPacketInfo(msg *icmp.Message, sourceIP, destIP net.IP) *PacketInfo {
	if msg == nil {
		return nil
	}

	// Calculate packet size (simplified)
	size := 8 // ICMP header size
	if msg.Body != nil {
		// Try to get body length, fallback to estimate
		if echo, ok := msg.Body.(*icmp.Echo); ok {
			size += len(echo.Data) + 4 // 4 bytes for ID and sequence
		} else {
			size += 8 // Default estimate for other body types
		}
	}

	return &PacketInfo{
		Type:      msg.Type.Protocol(),
		Code:      int(msg.Code),
		Size:      size,
		SourceIP:  sourceIP,
		DestIP:    destIP,
		Timestamp: time.Now(),
	}
}

// checksum calculates the Internet checksum (RFC 1071)
func checksum(data []byte) uint16 {
	sum := uint32(0)
	length := len(data)

	// Add up 16-bit words
	for i := 0; i < length-1; i += 2 {
		sum += uint32(data[i])<<8 + uint32(data[i+1])
	}

	// Add left-over byte, if any
	if length%2 == 1 {
		sum += uint32(data[length-1]) << 8
	}

	// Fold 32-bit sum to 16 bits
	for (sum >> 16) > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}

	// One's complement
	sum = ^sum

	return uint16(sum)
}

// ICMPError represents ICMP-specific error information
type ICMPError struct {
	Code    int         // ICMP error code
	Message string      // Human-readable error message
	OrigAddr net.Addr   // Original destination address
	Time     time.Time  // When the error occurred
}

// ICMP error code constants
const (
	ICMPCodeNetUnreachable         = 0  // Destination network unreachable
	ICMPCodeHostUnreachable        = 1  // Destination host unreachable
	ICMPCodeProtocolUnreachable    = 2  // Destination protocol unreachable
	ICMPCodePortUnreachable        = 3  // Destination port unreachable
	ICMPCodeFragmentNeeded         = 4  // Fragmentation needed and DF set
	ICMPCodeSourceRouteFailed      = 5  // Source route failed
)

// NewDestinationUnreachable creates an ICMP destination unreachable error
func NewDestinationUnreachable(code int, origAddr net.Addr) *ICMPError {
	return &ICMPError{
		Code:    code,
		Message: fmt.Sprintf("destination unreachable: code %d", code),
		OrigAddr: origAddr,
		Time:     time.Now(),
	}
}

// NewTimeExceeded creates an ICMP time exceeded error
func NewTimeExceeded(code int, origAddr net.Addr) *ICMPError {
	return &ICMPError{
		Code:    code,
		Message: fmt.Sprintf("time exceeded: code %d", code),
		OrigAddr: origAddr,
		Time:     time.Now(),
	}
}

// String returns a string representation of the ICMP error
func (e *ICMPError) String() string {
	return fmt.Sprintf("ICMP Error: %s (code: %d, time: %s)", e.Message, e.Code, e.Time.Format(time.RFC3339Nano))
}

// IsDestinationUnreachable checks if this is a destination unreachable error
func (e *ICMPError) IsDestinationUnreachable() bool {
	return e.Code >= ICMPCodeNetUnreachable && e.Code <= ICMPCodeSourceRouteFailed
}

// IsTimeExceeded checks if this is a time exceeded error
func (e *ICMPError) IsTimeExceeded() bool {
	// Time exceeded codes are typically 0 or 1
	return e.Code >= 0 && e.Code <= 1
}

// PacketInfo provides metadata about an ICMP packet
type PacketInfo struct {
	Type      int         // ICMP packet type
	Code      int         // ICMP packet code
	Size      int         // Total packet size in bytes
	SourceIP  net.IP      // Source IP address
	DestIP    net.IP      // Destination IP address
	Timestamp time.Time   // When the packet was received
}

// String returns a string representation of the packet info
func (p *PacketInfo) String() string {
	return fmt.Sprintf("ICMP Packet: Type=%d, Code=%d, Size=%d, From=%s, To=%s, Time=%s",
		p.Type, p.Code, p.Size, p.SourceIP, p.DestIP, p.Timestamp.Format(time.RFC3339Nano))
}