package uom

import (
	"errors"
	"time"
)

// Domain errors.
var (
	ErrNotFound        = errors.New("uom not found")
	ErrAlreadyExists   = errors.New("uom already exists")
	ErrEmptyName       = errors.New("uom name cannot be empty")
	ErrEmptyCreatedBy  = errors.New("created_by cannot be empty")
	ErrInvalidUOMCode  = errors.New("invalid uom code format")
	ErrInvalidCategory = errors.New("invalid uom category")
)

// UOM is the aggregate root for Unit of Measure.
type UOM struct {
	code      Code
	name      string
	category  Category
	isBaseUOM bool
	createdAt time.Time
	createdBy string
	updatedAt *time.Time
	updatedBy *string
}

// NewUOM creates a new UOM with validation.
func NewUOM(code Code, name string, category Category, createdBy string) (*UOM, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	if createdBy == "" {
		return nil, ErrEmptyCreatedBy
	}

	return &UOM{
		code:      code,
		name:      name,
		category:  category,
		isBaseUOM: false,
		createdAt: time.Now(),
		createdBy: createdBy,
	}, nil
}

// Reconstitute creates a UOM from persistence (no validation, used by repository).
func Reconstitute(
	code Code,
	name string,
	category Category,
	isBaseUOM bool,
	createdAt time.Time,
	createdBy string,
	updatedAt *time.Time,
	updatedBy *string,
) *UOM {
	return &UOM{
		code:      code,
		name:      name,
		category:  category,
		isBaseUOM: isBaseUOM,
		createdAt: createdAt,
		createdBy: createdBy,
		updatedAt: updatedAt,
		updatedBy: updatedBy,
	}
}

// Getters - expose internal state read-only.
func (u *UOM) Code() Code            { return u.code }
func (u *UOM) Name() string          { return u.name }
func (u *UOM) Category() Category    { return u.category }
func (u *UOM) IsBaseUOM() bool       { return u.isBaseUOM }
func (u *UOM) CreatedAt() time.Time  { return u.createdAt }
func (u *UOM) CreatedBy() string     { return u.createdBy }
func (u *UOM) UpdatedAt() *time.Time { return u.updatedAt }
func (u *UOM) UpdatedBy() *string    { return u.updatedBy }

// SetAsBaseUOM marks this UOM as the base unit for its category.
func (u *UOM) SetAsBaseUOM() {
	u.isBaseUOM = true
}

// Update updates the UOM properties.
func (u *UOM) Update(name string, category Category, isBaseUOM bool, updatedBy string) error {
	if name == "" {
		return ErrEmptyName
	}
	if updatedBy == "" {
		return ErrEmptyCreatedBy
	}

	u.name = name
	u.category = category
	u.isBaseUOM = isBaseUOM
	now := time.Now()
	u.updatedAt = &now
	u.updatedBy = &updatedBy
	return nil
}
