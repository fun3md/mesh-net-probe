package config

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ConflictManager handles configuration conflicts and provides error reporting
type ConflictManager struct {
	configMutex    sync.RWMutex
	activeConflicts map[string]*ConflictState
	errorHistory    *ErrorHistory
	resolutionRules map[string]*ResolutionRule
	ctx            context.Context
	cancel         context.CancelFunc
}

// ConflictState represents the state of a configuration conflict
type ConflictState struct {
	ConflictID     string                 `json:"conflict_id"`      // Unique conflict identifier
	ConfigKey      string                 `json:"config_key"`       // Configuration key that has conflict
	ConflictType   string                 `json:"conflict_type"`    // "version", "simultaneous", "validation"
	Status         string                 `json:"status"`           // "active", "resolved", "escalated"
	DetectedAt     time.Time              `json:"detected_at"`      // When conflict was detected
	ResolutionAt   *time.Time             `json:"resolution_at"`    // When it was resolved
	Participants   []*ConflictParticipant `json:"participants"`     // Entities involved in conflict
	ProposedValues map[string]interface{} `json:"proposed_values"`  // Proposed configuration values
	Resolution     *ConfigResolution      `json:"resolution"`       // How the conflict was resolved
}

// ConflictParticipant represents an entity involved in a conflict
type ConflictParticipant struct {
	ID        string    `json:"id"`        // Participant identifier
	Source    string    `json:"source"`    // Source (etcd, consul, probe, etc.)
	Value     string    `json:"value"`     // Proposed value (serialized)
	Timestamp time.Time `json:"timestamp"` // When the value was proposed
	Priority  int       `json:"priority"`  // Resolution priority
}

// ConfigResolution describes how a conflict was resolved
type ConfigResolution struct {
	Method     string                 `json:"method"`      // "manual", "auto_last_writer", "auto_priority", "merge"
	ResolvedBy string                 `json:"resolved_by"` // Who resolved it
	Value      map[string]interface{} `json:"value"`       // Final resolved value
	Reason     string                 `json:"reason"`      // Why this resolution was chosen
}

// ErrorHistory contains historical configuration errors
type ErrorHistory struct {
	errors       []ConfigError
	errorMutex   sync.RWMutex
	maxHistory   int
	lastCleanup  time.Time
}

// ConfigError represents a configuration error
type ConfigError struct {
	ErrorID      string                 `json:"error_id"`       // Unique error identifier
	Timestamp    time.Time              `json:"timestamp"`      // When error occurred
	Severity     string                 `json:"severity"`       // "low", "medium", "high", "critical"
	Category     string                 `json:"category"`       // "validation", "network", "conflict", "timeout"
	Message      string                 `json:"message"`        // Error message
	Context      map[string]interface{} `json:"context"`        // Additional context
	Source       string                 `json:"source"`         // Error source
	Resolved     bool                   `json:"resolved"`       // Whether error is resolved
	Resolution   string                 `json:"resolution"`     // How it was resolved
}

// ResolutionRule defines how conflicts should be automatically resolved
type ResolutionRule struct {
	RuleID       string    `json:"rule_id"`       // Rule identifier
	Name         string    `json:"name"`          // Rule name
	Priority     int       `json:"priority"`      // Rule priority (higher wins)
	MatchPattern string    `json:"match_pattern"` // Configuration key pattern to match
	Method       string    `json:"method"`        // Resolution method
	Enabled      bool      `json:"enabled"`       // Whether rule is enabled
	CreatedAt    time.Time `json:"created_at"`    // When rule was created
	UpdatedAt    time.Time `json:"updated_at"`    // When rule was last updated
}

// ConflictOption allows configuration of the conflict resolution system
type ConflictOption func(*ConflictManager)

// WithResolutionRule adds a resolution rule
func WithResolutionRule(rule *ResolutionRule) ConflictOption {
	return func(cm *ConflictManager) {
		cm.resolutionRules[rule.RuleID] = rule
	}
}

// WithErrorHistorySize sets the maximum error history size
func WithErrorHistorySize(size int) ConflictOption {
	return func(cm *ConflictManager) {
		cm.errorHistory.maxHistory = size
	}
}

// NewConflictManager creates a new conflict resolution system
func NewConflictManager(options ...ConflictOption) *ConflictManager {
	cm := &ConflictManager{
		activeConflicts: make(map[string]*ConflictState),
		errorHistory: &ErrorHistory{
			errors:     make([]ConfigError, 0),
			maxHistory: 1000,
		},
		resolutionRules: make(map[string]*ResolutionRule),
	}

	for _, option := range options {
		option(cm)
	}

	// Add default resolution rules
	cm.addDefaultResolutionRules()

	return cm
}

// Start starts the conflict resolution system
func (cm *ConflictManager) Start(ctx context.Context) error {
	cm.ctx, cm.cancel = context.WithCancel(ctx)

	// Start background processes
	go cm.errorHistoryCleanup()
	go cm.conflictMonitor()

	return nil
}

// Stop stops the conflict resolution system
func (cm *ConflictManager) Stop(ctx context.Context) error {
	if cm.cancel != nil {
		cm.cancel()
	}
	return nil
}

// DetectConflict detects configuration conflicts
func (cm *ConflictManager) DetectConflict(configKey string, participants []*ConflictParticipant) (*ConflictState, error) {
	cm.configMutex.Lock()
	defer cm.configMutex.Unlock()

	// Check if there's already an active conflict for this key
	if existing, exists := cm.activeConflicts[configKey]; exists && existing.Status == "active" {
		// Add new participants to existing conflict
		existing.Participants = append(existing.Participants, participants...)
		return existing, nil
	}

	// Create new conflict
	conflict := &ConflictState{
		ConflictID:     generateConflictID(),
		ConfigKey:      configKey,
		ConflictType:   cm.determineConflictType(participants),
		Status:         "active",
		DetectedAt:     time.Now(),
		Participants:   participants,
		ProposedValues: cm.extractProposedValues(participants),
	}

	cm.activeConflicts[configKey] = conflict

	// Log the conflict
	cm.logError(&ConfigError{
		ErrorID:   generateErrorID(),
		Timestamp: time.Now(),
		Severity:  cm.getConflictSeverity(conflict),
		Category:  "conflict",
		Message:   fmt.Sprintf("Configuration conflict detected for key '%s' with %d participants", configKey, len(participants)),
		Context: map[string]interface{}{
			"conflict_id":  conflict.ConflictID,
			"config_key":   configKey,
			"participants": len(participants),
		},
	})

	return conflict, nil
}

// ResolveConflict attempts to resolve a configuration conflict
func (cm *ConflictManager) ResolveConflict(conflictID string, method string, resolvedBy string) (*ConfigResolution, error) {
	cm.configMutex.Lock()
	defer cm.configMutex.Unlock()

	// Find the conflict
	var conflict *ConflictState
	for _, conf := range cm.activeConflicts {
		if conf.ConflictID == conflictID {
			conflict = conf
			break
		}
	}

	if conflict == nil {
		return nil, fmt.Errorf("conflict not found: %s", conflictID)
	}

	if conflict.Status != "active" {
		return nil, fmt.Errorf("conflict is not active: %s", conflictID)
	}

	// Apply resolution
	resolution := cm.applyResolution(conflict, method, resolvedBy)
	conflict.Resolution = resolution
	conflict.Status = "resolved"
	conflict.ResolutionAt = &[]time.Time{time.Now()}[0]

	// Remove from active conflicts
	delete(cm.activeConflicts, conflict.ConfigKey)

	// Log resolution
	cm.logError(&ConfigError{
		ErrorID:   generateErrorID(),
		Timestamp: time.Now(),
		Severity:  "low",
		Category:  "resolution",
		Message:   fmt.Sprintf("Configuration conflict resolved: %s", conflictID),
		Context: map[string]interface{}{
			"conflict_id": conflictID,
			"method":      method,
			"resolved_by": resolvedBy,
		},
		Resolved:   true,
		Resolution: fmt.Sprintf("Resolved using method: %s", method),
	})

	return resolution, nil
}

// GetConflict returns a configuration conflict by ID
func (cm *ConflictManager) GetConflict(conflictID string) (*ConflictState, error) {
	cm.configMutex.RLock()
	defer cm.configMutex.RUnlock()

	for _, conflict := range cm.activeConflicts {
		if conflict.ConflictID == conflictID {
			return conflict, nil
		}
	}

	return nil, fmt.Errorf("conflict not found: %s", conflictID)
}

// GetActiveConflicts returns all active configuration conflicts
func (cm *ConflictManager) GetActiveConflicts() map[string]*ConflictState {
	cm.configMutex.RLock()
	defer cm.configMutex.RUnlock()

	result := make(map[string]*ConflictState)
	for k, v := range cm.activeConflicts {
		result[k] = v
	}
	return result
}

// LogError logs a configuration error
func (cm *ConflictManager) LogError(error *ConfigError) error {
	return cm.logError(error)
}

// GetErrorHistory returns the error history
func (cm *ConflictManager) GetErrorHistory(limit int) []ConfigError {
	return cm.errorHistory.getErrors(limit)
}

// GetConflictsByStatus returns conflicts filtered by status
func (cm *ConflictManager) GetConflictsByStatus(status string) []*ConflictState {
	cm.configMutex.RLock()
	defer cm.configMutex.RUnlock()

	var conflicts []*ConflictState
	for _, conflict := range cm.activeConflicts {
		if conflict.Status == status {
			conflicts = append(conflicts, conflict)
		}
	}
	return conflicts
}

// determineConflictType determines the type of conflict based on participants
func (cm *ConflictManager) determineConflictType(participants []*ConflictParticipant) string {
	if len(participants) > 2 {
		return "simultaneous"
	}

	// Check if versions are different
	if len(participants) >= 2 && participants[0].Timestamp != participants[1].Timestamp {
		return "version"
	}

	return "validation"
}

// extractProposedValues extracts the proposed values from participants
func (cm *ConflictManager) extractProposedValues(participants []*ConflictParticipant) map[string]interface{} {
	values := make(map[string]interface{})
	for _, participant := range participants {
		values[participant.ID] = participant.Value
	}
	return values
}

// applyResolution applies a resolution method to a conflict
func (cm *ConflictManager) applyResolution(conflict *ConflictState, method, resolvedBy string) *ConfigResolution {
	var resolution *ConfigResolution

	switch method {
	case "manual":
		resolution = &ConfigResolution{
			Method:     method,
			ResolvedBy: resolvedBy,
			Reason:     "Manual resolution",
		}
	case "auto_last_writer":
		if len(conflict.Participants) > 0 {
			lastWriter := conflict.Participants[0]
			for _, p := range conflict.Participants {
				if p.Timestamp.After(lastWriter.Timestamp) {
					lastWriter = p
				}
			}
			resolution = &ConfigResolution{
				Method:     method,
				ResolvedBy: "system",
				Value:      map[string]interface{}{"value": lastWriter.Value},
				Reason:     "Last writer wins",
			}
		}
	case "auto_priority":
		if len(conflict.Participants) > 0 {
			highest := conflict.Participants[0]
			for _, p := range conflict.Participants {
				if p.Priority > highest.Priority {
					highest = p
				}
			}
			resolution = &ConfigResolution{
				Method:     method,
				ResolvedBy: "system",
				Value:      map[string]interface{}{"value": highest.Value},
				Reason:     "Highest priority wins",
			}
		}
	case "merge":
		// Simple merge strategy - take the most recent value for each key
		if len(conflict.Participants) > 0 {
			resolvedValue := conflict.Participants[0].Value
			resolution = &ConfigResolution{
				Method:     method,
				ResolvedBy: "system",
				Value:      map[string]interface{}{"value": resolvedValue},
				Reason:     "Merge strategy applied",
			}
		}
	default:
		resolution = &ConfigResolution{
			Method:     "error",
			ResolvedBy: "system",
			Reason:     fmt.Sprintf("Unknown resolution method: %s", method),
		}
	}

	return resolution
}

// getConflictSeverity determines the severity of a conflict
func (cm *ConflictManager) getConflictSeverity(conflict *ConflictState) string {
	switch conflict.ConflictType {
	case "simultaneous":
		return "high"
	case "version":
		return "medium"
	case "validation":
		return "low"
	default:
		return "medium"
	}
}

// logError logs an error to the error history
func (cm *ConflictManager) logError(error *ConfigError) error {
	cm.errorHistory.addError(error)
	return nil
}

// addDefaultResolutionRules adds default resolution rules
func (cm *ConflictManager) addDefaultResolutionRules() {
	cm.resolutionRules["last_writer_wins"] = &ResolutionRule{
		RuleID:       "last_writer_wins",
		Name:         "Last Writer Wins",
		Priority:     5,
		MatchPattern: "*",
		Method:       "auto_last_writer",
		Enabled:      true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	cm.resolutionRules["priority_based"] = &ResolutionRule{
		RuleID:       "priority_based",
		Name:         "Priority Based",
		Priority:     10,
		MatchPattern: "*",
		Method:       "auto_priority",
		Enabled:      true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// errorHistoryCleanup performs periodic cleanup of error history
func (cm *ConflictManager) errorHistoryCleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.errorHistory.cleanup()
		}
	}
}

// conflictMonitor monitors active conflicts for timeout
func (cm *ConflictManager) conflictMonitor() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.checkConflictTimeouts()
		}
	}
}

// checkConflictTimeouts checks for conflicts that have timed out
func (cm *ConflictManager) checkConflictTimeouts() {
	cm.configMutex.Lock()
	defer cm.configMutex.Unlock()

	timeout := 5 * time.Minute // 5-minute timeout
	escalated := 0

	for key, conflict := range cm.activeConflicts {
		if conflict.Status == "active" && time.Since(conflict.DetectedAt) > timeout {
			// Escalate the conflict
			conflict.Status = "escalated"
			escalated++

			// Log escalation
			cm.logError(&ConfigError{
				ErrorID:   generateErrorID(),
				Timestamp: time.Now(),
				Severity:  "high",
				Category:  "escalation",
				Message:   fmt.Sprintf("Configuration conflict escalated: %s", conflict.ConflictID),
				Context: map[string]interface{}{
					"conflict_id":   conflict.ConflictID,
					"config_key":    key,
					"duration":      time.Since(conflict.DetectedAt).String(),
					"participants":  len(conflict.Participants),
				},
			})
		}
	}

	if escalated > 0 {
		fmt.Printf("ESCALATION: %d configuration conflicts escalated due to timeout\n", escalated)
	}
}

// generateConflictID generates a unique conflict ID
func generateConflictID() string {
	return fmt.Sprintf("conf_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// generateErrorID generates a unique error ID
func generateErrorID() string {
	return fmt.Sprintf("err_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// addError adds an error to the history
func (h *ErrorHistory) addError(error *ConfigError) {
	h.errorMutex.Lock()
	defer h.errorMutex.Unlock()

	h.errors = append(h.errors, *error)

	// Cleanup if we exceed the maximum history
	if len(h.errors) > h.maxHistory {
		h.errors = h.errors[1:]
	}
}

// getErrors returns the error history
func (h *ErrorHistory) getErrors(limit int) []ConfigError {
	h.errorMutex.RLock()
	defer h.errorMutex.RUnlock()

	if limit <= 0 || limit > len(h.errors) {
		limit = len(h.errors)
	}

	// Return the most recent errors
	result := make([]ConfigError, limit)
	copy(result, h.errors[len(h.errors)-limit:])
	return result
}

// cleanup performs cleanup of old errors
func (h *ErrorHistory) cleanup() {
	h.errorMutex.Lock()
	defer h.errorMutex.Unlock()

	// Remove errors older than 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	cleaned := 0

	for i := 0; i < len(h.errors); {
		if h.errors[i].Timestamp.Before(cutoff) {
			cleaned++
			h.errors = append(h.errors[:i], h.errors[i+1:]...)
		} else {
			i++
		}
	}

	h.lastCleanup = time.Now()
	
	if cleaned > 0 {
		fmt.Printf("CONFIG: Cleaned up %d old configuration errors\n", cleaned)
	}
}