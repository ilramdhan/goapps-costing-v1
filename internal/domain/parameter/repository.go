package parameter

import "context"

// Repository defines the interface for Parameter persistence.
type Repository interface {
	// Create persists a new Parameter.
	Create(ctx context.Context, param *Parameter) error

	// GetByCode retrieves a Parameter by its code.
	GetByCode(ctx context.Context, code Code) (*Parameter, error)

	// List retrieves Parameters with optional filtering.
	List(ctx context.Context, filter ListFilter) ([]*Parameter, int64, error)

	// Update persists changes to an existing Parameter.
	Update(ctx context.Context, param *Parameter) error

	// Delete removes a Parameter by its code.
	Delete(ctx context.Context, code Code) error

	// ExistsByCode checks if a Parameter with the given code exists.
	ExistsByCode(ctx context.Context, code Code) (bool, error)
}

// ListFilter contains filtering and pagination options.
type ListFilter struct {
	Category *Category
	IsActive *bool
	Page     int
	PageSize int
}

// Offset calculates the offset for pagination.
func (f ListFilter) Offset() int {
	if f.Page <= 0 {
		f.Page = 1
	}
	return (f.Page - 1) * f.PageSize
}

// Limit returns the page size.
func (f ListFilter) Limit() int {
	if f.PageSize <= 0 {
		return 10
	}
	if f.PageSize > 100 {
		return 100
	}
	return f.PageSize
}
