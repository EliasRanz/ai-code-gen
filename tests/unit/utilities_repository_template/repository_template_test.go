package utilities_repository_template

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// Simplified mock implementations focusing on working functionality

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) InfoContext(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) ErrorContext(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) WarnContext(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

type MockMetrics struct {
	mock.Mock
	counters   map[string]int
	histograms map[string][]float64
}

func NewMockMetrics() *MockMetrics {
	return &MockMetrics{
		counters:   make(map[string]int),
		histograms: make(map[string][]float64),
	}
}

func (m *MockMetrics) IncrementCounter(name string, labels map[string]string) {
	m.Called(name, labels)
	m.counters[name]++
}

func (m *MockMetrics) RecordHistogram(name string, value float64, labels map[string]string) {
	m.Called(name, value, labels)
	m.histograms[name] = append(m.histograms[name], value)
}

type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) BeginTx(ctx context.Context) (utilities.Transaction, error) {
	args := m.Called(ctx)
	if tx := args.Get(0); tx != nil {
		return tx.(utilities.Transaction), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDatabase) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockQueryOperation struct {
	mock.Mock
}

func (m *MockQueryOperation) Execute(ctx context.Context, db utilities.Database) (interface{}, error) {
	args := m.Called(ctx, db)
	return args.Get(0), args.Error(1)
}

type MockBatchOperation struct {
	mock.Mock
}

func (m *MockBatchOperation) Execute(ctx context.Context, db utilities.Database) error {
	args := m.Called(ctx, db)
	return args.Error(0)
}

// Test BaseRepository creation
func TestNewBaseRepository(t *testing.T) {
	db := &MockDatabase{}
	logger := &MockLogger{}
	metrics := NewMockMetrics()

	repo := utilities.NewBaseRepository(db, logger, metrics)

	assert.NotNil(t, repo)
	assert.IsType(t, &utilities.BaseRepository{}, repo)
}

// Test ExecuteQuery - Success Path
func TestBaseRepository_ExecuteQuery_Success(t *testing.T) {
	db := &MockDatabase{}
	logger := &MockLogger{}
	metrics := NewMockMetrics()
	query := &MockQueryOperation{}

	repo := utilities.NewBaseRepository(db, logger, metrics)
	ctx := context.Background()
	expectedResult := map[string]interface{}{"data": "test"}

	// Setup expectations
	query.On("Execute", ctx, db).Return(expectedResult, nil)

	// Setup logging and metrics expectations
	logger.On("InfoContext", ctx, "Starting repository operation", mock.Anything).Return()
	logger.On("InfoContext", ctx, "Completed repository operation", mock.Anything).Return()
	metrics.On("IncrementCounter", "repository_operations", mock.Anything).Return()
	metrics.On("IncrementCounter", "repository_operations_success", mock.Anything).Return()
	metrics.On("RecordHistogram", "repository_query_duration", mock.AnythingOfType("float64"), mock.Anything).Return()

	// Execute
	result, err := repo.ExecuteQuery(ctx, query)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	query.AssertExpectations(t)
	logger.AssertExpectations(t)
	metrics.AssertExpectations(t)
}

// Test ExecuteQuery - Query Failure
func TestBaseRepository_ExecuteQuery_Failure(t *testing.T) {
	db := &MockDatabase{}
	logger := &MockLogger{}
	metrics := NewMockMetrics()
	query := &MockQueryOperation{}

	repo := utilities.NewBaseRepository(db, logger, metrics)
	ctx := context.Background()
	queryErr := errors.New("query execution failed")

	// Setup expectations
	query.On("Execute", ctx, db).Return(nil, queryErr)

	// Setup logging and metrics expectations
	logger.On("InfoContext", ctx, "Starting repository operation", mock.Anything).Return()
	logger.On("ErrorContext", ctx, "Repository operation failed", mock.Anything).Return()
	metrics.On("IncrementCounter", "repository_operations", mock.Anything).Return()
	metrics.On("IncrementCounter", "repository_operations_error", mock.Anything).Return()
	metrics.On("RecordHistogram", "repository_query_duration", mock.AnythingOfType("float64"), mock.Anything).Return()

	// Execute
	result, err := repo.ExecuteQuery(ctx, query)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, queryErr, err)
	query.AssertExpectations(t)
	logger.AssertExpectations(t)
	metrics.AssertExpectations(t)
}

// Test ExecuteBatch - Success Path
func TestBaseRepository_ExecuteBatch_Success(t *testing.T) {
	db := &MockDatabase{}
	logger := &MockLogger{}
	metrics := NewMockMetrics()

	// Create multiple batch operations
	op1 := &MockBatchOperation{}
	op2 := &MockBatchOperation{}
	op3 := &MockBatchOperation{}
	operations := []utilities.BatchOperation{op1, op2, op3}

	repo := utilities.NewBaseRepository(db, logger, metrics)
	ctx := context.Background()

	// Setup expectations for all operations
	op1.On("Execute", ctx, db).Return(nil)
	op2.On("Execute", ctx, db).Return(nil)
	op3.On("Execute", ctx, db).Return(nil)

	// Setup logging and metrics expectations
	logger.On("InfoContext", ctx, "Starting repository operation", mock.Anything).Return()
	logger.On("InfoContext", ctx, "Completed repository operation", mock.Anything).Return()
	metrics.On("IncrementCounter", "repository_operations", mock.Anything).Return()
	metrics.On("IncrementCounter", "repository_operations_success", mock.Anything).Return()
	metrics.On("RecordHistogram", "repository_batch_duration", mock.AnythingOfType("float64"), mock.Anything).Return()

	// Execute
	err := repo.ExecuteBatch(ctx, operations)

	// Verify
	assert.NoError(t, err)
	op1.AssertExpectations(t)
	op2.AssertExpectations(t)
	op3.AssertExpectations(t)
	logger.AssertExpectations(t)
	metrics.AssertExpectations(t)
}

// Test ExecuteBatch - Operation Failure
func TestBaseRepository_ExecuteBatch_OperationFailure(t *testing.T) {
	db := &MockDatabase{}
	logger := &MockLogger{}
	metrics := NewMockMetrics()

	// Create batch operations with one failure
	op1 := &MockBatchOperation{}
	op2 := &MockBatchOperation{}
	op3 := &MockBatchOperation{}
	operations := []utilities.BatchOperation{op1, op2, op3}

	repo := utilities.NewBaseRepository(db, logger, metrics)
	ctx := context.Background()
	operationErr := errors.New("batch operation failed")

	// Setup expectations - op2 fails, op3 never executed
	op1.On("Execute", ctx, db).Return(nil)
	op2.On("Execute", ctx, db).Return(operationErr)

	// Setup logging and metrics expectations
	logger.On("InfoContext", ctx, "Starting repository operation", mock.Anything).Return()
	logger.On("ErrorContext", ctx, "Repository operation failed", mock.Anything).Return()
	metrics.On("IncrementCounter", "repository_operations", mock.Anything).Return()
	metrics.On("IncrementCounter", "repository_operations_error", mock.Anything).Return()

	// Execute
	err := repo.ExecuteBatch(ctx, operations)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch operation 1 failed")
	op1.AssertExpectations(t)
	op2.AssertExpectations(t)
	// op3 should not have been called
	op3.AssertNotCalled(t, "Execute")
	logger.AssertExpectations(t)
	metrics.AssertExpectations(t)
}

// Test ExecuteBatch - Empty Operations
func TestBaseRepository_ExecuteBatch_EmptyOperations(t *testing.T) {
	db := &MockDatabase{}
	logger := &MockLogger{}
	metrics := NewMockMetrics()

	repo := utilities.NewBaseRepository(db, logger, metrics)
	ctx := context.Background()

	// Setup logging and metrics expectations
	logger.On("InfoContext", ctx, "Starting repository operation", mock.Anything).Return()
	logger.On("InfoContext", ctx, "Completed repository operation", mock.Anything).Return()
	metrics.On("IncrementCounter", "repository_operations", mock.Anything).Return()
	metrics.On("IncrementCounter", "repository_operations_success", mock.Anything).Return()
	metrics.On("RecordHistogram", "repository_batch_duration", mock.AnythingOfType("float64"), mock.Anything).Return()

	// Execute
	err := repo.ExecuteBatch(ctx, []utilities.BatchOperation{})

	// Verify
	assert.NoError(t, err)
	logger.AssertExpectations(t)
	metrics.AssertExpectations(t)
}

// Test Hook Methods
func TestBaseRepository_HookMethods(t *testing.T) {
	db := &MockDatabase{}
	logger := &MockLogger{}
	metrics := NewMockMetrics()

	repo := utilities.NewBaseRepository(db, logger, metrics)
	ctx := context.Background()

	t.Run("BeforeOperation", func(t *testing.T) {
		logger.On("InfoContext", ctx, "Starting repository operation", mock.Anything).Return()
		metrics.On("IncrementCounter", "repository_operations", mock.Anything).Return()

		err := repo.BeforeOperation(ctx, utilities.OperationTypeRead)
		assert.NoError(t, err)

		logger.AssertExpectations(t)
		metrics.AssertExpectations(t)
	})

	t.Run("AfterOperation", func(t *testing.T) {
		logger.On("InfoContext", ctx, "Completed repository operation", mock.Anything).Return()
		metrics.On("IncrementCounter", "repository_operations_success", mock.Anything).Return()

		err := repo.AfterOperation(ctx, utilities.OperationTypeCreate, "result")
		assert.NoError(t, err)

		logger.AssertExpectations(t)
		metrics.AssertExpectations(t)
	})

	t.Run("OnError", func(t *testing.T) {
		testErr := errors.New("test error")

		logger.On("ErrorContext", ctx, "Repository operation failed", mock.Anything).Return()
		metrics.On("IncrementCounter", "repository_operations_error", mock.Anything).Return()

		returnedErr := repo.OnError(ctx, utilities.OperationTypeUpdate, testErr)
		assert.Equal(t, testErr, returnedErr)

		logger.AssertExpectations(t)
		metrics.AssertExpectations(t)
	})
}

// Test Error Classification through OnError
func TestBaseRepository_ErrorClassification(t *testing.T) {
	tests := []struct {
		name         string
		error        error
		expectedType string
	}{
		{
			name:         "nil error",
			error:        nil,
			expectedType: "none",
		},
		{
			name:         "connection error",
			error:        errors.New("database connection failed"),
			expectedType: "connection",
		},
		{
			name:         "timeout error",
			error:        errors.New("query timeout exceeded"),
			expectedType: "timeout",
		},
		{
			name:         "not found error",
			error:        errors.New("record not found"),
			expectedType: "not_found",
		},
		{
			name:         "duplicate error",
			error:        errors.New("duplicate key violation"),
			expectedType: "duplicate",
		},
		{
			name:         "unknown error",
			error:        errors.New("something went wrong"),
			expectedType: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &MockDatabase{}
			logger := &MockLogger{}
			metrics := NewMockMetrics()

			repo := utilities.NewBaseRepository(db, logger, metrics)
			ctx := context.Background()

			logger.On("ErrorContext", ctx, "Repository operation failed", mock.Anything).Return()
			metrics.On("IncrementCounter", "repository_operations_error", mock.MatchedBy(func(labels map[string]string) bool {
				return labels["error"] == tt.expectedType
			})).Return()

			repo.OnError(ctx, utilities.OperationTypeRead, tt.error)

			metrics.AssertExpectations(t)
		})
	}
}

// Test ZerologAdapter
func TestZerologAdapter(t *testing.T) {
	adapter := &utilities.ZerologAdapter{}
	ctx := context.Background()

	// These tests mainly ensure no panics occur
	t.Run("InfoContext", func(t *testing.T) {
		assert.NotPanics(t, func() {
			adapter.InfoContext(ctx, "test message", "key", "value")
		})
	})

	t.Run("ErrorContext", func(t *testing.T) {
		assert.NotPanics(t, func() {
			adapter.ErrorContext(ctx, "error message", "error", "test")
		})
	})

	t.Run("WarnContext", func(t *testing.T) {
		assert.NotPanics(t, func() {
			adapter.WarnContext(ctx, "warning message", "warning", "test")
		})
	})
}

// Test NoOpMetrics
func TestNoOpMetrics(t *testing.T) {
	metrics := &utilities.NoOpMetrics{}

	// These should not panic and have no effect
	t.Run("IncrementCounter", func(t *testing.T) {
		assert.NotPanics(t, func() {
			metrics.IncrementCounter("test_counter", map[string]string{"label": "value"})
		})
	})

	t.Run("RecordHistogram", func(t *testing.T) {
		assert.NotPanics(t, func() {
			metrics.RecordHistogram("test_histogram", 1.23, map[string]string{"label": "value"})
		})
	})
}

// Test utility functions (string operations)
func TestUtilityFunctions(t *testing.T) {
	db := &MockDatabase{}
	logger := &MockLogger{}
	metrics := NewMockMetrics()
	repo := utilities.NewBaseRepository(db, logger, metrics)
	ctx := context.Background()

	t.Run("contains function behavior through error classification", func(t *testing.T) {
		tests := []struct {
			errorMsg     string
			expectedType string
		}{
			{"connection failed", "connection"},
			{"timeout occurred", "timeout"},
			{"not found", "not_found"},
			{"duplicate entry", "duplicate"},
			{"random error", "unknown"},
		}

		for _, tt := range tests {
			logger.On("ErrorContext", ctx, "Repository operation failed", mock.Anything).Return()
			metrics.On("IncrementCounter", "repository_operations_error", mock.MatchedBy(func(labels map[string]string) bool {
				return labels["error"] == tt.expectedType
			})).Return()

			err := errors.New(tt.errorMsg)
			repo.OnError(ctx, utilities.OperationTypeRead, err)

			metrics.AssertExpectations(t)
			// Reset for next iteration
			metrics.Mock = mock.Mock{}
		}
	})
}

// Test metrics collection
func TestBaseRepository_MetricsCollection(t *testing.T) {
	db := &MockDatabase{}
	logger := &MockLogger{}
	metrics := NewMockMetrics()
	query := &MockQueryOperation{}

	repo := utilities.NewBaseRepository(db, logger, metrics)
	ctx := context.Background()

	// Setup expectations
	query.On("Execute", ctx, db).Return("result", nil)
	logger.On("InfoContext", mock.Anything, mock.Anything, mock.Anything).Return()
	metrics.On("IncrementCounter", mock.Anything, mock.Anything).Return()
	metrics.On("RecordHistogram", mock.Anything, mock.Anything, mock.Anything).Return()

	// Execute
	result, err := repo.ExecuteQuery(ctx, query)

	// Verify metrics were collected
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify counter was incremented (both before and success counters)
	assert.True(t, metrics.counters["repository_operations"] >= 1)
	assert.True(t, metrics.counters["repository_operations_success"] >= 1)

	// Verify histogram was recorded
	assert.Len(t, metrics.histograms["repository_query_duration"], 1)
	assert.True(t, metrics.histograms["repository_query_duration"][0] >= 0)
} // Test different operation types
func TestBaseRepository_OperationTypes(t *testing.T) {
	db := &MockDatabase{}
	logger := &MockLogger{}
	metrics := NewMockMetrics()

	repo := utilities.NewBaseRepository(db, logger, metrics)
	ctx := context.Background()

	operationTypes := []utilities.OperationType{
		utilities.OperationTypeCreate,
		utilities.OperationTypeRead,
		utilities.OperationTypeUpdate,
		utilities.OperationTypeDelete,
		utilities.OperationTypeBatch,
		utilities.OperationTypeTransaction,
	}

	for _, opType := range operationTypes {
		t.Run(string(opType), func(t *testing.T) {
			logger.On("InfoContext", ctx, "Starting repository operation", mock.Anything).Return()
			logger.On("InfoContext", ctx, "Completed repository operation", mock.Anything).Return()
			metrics.On("IncrementCounter", "repository_operations", mock.MatchedBy(func(labels map[string]string) bool {
				return labels["operation"] == string(opType)
			})).Return()
			metrics.On("IncrementCounter", "repository_operations_success", mock.MatchedBy(func(labels map[string]string) bool {
				return labels["operation"] == string(opType)
			})).Return()

			// Test the hooks with this operation type
			err := repo.BeforeOperation(ctx, opType)
			assert.NoError(t, err)

			err = repo.AfterOperation(ctx, opType, "result")
			assert.NoError(t, err)

			logger.AssertExpectations(t)
			metrics.AssertExpectations(t)

			// Reset mocks for next iteration
			logger.Mock = mock.Mock{}
			metrics.Mock = mock.Mock{}
		})
	}
}
