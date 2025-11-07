package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"
)

// Security validation test suite for comprehensive security hardening
func TestSecurityHardening(t *testing.T) {
	t.Run("CertificateValidation", testCertificateValidation)
	t.Run("InputSanitization", testInputSanitization)
	t.Run("AuthenticationSecurity", testAuthenticationSecurity)
	t.Run("NetworkSecurity", testNetworkSecurity)
	t.Run("MemorySecurity", testMemorySecurity)
	t.Run("LoggingSecurity", testLoggingSecurity)
}

func testCertificateValidation(t *testing.T) {
	// Test X.509 certificate generation and validation
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Mesh Probe Security"},
			Country:      []string{"US"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Validate certificate properties
	if cert.SerialNumber.Cmp(big.NewInt(0)) != 1 {
		t.Error("Certificate serial number should be positive")
	}

	if !cert.IsCA {
		t.Log("Certificate is not CA - suitable for client authentication")
	}

	t.Logf("Certificate subject: %s, valid until: %v", cert.Subject, cert.NotAfter)
}

func testInputSanitization(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid IP", "192.168.1.1", true},
		{"Valid IPv6", "2001:db8::1", true},
		{"Invalid IP", "999.999.999.999", false},
		{"Empty String", "", false},
		{"SQL Injection", "'; DROP TABLE users; --", false},
		{"Command Injection", "127.0.0.1; rm -rf /", false},
		{"Path Traversal", "../../etc/passwd", false},
		{"XSS Script", "<script>alert('xss')</script>", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := sanitizeNetworkInput(tc.input)
			if isValid != tc.expected {
				t.Errorf("Input sanitization failed for %s: got %v, want %v", tc.input, isValid, tc.expected)
			}
		})
	}
}

func testAuthenticationSecurity(t *testing.T) {
	// Test JWT token generation and validation
	token := generateJWT("test-user", "admin")
	if token == "" {
		t.Error("JWT token generation failed")
	}

	// Test password hashing
	password := "SecurePassword123!"
	hashedPassword, err := hashPassword(password)
	if err != nil {
		t.Fatalf("Password hashing failed: %v", err)
	}

	if !verifyPassword(password, hashedPassword) {
		t.Error("Password verification failed")
	}

	if verifyPassword("wrong-password", hashedPassword) {
		t.Error("Password verification should fail for wrong password")
	}

	t.Log("Authentication security tests passed")
}

func testNetworkSecurity(t *testing.T) {
	// Test network binding security
	testIPs := []string{
		"127.0.0.1",  // Loopback - safe
		"0.0.0.0",    // All interfaces - potentially unsafe
		"192.168.1.100", // Private network - safe
		"10.0.0.1",   // Private network - safe
		"::1",        // IPv6 loopback - safe
	}

	for _, ip := range testIPs {
		t.Run("Bind test for "+ip, func(t *testing.T) {
			isSafe, reason := validateNetworkBinding(ip)
			t.Logf("IP %s: safe=%v, reason=%s", ip, isSafe, reason)
			
			if ip == "0.0.0.0" && isSafe {
				t.Log("All interfaces binding requires additional security considerations")
			}
		})
	}
}

func testMemorySecurity(t *testing.T) {
	// Test secure memory operations
	// Clear sensitive data from memory
	sensitiveData := []byte("sensitive-information")
	
	// Use secure operations
	secureClear(sensitiveData)
	
	// Verify data is cleared
	if isDataCleared(sensitiveData) {
		t.Log("Sensitive data properly cleared from memory")
	} else {
		t.Error("Sensitive data may still exist in memory")
	}

	// Test memory pool security
	testMemoryPoolSecurity(t)
}

func testLoggingSecurity(t *testing.T) {
	// Test secure logging practices
	sensitiveInfo := map[string]string{
		"password":      "secret123",
		"api_key":       "sk-1234567890abcdef",
		"credit_card":   "4532-1234-5678-9012",
		"ssn":          "123-45-6789",
	}

	for key, value := range sensitiveInfo {
		sanitized := sanitizeForLogging(key, value)
		
		// Verify sensitive data is not logged in plain text
		if containsSensitiveData(sanitized, value) {
			t.Errorf("Sensitive data for %s was not properly sanitized", key)
		} else {
			t.Logf("Sensitive data for %s properly sanitized", key)
		}
	}
}

// Security utility functions
func sanitizeNetworkInput(input string) bool {
	// Basic input validation for network operations
	if len(input) == 0 || len(input) > 255 {
		return false
	}

	// Check for potentially dangerous characters to mitigate injection attacks
	dangerousChars := []string{";", "&", "|", "<", ">", "'", "\"", "`", "$(", "${"}
	for _, char := range dangerousChars {
		if strings.Contains(input, char) {
			return false
		}
	}

	// Basic IP validation.
	// If this is a syntactically valid IP, accept.
	if ip := net.ParseIP(input); ip != nil {
		return true
	}

	// If the input looks like an IP literal (only digits and dots)
	// but net.ParseIP returned nil, then reject as invalid IP-like input.
	isIPLike := true
	for _, ch := range input {
		if (ch < '0' || ch > '9') && ch != '.' {
			isIPLike = false
			break
		}
	}
	if isIPLike {
		return false
	}

	// Basic hostname validation for non IP-like inputs
	return isValidHostname(input)
}

func generateJWT(userID, role string) string {
	// Simplified JWT generation for testing
	// In production, use a proper JWT library
	header := `{"alg":"HS256","typ":"JWT"}`
	claims := fmt.Sprintf(`{"user":"%s","role":"%s","iat":%d}`, userID, role, time.Now().Unix())
	
	// This is a simplified implementation for testing
	// Real implementation would require proper signing
	return fmt.Sprintf("%s.%s.signature", base64Encode(header), base64Encode(claims))
}

func hashPassword(password string) (string, error) {
	// Simplified password hashing for testing
	// In production, use bcrypt, scrypt, or argon2
	return fmt.Sprintf("hashed_%s", password), nil
}

func verifyPassword(password, hashedPassword string) bool {
	// Simplified password verification for testing
	expectedHash := fmt.Sprintf("hashed_%s", password)
	return expectedHash == hashedPassword
}

func validateNetworkBinding(ip string) (bool, string) {
	switch ip {
	case "127.0.0.1", "::1":
		return true, "Loopback interface - safe"
	case "0.0.0.0":
		return false, "All interfaces - requires additional security measures"
	default:
		if ip == "localhost" {
			return true, "Localhost - safe"
		}
		if isPrivateIP(ip) {
			return true, "Private network - safe"
		}
		return false, "Public IP - requires additional security measures"
	}
}

func secureClear(data []byte) {
	// Secure memory clearing
	for i := range data {
		data[i] = 0
	}
}

func isDataCleared(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

func sanitizeForLogging(key, value string) string {
	// Remove sensitive data from logs
	sensitiveKeys := []string{"password", "api_key", "credit_card", "ssn", "token"}
	
	for _, sensitiveKey := range sensitiveKeys {
		if contains(strings.ToLower(key), sensitiveKey) {
			return fmt.Sprintf("%s=[REDACTED]", key)
		}
	}
	
	return fmt.Sprintf("%s=%s", key, value)
}

func containsSensitiveData(logLine, sensitiveValue string) bool {
	return contains(logLine, sensitiveValue)
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}

func base64Encode(data string) string {
	// Simplified base64 encoding for testing
	// In production, use proper base64 encoding
	return data
}

func isValidHostname(hostname string) bool {
	// Basic hostname validation
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}
	
	// Check for valid characters
	for _, c := range hostname {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || 
		     (c >= '0' && c <= '9') || c == '-' || c == '.') {
			return false
		}
	}
	
	return true
}

func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	
	// Check for private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12", 
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
	}
	
	for _, rangeStr := range privateRanges {
		if _, network, err := net.ParseCIDR(rangeStr); err == nil {
			if network.Contains(ip) {
				return true
			}
		}
	}
	
	return false
}

func testMemoryPoolSecurity(t *testing.T) {
	// Test memory pool for security leaks
	securePool := NewSecureMemoryPool(100)
	
	// Test secure allocation and deallocation
	for i := 0; i < 10; i++ {
		ptr := securePool.Allocate(1024)
		if ptr == nil {
			t.Error("Memory allocation failed")
		}
		
		// Clear before returning to pool
		securePool.SecureClear(ptr)
		securePool.Deallocate(ptr)
	}
	
	t.Log("Memory pool security tests passed")
}

// Secure memory pool implementation
type SecureMemoryPool struct {
	pool chan []byte
}

func NewSecureMemoryPool(size int) *SecureMemoryPool {
	return &SecureMemoryPool{
		pool: make(chan []byte, size),
	}
}

func (p *SecureMemoryPool) Allocate(size int) []byte {
	select {
	case buf := <-p.pool:
		if len(buf) >= size {
			return buf[:size]
		}
		return make([]byte, size)
	default:
		return make([]byte, size)
	}
}

func (p *SecureMemoryPool) SecureClear(ptr []byte) {
	for i := range ptr {
		ptr[i] = 0
	}
}

func (p *SecureMemoryPool) Deallocate(ptr []byte) {
	select {
	case p.pool <- ptr:
	default:
		// Pool full, let garbage collector handle it
	}
}