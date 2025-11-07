package auth

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mesh-net-probe/probe/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Middleware struct {
	logger         logger.Logger
	mu             sync.RWMutex
	rateLimitMap   map[string]time.Time
	allowedOrigins []string

	// probeSharedKey is a shared secret used to authenticate probe/agent calls.
	// It is loaded from the PROBE_SHARED_API_KEY environment variable.
	probeSharedKey string
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewMiddleware creates a new authentication middleware
func NewMiddleware() *Middleware {
	return &Middleware{
		logger:       logger.GetGlobalLogger(),
		rateLimitMap: make(map[string]time.Time),
		allowedOrigins: []string{
			"http://localhost:3001",  // React dev server
			"http://localhost:5173",  // Vite dev server
			"https://localhost:3001", // HTTPS dev server
		},
		probeSharedKey: strings.TrimSpace(os.Getenv("PROBE_SHARED_API_KEY")),
	}
}

// JWT creates JWT authentication middleware
func (m *Middleware) JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			c.Abort()
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := m.validateToken(token)
		if err != nil {
			m.logger.Error(c.Request.Context(), err, "JWT validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			c.Abort()
			return
		}

		// Add claims to context
		c.Set("user", claims)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// ProbeSharedKey returns middleware that authenticates probes/agents
// using a shared secret from PROBE_SHARED_API_KEY.
//
// Behavior:
// - If PROBE_SHARED_API_KEY is NOT set, this middleware becomes a no-op (allows all).
//   This is intentional for local/dev/test environments.
// - If PROBE_SHARED_API_KEY is set, it expects:
//     Authorization: Bearer <PROBE_SHARED_API_KEY>
func (m *Middleware) ProbeSharedKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If no shared key configured, do not enforce; let requests pass through.
		if m.probeSharedKey == "" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			c.Abort()
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		if subtleConstantTimeCompare(token, m.probeSharedKey) {
			c.Set("probe_authenticated", true)
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid probe shared key",
		})
		c.Abort()
	}
}

// subtleConstantTimeCompare performs a constant-time comparison for two strings.
func subtleConstantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	// Simple constant-time loop; avoids importing crypto/subtle just for this.
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

// CORS creates CORS middleware
func (m *Middleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		m.mu.RLock()
		defer m.mu.RUnlock()

		// Check if origin is allowed
		isAllowed := false
		for _, allowedOrigin := range m.allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			// In development, allow localhost origins
			if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "https://localhost:") {
				c.Header("Access-Control-Allow-Origin", origin)
			} else {
				c.Header("Access-Control-Allow-Origin", "")
			}
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// RateLimit creates rate limiting middleware
func (m *Middleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		m.mu.Lock()
		defer m.mu.Unlock()

		// Check if client has exceeded rate limit
		if lastRequest, exists := m.rateLimitMap[clientIP]; exists {
			// Allow 100 requests per minute
			if now.Sub(lastRequest) < time.Millisecond*600 {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "rate limit exceeded",
				})
				c.Abort()
				return
			}
		}

		// Update last request time
		m.rateLimitMap[clientIP] = now

		c.Next()
	}
}

// validateToken validates a JWT token
func (m *Middleware) validateToken(tokenString string) (*Claims, error) {
	// In a real implementation, you would load the secret from environment
	secret := []byte("your-secret-key") // Replace with environment variable

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrInvalidKey
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrInvalidKey
	}

	return claims, nil
}

// GenerateToken generates a JWT token for a user
func (m *Middleware) GenerateToken(userID, username, role string) (string, error) {
	secret := []byte("your-secret-key") // Replace with environment variable

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mesh-probe-admin",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// AuthenticateRequest authenticates a request and returns claims
func (m *Middleware) AuthenticateRequest(r *http.Request) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, nil // No authentication required for some endpoints
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	return m.validateToken(token)
}

// RequireRole requires a specific role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "user role not found",
			})
			c.Abort()
			return
		}

		if userRole != role && userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "insufficient permissions",
				"required": role,
				"actual":   userRole,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin requires admin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// RequireOperator requires operator or admin role
func RequireOperator() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "user role not found",
			})
			c.Abort()
			return
		}

		allowedRoles := []string{"admin", "operator"}
		hasPermission := false
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "insufficient permissions",
				"required": "operator or admin",
				"actual":   userRole,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
