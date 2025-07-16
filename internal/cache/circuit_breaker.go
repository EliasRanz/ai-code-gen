package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CircuitState represents the current state of the circuit breaker
type CircuitState string

const (
	StateClosed   CircuitState = "closed"    // Normal operation
	StateOpen     CircuitState = "open"      // Circuit is open, failing fast
	StateHalfOpen CircuitState = "half_open" // Testing recovery
)

// CircuitBreaker implements the Circuit Breaker Pattern for cache operations
type CircuitBreaker interface {
	Execute(ctx context.Context, operation func() (interface{}, error)) (interface{}, error)
	State() CircuitState
	Reset() error
	GetMetrics() CircuitMetrics
}

// CircuitMetrics holds circuit breaker metrics
type CircuitMetrics struct {
	Requests         int64        `json:"requests"`
	Successes        int64        `json:"successes"`
	Failures         int64        `json:"failures"`
	ConsecutiveFails int64        `json:"consecutive_fails"`
	LastFailure      time.Time    `json:"last_failure"`
	State            CircuitState `json:"state"`
	StateChanged     time.Time    `json:"state_changed"`
}

// CircuitBreakerConfig holds configuration for the circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold       int           `json:"failure_threshold"`
	RequestVolumeThreshold int           `json:"request_volume_threshold"`
	RecoveryTimeout        time.Duration `json:"recovery_timeout"`
	MaxConcurrentRequests  int           `json:"max_concurrent_requests"`
}

// DefaultCircuitBreakerConfig returns sensible default configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold:       5,
		RequestVolumeThreshold: 10,
		RecoveryTimeout:        30 * time.Second,
		MaxConcurrentRequests:  100,
	}
}

// cacheCircuitBreaker implements CircuitBreaker for cache operations
type cacheCircuitBreaker struct {
	mutex              sync.RWMutex
	state              CircuitState
	config             CircuitBreakerConfig
	metrics            CircuitMetrics
	stateChanged       time.Time
	concurrentRequests int64
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) CircuitBreaker {
	return &cacheCircuitBreaker{
		state:        StateClosed,
		config:       config,
		stateChanged: time.Now(),
		metrics: CircuitMetrics{
			State:        StateClosed,
			StateChanged: time.Now(),
		},
	}
}

// Execute runs the operation through the circuit breaker
func (cb *cacheCircuitBreaker) Execute(ctx context.Context, operation func() (interface{}, error)) (interface{}, error) {
	// Check if we should reject the request
	if !cb.canExecute() {
		return nil, errors.New("circuit breaker is open")
	}

	// Increment concurrent requests
	cb.incrementConcurrent()
	defer cb.decrementConcurrent()

	// Execute the operation
	result, err := operation()

	// Record the result
	cb.recordResult(err == nil)

	return result, err
}

// State returns the current circuit breaker state
func (cb *cacheCircuitBreaker) State() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// Reset manually resets the circuit breaker to closed state
func (cb *cacheCircuitBreaker) Reset() error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.state = StateClosed
	cb.metrics.State = StateClosed
	cb.metrics.ConsecutiveFails = 0
	cb.stateChanged = time.Now()
	cb.metrics.StateChanged = time.Now()

	return nil
}

// GetMetrics returns current circuit breaker metrics
func (cb *cacheCircuitBreaker) GetMetrics() CircuitMetrics {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.metrics
}

// canExecute determines if a request should be allowed
func (cb *cacheCircuitBreaker) canExecute() bool {
	cb.mutex.RLock()
	state := cb.state
	stateChanged := cb.stateChanged
	concurrentRequests := cb.concurrentRequests
	cb.mutex.RUnlock()

	switch state {
	case StateClosed:
		return concurrentRequests < int64(cb.config.MaxConcurrentRequests)
	case StateOpen:
		// Check if we should transition to half-open
		if time.Since(stateChanged) > cb.config.RecoveryTimeout {
			cb.transitionToHalfOpen()
			return true
		}
		return false
	case StateHalfOpen:
		return concurrentRequests == 0 // Only allow one request in half-open
	default:
		return false
	}
}

// recordResult records the result of an operation and updates state if needed
func (cb *cacheCircuitBreaker) recordResult(success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.metrics.Requests++

	if success {
		cb.metrics.Successes++
		cb.metrics.ConsecutiveFails = 0

		// If we're in half-open state and succeed, transition to closed
		if cb.state == StateHalfOpen {
			cb.transitionToClosed()
		}
	} else {
		cb.metrics.Failures++
		cb.metrics.ConsecutiveFails++
		cb.metrics.LastFailure = time.Now()

		// Check if we should open the circuit
		if cb.shouldOpen() {
			cb.transitionToOpen()
		}
	}
}

// shouldOpen determines if the circuit should be opened based on failure rate
func (cb *cacheCircuitBreaker) shouldOpen() bool {
	// Open if consecutive failures exceed threshold regardless of request volume
	// This is more appropriate for a circuit breaker protecting against consecutive failures
	return cb.metrics.ConsecutiveFails >= int64(cb.config.FailureThreshold)
}

// State transition methods
func (cb *cacheCircuitBreaker) transitionToOpen() {
	cb.state = StateOpen
	cb.metrics.State = StateOpen
	cb.stateChanged = time.Now()
	cb.metrics.StateChanged = time.Now()
}

func (cb *cacheCircuitBreaker) transitionToHalfOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.state = StateHalfOpen
	cb.metrics.State = StateHalfOpen
	cb.stateChanged = time.Now()
	cb.metrics.StateChanged = time.Now()
}

func (cb *cacheCircuitBreaker) transitionToClosed() {
	cb.state = StateClosed
	cb.metrics.State = StateClosed
	cb.stateChanged = time.Now()
	cb.metrics.StateChanged = time.Now()
}

// Concurrent request tracking
func (cb *cacheCircuitBreaker) incrementConcurrent() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.concurrentRequests++
}

func (cb *cacheCircuitBreaker) decrementConcurrent() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.concurrentRequests--
}
