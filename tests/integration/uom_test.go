package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/homindolenern/goapps-costing-v1/internal/domain/uom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUOMDomain_CreateValidation(t *testing.T) {
	// Test valid UOM creation
	code, err := uom.NewUOMCode("KG")
	require.NoError(t, err)
	assert.Equal(t, "KG", code.String())

	category, err := uom.NewCategory("WEIGHT")
	require.NoError(t, err)
	assert.Equal(t, "WEIGHT", category.String())

	entity, err := uom.NewUOM(code, "Kilogram", category, "admin")
	require.NoError(t, err)
	assert.Equal(t, "KG", entity.Code().String())
	assert.Equal(t, "Kilogram", entity.Name())
	assert.False(t, entity.IsBaseUOM()) // Default is false
	assert.Equal(t, "admin", entity.CreatedBy())

	// Set as base UOM
	entity.SetAsBaseUOM()
	assert.True(t, entity.IsBaseUOM())
}

func TestUOMDomain_InvalidCode(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{"empty code", ""},
		{"lowercase", "kg"},
		{"starts with number", "1KG"},
		{"too long", "ABCDEFGHIJKLMNOPQRSTUV"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uom.NewUOMCode(tc.code)
			assert.Error(t, err)
		})
	}
}

func TestUOMDomain_InvalidCategory(t *testing.T) {
	_, err := uom.NewCategory("INVALID")
	assert.Error(t, err)
}

func TestUOMDomain_Update(t *testing.T) {
	code, _ := uom.NewUOMCode("KG")
	category, _ := uom.NewCategory("WEIGHT")
	entity, _ := uom.NewUOM(code, "Kilogram", category, "admin")
	entity.SetAsBaseUOM()

	// Wait a bit to ensure update time is different
	time.Sleep(10 * time.Millisecond)

	err := entity.Update("Kilogramme", category, false, "updater")
	require.NoError(t, err)
	assert.Equal(t, "Kilogramme", entity.Name())
	assert.False(t, entity.IsBaseUOM())
	assert.NotNil(t, entity.UpdatedAt())
	assert.NotNil(t, entity.UpdatedBy())
	assert.Equal(t, "updater", *entity.UpdatedBy())
}

func TestUOMRepository_Interface(t *testing.T) {
	// This test verifies that the repository interface is properly defined
	// The actual implementation tests would require a database connection
	var repo uom.Repository
	_ = repo // Compile check
}

func TestContext_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Simulating a quick operation
	select {
	case <-time.After(10 * time.Millisecond):
		// Expected - operation completed before timeout
	case <-ctx.Done():
		t.Fatal("Context timed out unexpectedly")
	}
}
