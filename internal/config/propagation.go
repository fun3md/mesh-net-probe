package config

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ChangePropagation handles configuration change propagation with SLA guarantees
type ChangePropagation struct {
	configMutex    sync.RWMutex
	activeTargets  map[string]*TargetState
	propagationCh  chan *PropagationRequest
	stats          *PropagationStats
	ctx            context.Context
	cancel         context.CancelFunc
}

// TargetState represents the propagation state for a single target
type TargetState struct {
	TargetID         string    `json:"target_id"`          // Target identifier
	Status           string    `json:"status"`             // "pending", "propagating", "success", "failed"
	LastAttempt      time.Time `json:"last_attempt"`       // Last propagation attempt
	NextRetry        time.Time `json:"next_retry"`         // Next retry time
	RetryCount       int       `json:"retry_count"`        // Number of retries
	CurrentVersion   string    `json:"current_version"`    // Current configuration version
	TargetVersion    string    `json:"target_version"`     // Target configuration version
	PropagationDelay time.Duration `json:"propagation_delay"` // Actual propagation delay
}

// PropagationRequest represents a request to propagate configuration changes
type PropagationRequest struct {
	TargetID   string                 `json:"target_id"`
	ConfigID   string                 `json:"config_id"`
	Version    string                 `json:"version"`
	Source     string                 `json:"source"`
	Data       map[string]interface{} `json:"data"`
	Priority   int                    `json:"priority"`
	Timeout    time.Duration          `json:"timeout"`
}

// PropagationStats contains propagation statistics
type PropagationStats struct {
	TotalRequests    int64                   `json:"total_requests"`
	Successful       int64                   `json:"successful"`
	Failed           int64                   `json:"failed"`
	AverageDelay     time.Duration           `json:"average_delay"`
	TargetsByStatus  map[string]int          `json:"targets_by_status"`
	LastPropagation  time.Time               `json:"last_propagation"`
	CurrentSLA       time.Duration           `json:"current_sla"` // Target SLA (60 seconds)
}

// PropagationOption allows configuration of the propagation system
type PropagationOption func(*ChangePropagation)

// WithPropagationTarget adds a target to propagate to
func WithPropagationTarget(targetID string, initialVersion string) PropagationOption {
	return func(p *ChangePropagation) {
		p.activeTargets[targetID] = &TargetState{
			TargetID:       targetID,
			Status:         "pending",
			CurrentVersion: initialVersion,
			TargetVersion:  initialVersion,
		}
	}
}

// WithSLATarget sets the target SLA for configuration propagation
func WithSLATarget(sla time.Duration) PropagationOption {
	return func(p *ChangePropagation) {
		p.stats.CurrentSLA = sla
	}
}

// NewChangePropagation creates a new change propagation system
func NewChangePropagation(options ...PropagationOption) *ChangePropagation {
	p := &ChangePropagation{
		activeTargets: make(map[string]*TargetState),
		propagationCh: make(chan *PropagationRequest, 1000),
		stats: &PropagationStats{
			CurrentSLA: 60 * time.Second, // Default 60-second SLA
			TargetsByStatus: make(map[string]int),
		},
	}

	for _, option := range options {
		option(p)
	}

	return p
}

// Start starts the change propagation system
func (p *ChangePropagation) Start(ctx context.Context) error {
	p.ctx, p.cancel = context.WithCancel(ctx)

	// Start the propagation worker
	go p.propagationWorker()

	// Start the SLA monitoring
	go p.slaMonitor()

	return nil
}

// Stop stops the change propagation system
func (p *ChangePropagation) Stop(ctx context.Context) error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

// PropagateChange initiates configuration change propagation
func (p *ChangePropagation) PropagateChange(req *PropagationRequest) error {
	// Validate the request
	if err := p.validatePropagationRequest(req); err != nil {
		return fmt.Errorf("invalid propagation request: %w", err)
	}

	// Check if target exists
	if _, exists := p.activeTargets[req.TargetID]; !exists {
		return fmt.Errorf("unknown target: %s", req.TargetID)
	}

	// Add to propagation queue
	select {
	case p.propagationCh <- req:
		p.stats.TotalRequests++
		return nil
	default:
		return fmt.Errorf("propagation queue full")
	}
}

// GetPropagationStatus returns the current propagation status
func (p *ChangePropagation) GetPropagationStatus(targetID string) (*TargetState, error) {
	p.configMutex.RLock()
	defer p.configMutex.RUnlock()

	if target, exists := p.activeTargets[targetID]; exists {
		return target, nil
	}

	return nil, fmt.Errorf("unknown target: %s", targetID)
}

// GetActiveTargets returns all active targets
func (p *ChangePropagation) GetActiveTargets() map[string]*TargetState {
	p.configMutex.RLock()
	defer p.configMutex.RUnlock()

	result := make(map[string]*TargetState)
	for k, v := range p.activeTargets {
		result[k] = v
	}
	return result
}

// GetStats returns propagation statistics
func (p *ChangePropagation) GetStats() *PropagationStats {
	p.configMutex.RLock()
	defer p.configMutex.RUnlock()

	// Update current statistics
	p.stats.TargetsByStatus = make(map[string]int)
	for _, target := range p.activeTargets {
		p.stats.TargetsByStatus[target.Status]++
	}

	return p.stats
}

// AddTarget adds a new propagation target
func (p *ChangePropagation) AddTarget(targetID string, version string) error {
	p.configMutex.Lock()
	defer p.configMutex.Unlock()

	if _, exists := p.activeTargets[targetID]; exists {
		return fmt.Errorf("target already exists: %s", targetID)
	}

	p.activeTargets[targetID] = &TargetState{
		TargetID:       targetID,
		Status:         "pending",
		CurrentVersion: version,
		TargetVersion:  version,
	}

	return nil
}

// RemoveTarget removes a propagation target
func (p *ChangePropagation) RemoveTarget(targetID string) error {
	p.configMutex.Lock()
	defer p.configMutex.Unlock()

	if _, exists := p.activeTargets[targetID]; !exists {
		return fmt.Errorf("unknown target: %s", targetID)
	}

	delete(p.activeTargets, targetID)
	return nil
}

// propagationWorker processes propagation requests
func (p *ChangePropagation) propagationWorker() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case req := <-p.propagationCh:
			p.processPropagationRequest(req)
		}
	}
}

// processPropagationRequest processes a single propagation request
func (p *ChangePropagation) processPropagationRequest(req *PropagationRequest) {
	startTime := time.Now()

	// Update target state
	p.configMutex.Lock()
	target, exists := p.activeTargets[req.TargetID]
	if !exists {
		p.configMutex.Unlock()
		p.stats.Failed++
		return
	}

	target.Status = "propagating"
	target.LastAttempt = startTime
	p.configMutex.Unlock()

	// Propagate the change (simulated)
	err := p.doPropagation(req)

	// Update target state based on result
	p.configMutex.Lock()
	if err != nil {
		target.Status = "failed"
		target.RetryCount++
		target.NextRetry = startTime.Add(p.calculateRetryDelay(target.RetryCount))
		p.stats.Failed++
	} else {
		target.Status = "success"
		target.CurrentVersion = req.Version
		target.TargetVersion = req.Version
		target.PropagationDelay = time.Since(startTime)
		p.stats.Successful++
		p.stats.LastPropagation = startTime
	}
	p.configMutex.Unlock()
}

// doPropagation performs the actual configuration propagation
func (p *ChangePropagation) doPropagation(req *PropagationRequest) error {
	// Simulate network latency and processing time
	latency := time.Millisecond * time.Duration(10+req.Priority*5)
	select {
	case <-time.After(latency):
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

// validatePropagationRequest validates a propagation request
func (p *ChangePropagation) validatePropagationRequest(req *PropagationRequest) error {
	if req.TargetID == "" {
		return fmt.Errorf("target ID is required")
	}
	if req.ConfigID == "" {
		return fmt.Errorf("config ID is required")
	}
	if req.Version == "" {
		return fmt.Errorf("version is required")
	}
	if req.Timeout <= 0 {
		req.Timeout = 30 * time.Second // Default timeout
	}
	if req.Priority < 1 || req.Priority > 10 {
		return fmt.Errorf("priority must be between 1 and 10")
	}

	return nil
}

// calculateRetryDelay calculates the retry delay based on retry count
func (p *ChangePropagation) calculateRetryDelay(retryCount int) time.Duration {
	// Exponential backoff with jitter
	baseDelay := 5 * time.Second
	maxDelay := 300 * time.Second // 5 minutes max

	delay := baseDelay * time.Duration(1<<uint(retryCount))
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// slaMonitor monitors SLA compliance
func (p *ChangePropagation) slaMonitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.checkSLACompliance()
		}
	}
}

// checkSLACompliance checks if all active propagation requests meet the SLA
func (p *ChangePropagation) checkSLACompliance() {
	p.configMutex.RLock()
	overdueTargets := make([]string, 0)
	
	for targetID, target := range p.activeTargets {
		if target.Status == "propagating" {
			elapsed := time.Since(target.LastAttempt)
			if elapsed > p.stats.CurrentSLA {
				overdueTargets = append(overdueTargets, targetID)
			}
		}
	}
	p.configMutex.RUnlock()

	// Log SLA violations
	if len(overdueTargets) > 0 {
		// In a real implementation, this would log to a monitoring system
		fmt.Printf("SLA VIOLATION: %d targets exceeded 60-second SLA: %v\n", len(overdueTargets), overdueTargets)
	}
}

// GetSLAStatus returns current SLA compliance status
func (p *ChangePropagation) GetSLAStatus() (*SLAStatus, error) {
	p.configMutex.RLock()
	defer p.configMutex.RUnlock()

	sla := &SLAStatus{
		TargetSLA:      p.stats.CurrentSLA,
		Compliant:      0,
		NonCompliant:   0,
		Pending:        0,
		Failed:         0,
		AverageDelay:   0,
	}

	var totalDelay time.Duration
	var delayCount int

	for _, target := range p.activeTargets {
		switch target.Status {
		case "pending":
			sla.Pending++
		case "propagating":
			elapsed := time.Since(target.LastAttempt)
			if elapsed <= p.stats.CurrentSLA {
				sla.Compliant++
			} else {
				sla.NonCompliant++
			}
		case "success":
			if target.PropagationDelay <= p.stats.CurrentSLA {
				sla.Compliant++
			} else {
				sla.NonCompliant++
			}
			totalDelay += target.PropagationDelay
			delayCount++
		case "failed":
			sla.Failed++
		}
	}

	if delayCount > 0 {
		sla.AverageDelay = totalDelay / time.Duration(delayCount)
	}

	return sla, nil
}

// SLAStatus represents SLA compliance status
type SLAStatus struct {
	TargetSLA     time.Duration `json:"target_sla"`
	Compliant     int           `json:"compliant"`
	NonCompliant  int           `json:"non_compliant"`
	Pending       int           `json:"pending"`
	Failed        int           `json:"failed"`
	AverageDelay  time.Duration `json:"average_delay"`
}