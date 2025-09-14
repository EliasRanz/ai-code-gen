// Package database_test contains tests for database utilities
package database_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase implements utilities.Database for testing
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) BeginTx(ctx context.Context) (utilities.Transaction, error) {
	args := m.Called(ctx)
	return args.Get(0).(utilities.Transaction), args.Error(1)
}

func (m *MockDatabase) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockTransaction implements utilities.Transaction for testing
type MockTransaction struct {
	mock.Mock
}

func (m *MockTransaction) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Repository() utilities.Repository {
	args := m.Called()
	return args.Get(0).(utilities.Repository)
}

// MockTransactionOperation implements utilities.TransactionOperation for testing
type MockTransactionOperation struct {
	mock.Mock
}

func (m *MockTransactionOperation) Execute(ctx context.Context, tx *sql.Tx) (interface{}, error) {
	args := m.Called(ctx, tx)
	return args.Get(0), args.Error(1)
}

// MockQueryOperation implements utilities.QueryOperation for testing
type MockQueryOperation struct {
	mock.Mock
}

func (m *MockQueryOperation) Execute(ctx context.Context, db utilities.Database) (interface{}, error) {
	args := m.Called(ctx, db)
	return args.Get(0), args.Error(1)
}

// MockBatchOperation implements utilities.BatchOperation for testing
type MockBatchOperation struct {
	mock.Mock
}

func (m *MockBatchOperation) Execute(ctx context.Context, db utilities.Database) error {
	args := m.Called(ctx, db)
	return args.Error(0)
}

// TestRepositoryTemplate tests the template method pattern implementation
func TestRepositoryTemplate(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (*utilities.BaseRepository, *MockDatabase, *MockTransactionOperation)
		validate func(t *testing.T, err error, mockDB *MockDatabase, mockOp *MockTransactionOperation)
	}{
		{
			name: "ExecuteWithTransaction_Success",
			setup: func() (*utilities.BaseRepository, *MockDatabase, *MockTransactionOperation) {
				mockDB := &MockDatabase{}
				mockTx := &MockTransaction{}
				mockOp := &MockTransactionOperation{}

				logger := &utilities.ZerologAdapter{}
				metrics := &utilities.NoOpMetrics{}

				repo := utilities.NewBaseRepository(mockDB, logger, metrics)

				mockDB.On("BeginTx", mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit").Return(nil)
				mockTx.On("Rollback").Return(nil)
				mockOp.On("Execute", mock.Anything, mock.Anything).Return("success", nil)

				return repo, mockDB, mockOp
			},
			validate: func(t *testing.T, err error, mockDB *MockDatabase, mockOp *MockTransactionOperation) {
				assert.NoError(t, err)
				mockDB.AssertExpectations(t)
				mockOp.AssertExpectations(t)
			},
		},
		{
			name: "ExecuteWithTransaction_OperationFailure",
			setup: func() (*utilities.BaseRepository, *MockDatabase, *MockTransactionOperation) {
				mockDB := &MockDatabase{}
				mockTx := &MockTransaction{}
				mockOp := &MockTransactionOperation{}

				logger := &utilities.ZerologAdapter{}
				metrics := &utilities.NoOpMetrics{}

				repo := utilities.NewBaseRepository(mockDB, logger, metrics)

				mockDB.On("BeginTx", mock.Anything).Return(mockTx, nil)
				mockTx.On("Rollback").Return(nil)
				mockOp.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("operation failed"))

				return repo, mockDB, mockOp
			},
			validate: func(t *testing.T, err error, mockDB *MockDatabase, mockOp *MockTransactionOperation) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "operation failed")
				mockDB.AssertExpectations(t)
				mockOp.AssertExpectations(t)
			},
		},
		{
			name: "ExecuteWithTransaction_BeginTxFailure",
			setup: func() (*utilities.BaseRepository, *MockDatabase, *MockTransactionOperation) {
				mockDB := &MockDatabase{}
				mockOp := &MockTransactionOperation{}

				logger := &utilities.ZerologAdapter{}
				metrics := &utilities.NoOpMetrics{}

				repo := utilities.NewBaseRepository(mockDB, logger, metrics)

				mockDB.On("BeginTx", mock.Anything).Return((*MockTransaction)(nil), errors.New("begin tx failed"))

				return repo, mockDB, mockOp
			},
			validate: func(t *testing.T, err error, mockDB *MockDatabase, mockOp *MockTransactionOperation) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to begin transaction")
				mockDB.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockDB, mockOp := tt.setup()

			ctx := context.Background()
			err := repo.ExecuteWithTransaction(ctx, mockOp)

			tt.validate(t, err, mockDB, mockOp)
		})
	}
}

// TestExecuteQuery tests query execution with template method pattern
func TestExecuteQuery(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (*utilities.BaseRepository, *MockDatabase, *MockQueryOperation)
		validate func(t *testing.T, result interface{}, err error, mockDB *MockDatabase, mockOp *MockQueryOperation)
	}{
		{
			name: "ExecuteQuery_Success",
			setup: func() (*utilities.BaseRepository, *MockDatabase, *MockQueryOperation) {
				mockDB := &MockDatabase{}
				mockOp := &MockQueryOperation{}

				logger := &utilities.ZerologAdapter{}
				metrics := &utilities.NoOpMetrics{}

				repo := utilities.NewBaseRepository(mockDB, logger, metrics)

				mockOp.On("Execute", mock.Anything, mockDB).Return("query result", nil)

				return repo, mockDB, mockOp
			},
			validate: func(t *testing.T, result interface{}, err error, mockDB *MockDatabase, mockOp *MockQueryOperation) {
				assert.NoError(t, err)
				assert.Equal(t, "query result", result)
				mockOp.AssertExpectations(t)
			},
		},
		{
			name: "ExecuteQuery_Failure",
			setup: func() (*utilities.BaseRepository, *MockDatabase, *MockQueryOperation) {
				mockDB := &MockDatabase{}
				mockOp := &MockQueryOperation{}

				logger := &utilities.ZerologAdapter{}
				metrics := &utilities.NoOpMetrics{}

				repo := utilities.NewBaseRepository(mockDB, logger, metrics)

				mockOp.On("Execute", mock.Anything, mockDB).Return(nil, errors.New("query failed"))

				return repo, mockDB, mockOp
			},
			validate: func(t *testing.T, result interface{}, err error, mockDB *MockDatabase, mockOp *MockQueryOperation) {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "query failed")
				mockOp.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockDB, mockOp := tt.setup()

			ctx := context.Background()
			result, err := repo.ExecuteQuery(ctx, mockOp)

			tt.validate(t, result, err, mockDB, mockOp)
		})
	}
}

// TestExecuteBatch tests batch execution with template method pattern
func TestExecuteBatch(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (*utilities.BaseRepository, *MockDatabase, []*MockBatchOperation)
		validate func(t *testing.T, err error, mockDB *MockDatabase, mockOps []*MockBatchOperation)
	}{
		{
			name: "ExecuteBatch_Success",
			setup: func() (*utilities.BaseRepository, *MockDatabase, []*MockBatchOperation) {
				mockDB := &MockDatabase{}
				mockOp1 := &MockBatchOperation{}
				mockOp2 := &MockBatchOperation{}

				logger := &utilities.ZerologAdapter{}
				metrics := &utilities.NoOpMetrics{}

				repo := utilities.NewBaseRepository(mockDB, logger, metrics)

				mockOp1.On("Execute", mock.Anything, mockDB).Return(nil)
				mockOp2.On("Execute", mock.Anything, mockDB).Return(nil)

				return repo, mockDB, []*MockBatchOperation{mockOp1, mockOp2}
			},
			validate: func(t *testing.T, err error, mockDB *MockDatabase, mockOps []*MockBatchOperation) {
				assert.NoError(t, err)
				for _, op := range mockOps {
					op.AssertExpectations(t)
				}
			},
		},
		{
			name: "ExecuteBatch_FirstOperationFails",
			setup: func() (*utilities.BaseRepository, *MockDatabase, []*MockBatchOperation) {
				mockDB := &MockDatabase{}
				mockOp1 := &MockBatchOperation{}
				mockOp2 := &MockBatchOperation{}

				logger := &utilities.ZerologAdapter{}
				metrics := &utilities.NoOpMetrics{}

				repo := utilities.NewBaseRepository(mockDB, logger, metrics)

				mockOp1.On("Execute", mock.Anything, mockDB).Return(errors.New("operation 1 failed"))
				// mockOp2 should not be called

				return repo, mockDB, []*MockBatchOperation{mockOp1, mockOp2}
			},
			validate: func(t *testing.T, err error, mockDB *MockDatabase, mockOps []*MockBatchOperation) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "batch operation 0 failed")
				mockOps[0].AssertExpectations(t)
				// mockOps[1] should not have been called
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockDB, mockOps := tt.setup()

			ctx := context.Background()
			operations := make([]utilities.BatchOperation, len(mockOps))
			for i, op := range mockOps {
				operations[i] = op
			}

			err := repo.ExecuteBatch(ctx, operations)

			tt.validate(t, err, mockDB, mockOps)
		})
	}
}

// TestRepositoryHooks tests the hook methods in BaseRepository
func TestRepositoryHooks(t *testing.T) {
	logger := &utilities.ZerologAdapter{}
	metrics := &utilities.NoOpMetrics{}
	mockDB := &MockDatabase{}

	repo := utilities.NewBaseRepository(mockDB, logger, metrics)

	ctx := context.Background()

	// Test BeforeOperation
	err := repo.BeforeOperation(ctx, utilities.OperationTypeCreate)
	assert.NoError(t, err)

	// Test AfterOperation
	err = repo.AfterOperation(ctx, utilities.OperationTypeCreate, "result")
	assert.NoError(t, err)

	// Test OnError
	testErr := errors.New("test error")
	returnedErr := repo.OnError(ctx, utilities.OperationTypeCreate, testErr)
	assert.Equal(t, testErr, returnedErr)
}

// TestMetricsIntegration tests that metrics are properly collected
func TestMetricsIntegration(t *testing.T) {
	// This test would require a real metrics implementation
	// For now, we test with NoOpMetrics
	metrics := &utilities.NoOpMetrics{}

	// These should not panic
	metrics.IncrementCounter("test_counter", map[string]string{"key": "value"})
	metrics.RecordHistogram("test_histogram", 1.5, map[string]string{"key": "value"})
}

// TestZerologAdapter tests the zerolog adapter
func TestZerologAdapter(t *testing.T) {
	logger := &utilities.ZerologAdapter{}
	ctx := context.Background()

	// These should not panic
	logger.InfoContext(ctx, "test info", "key", "value")
	logger.ErrorContext(ctx, "test error", "key", "value")
	logger.WarnContext(ctx, "test warn", "key", "value")
}

// BenchmarkRepositoryTemplate benchmarks the template method pattern
func BenchmarkRepositoryTemplate(b *testing.B) {
	mockDB := &MockDatabase{}
	logger := &utilities.ZerologAdapter{}
	metrics := &utilities.NoOpMetrics{}

	repo := utilities.NewBaseRepository(mockDB, logger, metrics)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Benchmark hook methods
			repo.BeforeOperation(ctx, utilities.OperationTypeRead)
			repo.AfterOperation(ctx, utilities.OperationTypeRead, "result")
		}
	})
}
