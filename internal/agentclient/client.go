package agentclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/mesh-net-probe/probe/internal/logger"
	"github.com/mesh-net-probe/probe/pkg/types"
)

// Client is a minimal backend client used by the probe daemon/agent to talk to the admin-web APIs.
// It is intentionally small and aligned with internal/web/api/routes.go semantics.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	log        *logger.ComponentLogger
	probeID    string
}

// Config holds configuration for the agent client.
type Config struct {
	BaseURL         string        // e.g. http://admin-web:8080/api/v1
	APIKey          string        // optional API key / token header (Authorization: Bearer ...)
	Timeout         time.Duration // per-request timeout
	TLSInsecureSkip bool          // allow self-signed certs in dev (trusted single-tenant)
	ProbeID         string        // stable identity for this probe instance
	LoggerManager   *logger.Manager
}

// NewClient creates a new agent backend client.
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("agentclient: base URL is required")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.TLSInsecureSkip, // controlled via config; safe in trusted single-tenant environments
		},
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	lm := cfg.LoggerManager
	if lm == nil {
		// Fallback minimal logger manager if not injected; ignore error.
		lm, _ = logger.NewManager(nil)
	}

	cl := &Client{
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		baseURL: trimTrailingSlash(cfg.BaseURL),
		apiKey:  cfg.APIKey,
		log:     lm.WithComponent("agentclient"),
		probeID: cfg.ProbeID,
	}

	return cl, nil
}

// RegisterProbe performs an idempotent registration of this probe using /probes semantics.
// It aligns with handleRegisterProbe and ProbeRegistry expectations.
func (c *Client) RegisterProbe(ctx context.Context, info *types.ProbeInstance) error {
	if info == nil {
		return fmt.Errorf("agentclient: probe info is required")
	}
	if info.ID == "" {
		info.ID = c.ensureProbeID()
	}

	// NOTE: types.ProbeInstance currently does not expose full metadata fields used by ProbeRegistry.
	// To avoid speculation, we populate only what we can infer and rely on backend/ProbeRegistry
	// to enrich or update as needed. This stays aligned with existing contracts without breaking types.
	ipAddr := detectLocalIP()

	payload := map[string]interface{}{
		"id":         info.ID,
		"name":       info.ID,
		"version":    "",        // unknown from ProbeInstance; left empty
		"platform":   "",        // can be extended when ProbeInstance exposes platform
		"arch":       "",        // can be extended when ProbeInstance exposes arch
		"ip_address": ipAddr,
		"tags":       []string{},          // reserved for future use
		"metadata":   map[string]any{},    // reserved for future use
	}

	reqCtx, cancel := context.WithTimeout(ctx, c.httpClient.Timeout)
	defer cancel()

	req, err := c.newJSONRequest(reqCtx, http.MethodPost, "/probes", payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error(reqCtx, err, "register", "failed to register probe", "probe_id", info.ID)
		return fmt.Errorf("agentclient: register probe request failed: %w", err)
	}
	defer resp.Body.Close()

	// Accept 201 (created) or 200 (already registered/updated).
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agentclient: unexpected register status %d", resp.StatusCode)
	}

	c.log.Info(reqCtx, "register", "probe registered", "probe_id", info.ID)
	return nil
}

// SendHeartbeat posts a heartbeat to /probes/:id/heartbeat.
func (c *Client) SendHeartbeat(ctx context.Context, probeID string) error {
	if probeID == "" {
		probeID = c.ensureProbeID()
	}
	path := fmt.Sprintf("/probes/%s/heartbeat", probeID)

	reqCtx, cancel := context.WithTimeout(ctx, c.httpClient.Timeout)
	defer cancel()

	req, err := c.newJSONRequest(reqCtx, http.MethodPost, path, map[string]any{})
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error(reqCtx, err, "heartbeat", "failed to send heartbeat", "probe_id", probeID)
		return fmt.Errorf("agentclient: heartbeat request failed: %w", err)
	}
	defer resp.Body.Close()

	// 200 OK expected; treat 404 as soft failure (probe not registered) so caller can re-register.
	if resp.StatusCode == http.StatusNotFound {
		c.log.Warn(reqCtx, "heartbeat", "probe not found on backend; re-registration required", "probe_id", probeID)
		return fmt.Errorf("agentclient: heartbeat rejected with 404 for probe %s", probeID)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("agentclient: unexpected heartbeat status %d", resp.StatusCode)
	}

	c.log.Debug(reqCtx, "heartbeat", "heartbeat acknowledged", "probe_id", probeID)
	return nil
}

// ReportConfigApplied notifies backend that this probe applied a configuration via /probes/:id/config-applied.
func (c *Client) ReportConfigApplied(ctx context.Context, probeID string, cfg *types.Configuration, source string, appliedAt time.Time) error {
	if probeID == "" {
		probeID = c.ensureProbeID()
	}
	if cfg == nil {
		return fmt.Errorf("agentclient: configuration is required for config-applied")
	}
	if appliedAt.IsZero() {
		appliedAt = time.Now()
	}

	body := map[string]any{
		"config_id":      cfg.ID,
		"config_version": cfg.Version,
		"config_source":  source,
		"applied_at":     appliedAt,
	}

	path := fmt.Sprintf("/probes/%s/config-applied", probeID)

	reqCtx, cancel := context.WithTimeout(ctx, c.httpClient.Timeout)
	defer cancel()

	req, err := c.newJSONRequest(reqCtx, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error(reqCtx, err, "config-applied", "failed to report config applied", "probe_id", probeID, "config_id", cfg.ID)
		return fmt.Errorf("agentclient: config-applied request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("agentclient: unexpected config-applied status %d", resp.StatusCode)
	}

	c.log.Info(reqCtx, "config-applied", "reported config applied",
		"probe_id", probeID,
		"config_id", cfg.ID,
		"config_version", cfg.Version,
		"source", source,
	)
	return nil
}

// ensureProbeID returns a stable probe ID, deriving from env/hostname if not set.
func (c *Client) ensureProbeID() string {
	if c.probeID != "" {
		return c.probeID
	}
	if envID := os.Getenv("PROBE_ID"); envID != "" {
		c.probeID = envID
		return c.probeID
	}
	hostname, _ := os.Hostname()
	c.probeID = fmt.Sprintf("%s-%d", hostname, time.Now().UnixNano())
	return c.probeID
}

// newJSONRequest creates an HTTP request with JSON body, standard headers, and auth.
func (c *Client) newJSONRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, fmt.Errorf("agentclient: failed to encode request: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("agentclient: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	return req, nil
}

// Helper utilities (internal to this package).

func trimTrailingSlash(s string) string {
	if s == "" {
		return s
	}
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func detectLocalIP() string {
	// Best-effort: find a non-loopback IPv4 for metadata.
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}
	return ""
}