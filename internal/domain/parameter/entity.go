package parameter

import (
	"errors"
	"time"
)

// Domain errors.
var (
	ErrNotFound          = errors.New("parameter not found")
	ErrAlreadyExists     = errors.New("parameter already exists")
	ErrEmptyName         = errors.New("parameter name cannot be empty")
	ErrEmptyCreatedBy    = errors.New("created_by cannot be empty")
	ErrInvalidCode       = errors.New("invalid parameter code format")
	ErrInvalidCategory   = errors.New("invalid parameter category")
	ErrInvalidDataType   = errors.New("invalid parameter data type")
	ErrMinGreaterThanMax = errors.New("min_value cannot be greater than max_value")
	ErrDropdownNoOptions = errors.New("dropdown type requires allowed_values")
)

// Parameter is the aggregate root for configuration parameters.
type Parameter struct {
	code          Code
	name          string
	category      Category
	dataType      DataType
	uom           *string
	minValue      *float64
	maxValue      *float64
	allowedValues []string
	isMandatory   bool
	description   *string
	isActive      bool
	createdAt     time.Time
	createdBy     string
	updatedAt     *time.Time
	updatedBy     *string
}

// NewParameter creates a new Parameter with validation.
func NewParameter(
	code Code,
	name string,
	category Category,
	dataType DataType,
	createdBy string,
) (*Parameter, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	if createdBy == "" {
		return nil, ErrEmptyCreatedBy
	}

	return &Parameter{
		code:        code,
		name:        name,
		category:    category,
		dataType:    dataType,
		isActive:    true,
		isMandatory: false,
		createdAt:   time.Now(),
		createdBy:   createdBy,
	}, nil
}

// Reconstitute creates a Parameter from persistence (no validation).
func Reconstitute(
	code Code,
	name string,
	category Category,
	dataType DataType,
	uom *string,
	minValue *float64,
	maxValue *float64,
	allowedValues []string,
	isMandatory bool,
	description *string,
	isActive bool,
	createdAt time.Time,
	createdBy string,
	updatedAt *time.Time,
	updatedBy *string,
) *Parameter {
	return &Parameter{
		code:          code,
		name:          name,
		category:      category,
		dataType:      dataType,
		uom:           uom,
		minValue:      minValue,
		maxValue:      maxValue,
		allowedValues: allowedValues,
		isMandatory:   isMandatory,
		description:   description,
		isActive:      isActive,
		createdAt:     createdAt,
		createdBy:     createdBy,
		updatedAt:     updatedAt,
		updatedBy:     updatedBy,
	}
}

// Getters.
func (p *Parameter) Code() Code              { return p.code }
func (p *Parameter) Name() string            { return p.name }
func (p *Parameter) Category() Category      { return p.category }
func (p *Parameter) DataType() DataType      { return p.dataType }
func (p *Parameter) UOM() *string            { return p.uom }
func (p *Parameter) MinValue() *float64      { return p.minValue }
func (p *Parameter) MaxValue() *float64      { return p.maxValue }
func (p *Parameter) AllowedValues() []string { return p.allowedValues }
func (p *Parameter) IsMandatory() bool       { return p.isMandatory }
func (p *Parameter) Description() *string    { return p.description }
func (p *Parameter) IsActive() bool          { return p.isActive }
func (p *Parameter) CreatedAt() time.Time    { return p.createdAt }
func (p *Parameter) CreatedBy() string       { return p.createdBy }
func (p *Parameter) UpdatedAt() *time.Time   { return p.updatedAt }
func (p *Parameter) UpdatedBy() *string      { return p.updatedBy }

// SetNumericConstraints sets min/max values for numeric parameters.
func (p *Parameter) SetNumericConstraints(minVal, maxVal *float64) error {
	if minVal != nil && maxVal != nil && *minVal > *maxVal {
		return ErrMinGreaterThanMax
	}
	p.minValue = minVal
	p.maxValue = maxVal
	return nil
}

// SetAllowedValues sets dropdown options.
func (p *Parameter) SetAllowedValues(values []string) error {
	if p.dataType == DataTypeDropdown && len(values) == 0 {
		return ErrDropdownNoOptions
	}
	p.allowedValues = values
	return nil
}

// SetUOM sets the unit of measure.
func (p *Parameter) SetUOM(uom *string) {
	p.uom = uom
}

// SetDescription sets the description.
func (p *Parameter) SetDescription(desc *string) {
	p.description = desc
}

// SetMandatory sets whether the parameter is mandatory.
func (p *Parameter) SetMandatory(mandatory bool) {
	p.isMandatory = mandatory
}

// Activate activates the parameter.
func (p *Parameter) Activate() {
	p.isActive = true
}

// Deactivate deactivates the parameter.
func (p *Parameter) Deactivate() {
	p.isActive = false
}

// Update updates the parameter.
func (p *Parameter) Update(
	name string,
	category Category,
	dataType DataType,
	updatedBy string,
) error {
	if name == "" {
		return ErrEmptyName
	}
	if updatedBy == "" {
		return ErrEmptyCreatedBy
	}

	p.name = name
	p.category = category
	p.dataType = dataType
	now := time.Now()
	p.updatedAt = &now
	p.updatedBy = &updatedBy
	return nil
}
