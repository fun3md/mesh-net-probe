package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mesh-net-probe/probe/internal/config"
	"github.com/mesh-net-probe/probe/internal/monitoring"
	"github.com/mesh-net-probe/probe/internal/web/api"
	types "github.com/mesh-net-probe/probe/pkg/types"
)

// fakeManager is an in-memory implementation of config.Manager for contract tests.
type fakeManager struct {
	cfg     *types.Configuration
	status  *config.ManagerStatus
	reloads int
}

func newFakeManager() *fakeManager {
	return &fakeManager{
		status: &config.ManagerStatus{
			Sources: make(map[string]config.SourceStatus),
		},
	}
}

func (f *fakeManager) Initialize(ctx context.Context, cfg *types.Configuration) error {
	f.cfg = cfg
	return nil
}

func (f *fakeManager) GetConfiguration(ctx context.Context) (*types.Configuration, error) {
	if f.cfg == nil {
		return nil, fmt.Errorf("no configuration available")
	}
	return f.cfg, nil
}

func (f *fakeManager) WatchConfiguration(ctx context.Context, handler config.ConfigurationHandler) error {
	return nil
}

func (f *fakeManager) ReloadConfiguration(ctx context.Context) error {
	f.reloads++
	return nil
}

func (f *fakeManager) GetStatus(ctx context.Context) (*config.ManagerStatus, error) {
	return f.status, nil
}

func (f *fakeManager) Close(ctx context.Context) error { return nil }

// newTestEngine creates a Gin engine with stubbed auth behavior.
// Phase 5.1 contract tests focus on Manager + ProbeRegistry usage, not JWT/roles.
func newTestEngine(fm *fakeManager, reg *monitoring.ProbeRegistry) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Use a group that injects a no-op "authenticated" context so that handlers
	// gated by JWT/roles in production are reachable in tests.
	noAuthGroup := r.Group("")
	noAuthGroup.Use(func(c *gin.Context) {
		// In real system, auth middleware would validate JWT and set roles.
		// For these contract tests we deliberately bypass it.
		c.Next()
	})

	// Register config and probe routes against the no-auth group.
	// The authMiddleware argument is nil because tests are not exercising it.
	api.RegisterConfigRoutes(noAuthGroup, fm, nil)
	api.RegisterProbeRoutes(noAuthGroup, reg, fm, nil)

	return r
}

// getAuthToken is not needed when auth is bypassed; kept only for compatibility.
func getAuthToken(_ *testing.T, _ *gin.Engine) string {
	return ""
}

// TestConfigStatus_UsesManager validates T094: /config/status exposes ManagerStatus via Manager.
func TestConfigStatus_UsesManager(t *testing.T) {
	fm := newFakeManager()
	fm.status.CurrentConfigID = "cfg-123"
	fm.status.HealthScore = 0.75

	reg := monitoring.NewProbeRegistry()
	r := newTestEngine(fm, reg)

	req := httptest.NewRequest(http.MethodGet, "/config/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from /config/status, got %d: %s", w.Code, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	if body["current_config_id"] != "cfg-123" {
		t.Errorf("expected current_config_id=cfg-123, got %#v", body["current_config_id"])
	}
}

// TestConfigApplied_UpdatesProbeRegistry validates T097: /probes/:id/config-applied updates ProbeRegistry.
func TestConfigApplied_UpdatesProbeRegistry(t *testing.T) {
	fm := newFakeManager()
	reg := monitoring.NewProbeRegistry()
	reg.RegisterProbe(&monitoring.Probe{ID: "probe-1", Name: "p1"})

	r := newTestEngine(fm, reg)

	payload := map[string]interface{}{
		"config_id":      "cfg-123",
		"config_version": 3,
		"config_source":  "etcd",
		"applied_at":     time.Now().UTC(),
	}
	data, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/probes/probe-1/config-applied", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from /probes/:id/config-applied, got %d: %s", w.Code, w.Body.String())
	}

	probe, ok := reg.GetProbe("probe-1")
	if !ok {
		t.Fatalf("probe not found after config-applied")
	}
	if probe.ConfigID != "cfg-123" || probe.ConfigVersion != 3 || probe.ConfigSource != "etcd" {
		t.Errorf("probe config metadata not updated as expected: %#v", probe)
	}
	if probe.ConfigApplied.IsZero() {
		t.Errorf("expected ConfigApplied timestamp to be set")
	}
}

// TestConfigAcceptance_VersionPriorityContract documents the Phase 5.1 contract expectations.
// Detailed behavior is implemented and should be unit-tested within internal/config if needed.
func TestConfigAcceptance_VersionPriorityContract(t *testing.T) {
	// Contract reminder (no assertions here):
	// - Higher Version wins.
	// - On equal Version, provider priority etcd(3) > consul(2) > file(1) applies.
	// If this behavior changes, add focused tests in internal/config.
}