package uom

import "context"

// Repository defines the interface for UOM persistence
// This interface is defined in domain, implemented in infrastructure
type Repository interface {
	// Create persists a new UOM
	Create(ctx context.Context, uom *UOM) error

	// GetByCode retrieves a UOM by its code
	GetByCode(ctx context.Context, code UOMCode) (*UOM, error)

	// List retrieves UOMs with optional filtering
	List(ctx context.Context, filter ListFilter) ([]*UOM, int64, error)

	// Update persists changes to an existing UOM
	Update(ctx context.Context, uom *UOM) error

	// Delete removes a UOM by its code
	Delete(ctx context.Context, code UOMCode) error

	// ExistsByCode checks if a UOM with the given code exists
	ExistsByCode(ctx context.Context, code UOMCode) (bool, error)
}

// ListFilter contains filtering and pagination options for listing UOMs
type ListFilter struct {
	Category *Category
	Page     int
	PageSize int
}

// Offset calculates the offset for pagination
func (f ListFilter) Offset() int {
	if f.Page <= 0 {
		f.Page = 1
	}
	return (f.Page - 1) * f.PageSize
}

// Limit returns the page size
func (f ListFilter) Limit() int {
	if f.PageSize <= 0 {
		return 10
	}
	if f.PageSize > 100 {
		return 100
	}
	return f.PageSize
}
