package uom

import (
	"regexp"
)

// UOMCode is a value object for UOM identifier
type UOMCode string

var uomCodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,19}$`)

// NewUOMCode creates a validated UOM code
func NewUOMCode(code string) (UOMCode, error) {
	if !uomCodePattern.MatchString(code) {
		return "", ErrInvalidUOMCode
	}
	return UOMCode(code), nil
}

// String returns the string representation
func (c UOMCode) String() string {
	return string(c)
}

// Category represents the type/category of UOM
type Category string

const (
	CategoryWeight   Category = "WEIGHT"
	CategoryVolume   Category = "VOLUME"
	CategoryQuantity Category = "QUANTITY"
	CategoryLength   Category = "LENGTH"
)

// NewCategory creates a validated category
func NewCategory(category string) (Category, error) {
	switch Category(category) {
	case CategoryWeight, CategoryVolume, CategoryQuantity, CategoryLength:
		return Category(category), nil
	default:
		return "", ErrInvalidCategory
	}
}

// String returns the string representation
func (c Category) String() string {
	return string(c)
}

// IsValid checks if the category is valid
func (c Category) IsValid() bool {
	switch c {
	case CategoryWeight, CategoryVolume, CategoryQuantity, CategoryLength:
		return true
	default:
		return false
	}
}
