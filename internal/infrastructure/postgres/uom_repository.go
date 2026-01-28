package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/homindolenern/goapps-costing-v1/internal/domain/uom"
)

// UOMRepository implements uom.Repository interface.
type UOMRepository struct {
	db *DB
}

// NewUOMRepository creates a new UOM repository.
func NewUOMRepository(db *DB) *UOMRepository {
	return &UOMRepository{db: db}
}

// Verify interface implementation at compile time.
var _ uom.Repository = (*UOMRepository)(nil)

// Create persists a new UOM.
func (r *UOMRepository) Create(ctx context.Context, entity *uom.UOM) error {
	query := `
		INSERT INTO mst_uom (uom_code, uom_name, uom_category, is_base_uom, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		entity.Code().String(),
		entity.Name(),
		entity.Category().String(),
		entity.IsBaseUOM(),
		entity.CreatedAt(),
		entity.CreatedBy(),
	)

	return err
}

// GetByCode retrieves a UOM by its code.
func (r *UOMRepository) GetByCode(ctx context.Context, code uom.Code) (*uom.UOM, error) {
	query := `
		SELECT uom_code, uom_name, uom_category, is_base_uom, 
		       created_at, created_by, updated_at, updated_by
		FROM mst_uom
		WHERE uom_code = $1
	`

	var (
		uomCode     string
		uomName     string
		uomCategory string
		isBaseUOM   bool
		createdAt   time.Time
		createdBy   string
		updatedAt   sql.NullTime
		updatedBy   sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query, code.String()).Scan(
		&uomCode,
		&uomName,
		&uomCategory,
		&isBaseUOM,
		&createdAt,
		&createdBy,
		&updatedAt,
		&updatedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, uom.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Create value objects
	uomCodeVO, _ := uom.NewUOMCode(uomCode)
	categoryVO, _ := uom.NewCategory(uomCategory)

	// Handle nullable fields
	var updatedAtPtr *time.Time
	var updatedByPtr *string
	if updatedAt.Valid {
		updatedAtPtr = &updatedAt.Time
	}
	if updatedBy.Valid {
		updatedByPtr = &updatedBy.String
	}

	return uom.Reconstitute(
		uomCodeVO,
		uomName,
		categoryVO,
		isBaseUOM,
		createdAt,
		createdBy,
		updatedAtPtr,
		updatedByPtr,
	), nil
}

// List retrieves UOMs with optional filtering.
func (r *UOMRepository) List(ctx context.Context, filter uom.ListFilter) ([]*uom.UOM, int64, error) {
	// Base query
	baseQuery := `FROM mst_uom WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	// Apply category filter
	if filter.Category != nil {
		baseQuery += ` AND uom_category = $` + itoa(argIndex)
		args = append(args, filter.Category.String())
		argIndex++
	}

	// Count query
	countQuery := `SELECT COUNT(*) ` + baseQuery
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Data query with pagination
	dataQuery := `SELECT uom_code, uom_name, uom_category, is_base_uom, 
	              created_at, created_by, updated_at, updated_by ` + baseQuery +
		` ORDER BY uom_code LIMIT $` + itoa(argIndex) + ` OFFSET $` + itoa(argIndex+1)
	args = append(args, filter.Limit(), filter.Offset())

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []*uom.UOM
	for rows.Next() {
		var (
			uomCode     string
			uomName     string
			uomCategory string
			isBaseUOM   bool
			createdAt   time.Time
			createdBy   string
			updatedAt   sql.NullTime
			updatedBy   sql.NullString
		)

		if err := rows.Scan(
			&uomCode,
			&uomName,
			&uomCategory,
			&isBaseUOM,
			&createdAt,
			&createdBy,
			&updatedAt,
			&updatedBy,
		); err != nil {
			return nil, 0, err
		}

		uomCodeVO, _ := uom.NewUOMCode(uomCode)
		categoryVO, _ := uom.NewCategory(uomCategory)

		var updatedAtPtr *time.Time
		var updatedByPtr *string
		if updatedAt.Valid {
			updatedAtPtr = &updatedAt.Time
		}
		if updatedBy.Valid {
			updatedByPtr = &updatedBy.String
		}

		entity := uom.Reconstitute(
			uomCodeVO,
			uomName,
			categoryVO,
			isBaseUOM,
			createdAt,
			createdBy,
			updatedAtPtr,
			updatedByPtr,
		)
		result = append(result, entity)
	}

	return result, total, rows.Err()
}

// Update persists changes to an existing UOM.
func (r *UOMRepository) Update(ctx context.Context, entity *uom.UOM) error {
	query := `
		UPDATE mst_uom
		SET uom_name = $2, uom_category = $3, is_base_uom = $4, 
		    updated_at = $5, updated_by = $6
		WHERE uom_code = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		entity.Code().String(),
		entity.Name(),
		entity.Category().String(),
		entity.IsBaseUOM(),
		entity.UpdatedAt(),
		entity.UpdatedBy(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return uom.ErrNotFound
	}

	return nil
}

// Delete removes a UOM by its code.
func (r *UOMRepository) Delete(ctx context.Context, code uom.Code) error {
	query := `DELETE FROM mst_uom WHERE uom_code = $1`

	result, err := r.db.ExecContext(ctx, query, code.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return uom.ErrNotFound
	}

	return nil
}

// ExistsByCode checks if a UOM with the given code exists.
func (r *UOMRepository) ExistsByCode(ctx context.Context, code uom.Code) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM mst_uom WHERE uom_code = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, code.String()).Scan(&exists)
	return exists, err
}

// Helper function.
func itoa(i int) string {
	return string(rune('0' + i))
}
