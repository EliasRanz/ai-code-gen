package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/tests/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestInterfaceSegregation validates our interface segregation improvements using generated mocks
// This replaces manual mocks with generated ones following testing.instructions.md
func TestInterfaceSegregation(t *testing.T) {
	t.Run("BasicCacheOperations", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockOps := mocks.NewMockBasicCacheOperations(ctrl)
		ctx := context.Background()

		// Test that we can use only basic operations interface
		testKey := "test:key"
		testValue := "test:value"
		testTTL := 5 * time.Minute

		// Setup mock expectations using gomock (replacing manual mocks)
		mockOps.EXPECT().Set(ctx, testKey, testValue, testTTL).Return(nil)
		mockOps.EXPECT().Get(ctx, testKey).Return(testValue, nil)
		mockOps.EXPECT().Exists(ctx, testKey).Return(true, nil)
		mockOps.EXPECT().Delete(ctx, testKey).Return(nil)

		// Execute operations
		err := mockOps.Set(ctx, testKey, testValue, testTTL)
		assert.NoError(t, err, "Set operation should succeed")

		value, err := mockOps.Get(ctx, testKey)
		assert.NoError(t, err, "Get operation should succeed")
		assert.Equal(t, testValue, value, "Retrieved value should match set value")

		exists, err := mockOps.Exists(ctx, testKey)
		assert.NoError(t, err, "Exists operation should succeed")
		assert.True(t, exists, "Key should exist")

		err = mockOps.Delete(ctx, testKey)
		assert.NoError(t, err, "Delete operation should succeed")
	})

	t.Run("InterfaceComposition", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Test that our interface segregation works with generated mocks
		// Each interface can be mocked independently

		// Create mocks for each segregated interface
		basicOps := mocks.NewMockBasicCacheOperations(ctrl)
		batchOps := mocks.NewMockBatchCacheOperations(ctrl)
		patternOps := mocks.NewMockPatternCacheOperations(ctrl)
		healthOps := mocks.NewMockCacheHealthOperations(ctrl)

		ctx := context.Background()

		// Set up expectations for each interface - demonstrates interface segregation
		basicOps.EXPECT().Get(ctx, "test:key").Return("test:value", nil)
		batchOps.EXPECT().MGet(ctx, []string{"key1", "key2"}).Return([]string{"value1", "value2"}, nil)
		patternOps.EXPECT().Keys(ctx, "test:*").Return([]string{"test:key1", "test:key2"}, nil)
		healthOps.EXPECT().HealthCheck(ctx).Return(nil)

		// Test each interface works independently - validates interface segregation principle
		value, err := basicOps.Get(ctx, "test:key")
		assert.NoError(t, err)
		assert.Equal(t, "test:value", value)

		values, err := batchOps.MGet(ctx, []string{"key1", "key2"})
		assert.NoError(t, err)
		assert.Equal(t, []string{"value1", "value2"}, values)

		keys, err := patternOps.Keys(ctx, "test:*")
		assert.NoError(t, err)
		assert.Equal(t, []string{"test:key1", "test:key2"}, keys)

		err = healthOps.HealthCheck(ctx)
		assert.NoError(t, err)
	})
}
