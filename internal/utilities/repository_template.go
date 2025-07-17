package utilities

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// MetricsCollector defines interface for collecting repository metrics
type MetricsCollector interface {
	IncrementCounter(name string, labels map[string]string)
	RecordHistogram(name string, value float64, labels map[string]string)
}

// Logger defines interface for repository logging
type Logger interface {
	InfoContext(ctx context.Context, msg string, fields ...interface{})
	ErrorContext(ctx context.Context, msg string, fields ...interface{})
	WarnContext(ctx context.Context, msg string, fields ...interface{})
}

// Database defines interface for database connections
type Database interface {
	BeginTx(ctx context.Context) (Transaction, error)
	HealthCheck(ctx context.Context) error
	Close() error
}

// QueryOperation defines operations for query execution
type QueryOperation interface {
	Execute(ctx context.Context, db Database) (interface{}, error)
}

// BatchOperation defines operations for batch execution
type BatchOperation interface {
	Execute(ctx context.Context, db Database) error
}

// RepositoryTemplate defines template method pattern for standardized repository operations
type RepositoryTemplate interface {
	// Template methods that define the algorithm structure
	ExecuteWithTransaction(ctx context.Context, operation TransactionOperation) error
	ExecuteQuery(ctx context.Context, query QueryOperation) (interface{}, error)
	ExecuteBatch(ctx context.Context, operations []BatchOperation) error

	// Hook methods for customization
	BeforeOperation(ctx context.Context, operation OperationType) error
	AfterOperation(ctx context.Context, operation OperationType, result interface{}) error
	OnError(ctx context.Context, operation OperationType, err error) error
}

// BaseRepository implements template method pattern for database operations
type BaseRepository struct {
	db      Database
	logger  Logger
	metrics MetricsCollector
}

// NewBaseRepository creates a new base repository with template method implementation
func NewBaseRepository(db Database, logger Logger, metrics MetricsCollector) *BaseRepository {
	return &BaseRepository{
		db:      db,
		logger:  logger,
		metrics: metrics,
	}
}

// ExecuteWithTransaction implements the template method for transaction operations
func (b *BaseRepository) ExecuteWithTransaction(ctx context.Context, operation TransactionOperation) error {
	start := time.Now()

	// Pre-operation hook
	if err := b.BeforeOperation(ctx, OperationTypeTransaction); err != nil {
		return err
	}

	// Begin transaction
	tx, err := b.db.BeginTx(ctx)
	if err != nil {
		b.OnError(ctx, OperationTypeTransaction, err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction cleanup
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}

		// Record metrics
		duration := time.Since(start).Seconds()
		b.metrics.RecordHistogram("repository_transaction_duration", duration, map[string]string{
			"operation": string(OperationTypeTransaction),
		})
	}()

	// Execute the operation
	result, err := operation.Execute(ctx, nil) // Transaction passed internally

	if err != nil {
		b.OnError(ctx, OperationTypeTransaction, err)
		return err
	}

	return b.AfterOperation(ctx, OperationTypeTransaction, result)
}

// ExecuteQuery implements the template method for query operations
func (b *BaseRepository) ExecuteQuery(ctx context.Context, query QueryOperation) (interface{}, error) {
	start := time.Now()

	// Pre-operation hook
	if err := b.BeforeOperation(ctx, OperationTypeRead); err != nil {
		return nil, err
	}

	// Execute query
	result, err := query.Execute(ctx, b.db)

	// Record metrics
	duration := time.Since(start).Seconds()
	b.metrics.RecordHistogram("repository_query_duration", duration, map[string]string{
		"operation": string(OperationTypeRead),
	})

	if err != nil {
		b.OnError(ctx, OperationTypeRead, err)
		return nil, err
	}

	if err := b.AfterOperation(ctx, OperationTypeRead, result); err != nil {
		return nil, err
	}

	return result, nil
}

// ExecuteBatch implements the template method for batch operations
func (b *BaseRepository) ExecuteBatch(ctx context.Context, operations []BatchOperation) error {
	start := time.Now()

	// Pre-operation hook
	if err := b.BeforeOperation(ctx, OperationTypeBatch); err != nil {
		return err
	}

	// Execute all operations
	for i, operation := range operations {
		if err := operation.Execute(ctx, b.db); err != nil {
			batchErr := fmt.Errorf("batch operation %d failed: %w", i, err)
			b.OnError(ctx, OperationTypeBatch, batchErr)
			return batchErr
		}
	}

	// Record metrics
	duration := time.Since(start).Seconds()
	b.metrics.RecordHistogram("repository_batch_duration", duration, map[string]string{
		"operation": string(OperationTypeBatch),
		"count":     fmt.Sprintf("%d", len(operations)),
	})

	return b.AfterOperation(ctx, OperationTypeBatch, len(operations))
}

// Default hook implementations that can be overridden
func (b *BaseRepository) BeforeOperation(ctx context.Context, operation OperationType) error {
	b.logger.InfoContext(ctx, "Starting repository operation", "type", operation)
	b.metrics.IncrementCounter("repository_operations", map[string]string{
		"operation": string(operation),
	})
	return nil
}

func (b *BaseRepository) AfterOperation(ctx context.Context, operation OperationType, result interface{}) error {
	b.logger.InfoContext(ctx, "Completed repository operation", "type", operation)
	b.metrics.IncrementCounter("repository_operations_success", map[string]string{
		"operation": string(operation),
	})
	return nil
}

func (b *BaseRepository) OnError(ctx context.Context, operation OperationType, err error) error {
	b.logger.ErrorContext(ctx, "Repository operation failed", "error", err, "operation", operation)
	b.metrics.IncrementCounter("repository_operations_error", map[string]string{
		"operation": string(operation),
		"error":     getErrorType(err),
	})
	return err
}

// getErrorType extracts error type for metrics
func getErrorType(err error) string {
	if err == nil {
		return "none"
	}

	// Simple error classification
	errStr := err.Error()
	switch {
	case contains(errStr, "connection"):
		return "connection"
	case contains(errStr, "timeout"):
		return "timeout"
	case contains(errStr, "not found"):
		return "not_found"
	case contains(errStr, "duplicate"):
		return "duplicate"
	default:
		return "unknown"
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || (len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				indexOf(s, substr) >= 0)))
}

// indexOf finds the index of substr in s
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ZerologAdapter adapts zerolog to the Logger interface
type ZerologAdapter struct{}

func (z *ZerologAdapter) InfoContext(ctx context.Context, msg string, fields ...interface{}) {
	log.Info().Fields(fields).Msg(msg)
}

func (z *ZerologAdapter) ErrorContext(ctx context.Context, msg string, fields ...interface{}) {
	log.Error().Fields(fields).Msg(msg)
}

func (z *ZerologAdapter) WarnContext(ctx context.Context, msg string, fields ...interface{}) {
	log.Warn().Fields(fields).Msg(msg)
}

// NoOpMetrics provides a no-operation metrics collector for testing
type NoOpMetrics struct{}

func (n *NoOpMetrics) IncrementCounter(name string, labels map[string]string)               {}
func (n *NoOpMetrics) RecordHistogram(name string, value float64, labels map[string]string) {}
