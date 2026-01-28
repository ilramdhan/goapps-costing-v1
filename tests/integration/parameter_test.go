package integration_test

import (
	"testing"

	"github.com/homindolenern/goapps-costing-v1/internal/domain/parameter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParameterDomain_CreateValidation(t *testing.T) {
	code, err := parameter.NewParameterCode("THREAD_COUNT")
	require.NoError(t, err)
	assert.Equal(t, "THREAD_COUNT", code.String())

	category, err := parameter.NewCategory("MACHINE")
	require.NoError(t, err)
	assert.Equal(t, "MACHINE", category.String())

	dataType, err := parameter.NewDataType("NUMERIC")
	require.NoError(t, err)
	assert.Equal(t, "NUMERIC", dataType.String())

	entity, err := parameter.NewParameter(code, "Thread Count", category, dataType, "admin")
	require.NoError(t, err)
	assert.Equal(t, "THREAD_COUNT", entity.Code().String())
	assert.Equal(t, "Thread Count", entity.Name())
	assert.True(t, entity.IsActive())
}

func TestParameterDomain_InvalidCode(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{"empty code", ""},
		{"lowercase", "thread_count"},
		{"starts with number", "1THREAD"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parameter.NewParameterCode(tc.code)
			assert.Error(t, err)
		})
	}
}

func TestParameterDomain_NumericConstraints(t *testing.T) {
	code, _ := parameter.NewParameterCode("RPM")
	category, _ := parameter.NewCategory("MACHINE")
	dataType, _ := parameter.NewDataType("NUMERIC")
	entity, _ := parameter.NewParameter(code, "Rotation Per Minute", category, dataType, "admin")

	// Valid range
	min := 100.0
	max := 10000.0
	err := entity.SetNumericConstraints(&min, &max)
	require.NoError(t, err)
	assert.Equal(t, &min, entity.MinValue())
	assert.Equal(t, &max, entity.MaxValue())

	// Invalid range (min > max)
	invalidMin := 500.0
	invalidMax := 100.0
	err = entity.SetNumericConstraints(&invalidMin, &invalidMax)
	assert.Error(t, err)
}

func TestParameterDomain_DropdownRequiresOptions(t *testing.T) {
	code, _ := parameter.NewParameterCode("QUALITY_GRADE")
	category, _ := parameter.NewCategory("QUALITY")
	dataType, _ := parameter.NewDataType("DROPDOWN")
	entity, _ := parameter.NewParameter(code, "Quality Grade", category, dataType, "admin")

	// Dropdown without options should fail validation on SetAllowedValues
	err := entity.SetAllowedValues([]string{})
	assert.Error(t, err)

	// Dropdown with options should succeed
	err = entity.SetAllowedValues([]string{"A", "B", "C"})
	require.NoError(t, err)
	assert.Equal(t, []string{"A", "B", "C"}, entity.AllowedValues())
}

func TestParameterDomain_ActivateDeactivate(t *testing.T) {
	code, _ := parameter.NewParameterCode("TEST_PARAM")
	category, _ := parameter.NewCategory("PROCESS")
	dataType, _ := parameter.NewDataType("TEXT")
	entity, _ := parameter.NewParameter(code, "Test Parameter", category, dataType, "admin")

	assert.True(t, entity.IsActive())

	entity.Deactivate()
	assert.False(t, entity.IsActive())

	entity.Activate()
	assert.True(t, entity.IsActive())
}
