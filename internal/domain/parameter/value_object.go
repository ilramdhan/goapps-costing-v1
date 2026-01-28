package parameter

import (
	"regexp"
)

// Code is a value object for parameter identifier.
type Code string

var parameterCodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,49}$`)

// NewParameterCode creates a validated parameter code.
func NewParameterCode(code string) (Code, error) {
	if !parameterCodePattern.MatchString(code) {
		return "", ErrInvalidCode
	}
	return Code(code), nil
}

// String returns the string representation.
func (c Code) String() string {
	return string(c)
}

// Category represents the type/category of Parameter.
type Category string

const (
	CategoryMachine  Category = "MACHINE"
	CategoryMaterial Category = "MATERIAL"
	CategoryQuality  Category = "QUALITY"
	CategoryOutput   Category = "OUTPUT"
	CategoryProcess  Category = "PROCESS"
)

// NewCategory creates a validated category.
func NewCategory(category string) (Category, error) {
	switch Category(category) {
	case CategoryMachine, CategoryMaterial, CategoryQuality, CategoryOutput, CategoryProcess:
		return Category(category), nil
	default:
		return "", ErrInvalidCategory
	}
}

// String returns the string representation.
func (c Category) String() string {
	return string(c)
}

// DataType represents the data type of parameter value.
type DataType string

const (
	DataTypeNumeric  DataType = "NUMERIC"
	DataTypeText     DataType = "TEXT"
	DataTypeBoolean  DataType = "BOOLEAN"
	DataTypeDropdown DataType = "DROPDOWN"
)

// NewDataType creates a validated data type.
func NewDataType(dataType string) (DataType, error) {
	switch DataType(dataType) {
	case DataTypeNumeric, DataTypeText, DataTypeBoolean, DataTypeDropdown:
		return DataType(dataType), nil
	default:
		return "", ErrInvalidDataType
	}
}

// String returns the string representation.
func (d DataType) String() string {
	return string(d)
}
