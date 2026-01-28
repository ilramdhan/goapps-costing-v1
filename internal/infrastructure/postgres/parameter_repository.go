package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/homindolenern/goapps-costing-v1/internal/domain/parameter"
)

// ParameterRepository implements parameter.Repository interface
type ParameterRepository struct {
	db *DB
}

// NewParameterRepository creates a new Parameter repository
func NewParameterRepository(db *DB) *ParameterRepository {
	return &ParameterRepository{db: db}
}

// Verify interface implementation at compile time
var _ parameter.Repository = (*ParameterRepository)(nil)

// Create persists a new Parameter
func (r *ParameterRepository) Create(ctx context.Context, entity *parameter.Parameter) error {
	// Convert allowed_values to JSONB
	var allowedValuesJSON []byte
	var err error
	if len(entity.AllowedValues()) > 0 {
		allowedValuesJSON, err = json.Marshal(entity.AllowedValues())
		if err != nil {
			return fmt.Errorf("failed to marshal allowed_values: %w", err)
		}
	}

	query := `
		INSERT INTO mst_parameter (
			parameter_code, parameter_name, parameter_category, data_type,
			uom, min_value, max_value, allowed_values, is_mandatory,
			description, is_active, created_at, created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = r.db.ExecContext(ctx, query,
		entity.Code().String(),
		entity.Name(),
		entity.Category().String(),
		entity.DataType().String(),
		entity.UOM(),
		entity.MinValue(),
		entity.MaxValue(),
		allowedValuesJSON,
		entity.IsMandatory(),
		entity.Description(),
		entity.IsActive(),
		entity.CreatedAt(),
		entity.CreatedBy(),
	)

	return err
}

// GetByCode retrieves a Parameter by its code
func (r *ParameterRepository) GetByCode(ctx context.Context, code parameter.ParameterCode) (*parameter.Parameter, error) {
	query := `
		SELECT parameter_code, parameter_name, parameter_category, data_type,
		       uom, min_value, max_value, allowed_values, is_mandatory,
		       description, is_active, created_at, created_by, updated_at, updated_by
		FROM mst_parameter
		WHERE parameter_code = $1
	`

	var (
		paramCode        string
		paramName        string
		paramCategory    string
		dataType         string
		uom              sql.NullString
		minValue         sql.NullFloat64
		maxValue         sql.NullFloat64
		allowedValuesRaw []byte
		isMandatory      bool
		description      sql.NullString
		isActive         bool
		createdAt        time.Time
		createdBy        string
		updatedAt        sql.NullTime
		updatedBy        sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query, code.String()).Scan(
		&paramCode,
		&paramName,
		&paramCategory,
		&dataType,
		&uom,
		&minValue,
		&maxValue,
		&allowedValuesRaw,
		&isMandatory,
		&description,
		&isActive,
		&createdAt,
		&createdBy,
		&updatedAt,
		&updatedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, parameter.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Parse allowed_values from JSONB
	var allowedValues []string
	if len(allowedValuesRaw) > 0 {
		if err := json.Unmarshal(allowedValuesRaw, &allowedValues); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed_values: %w", err)
		}
	}

	// Create value objects
	codeVO, _ := parameter.NewParameterCode(paramCode)
	categoryVO, _ := parameter.NewCategory(paramCategory)
	dataTypeVO, _ := parameter.NewDataType(dataType)

	// Handle nullable fields
	var uomPtr, descPtr, updatedByPtr *string
	var minPtr, maxPtr *float64
	var updatedAtPtr *time.Time

	if uom.Valid {
		uomPtr = &uom.String
	}
	if minValue.Valid {
		minPtr = &minValue.Float64
	}
	if maxValue.Valid {
		maxPtr = &maxValue.Float64
	}
	if description.Valid {
		descPtr = &description.String
	}
	if updatedAt.Valid {
		updatedAtPtr = &updatedAt.Time
	}
	if updatedBy.Valid {
		updatedByPtr = &updatedBy.String
	}

	return parameter.Reconstitute(
		codeVO,
		paramName,
		categoryVO,
		dataTypeVO,
		uomPtr,
		minPtr,
		maxPtr,
		allowedValues,
		isMandatory,
		descPtr,
		isActive,
		createdAt,
		createdBy,
		updatedAtPtr,
		updatedByPtr,
	), nil
}

// List retrieves Parameters with optional filtering
func (r *ParameterRepository) List(ctx context.Context, filter parameter.ListFilter) ([]*parameter.Parameter, int64, error) {
	// Base query
	baseQuery := `FROM mst_parameter WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filter.Category != nil {
		baseQuery += fmt.Sprintf(` AND parameter_category = $%d`, argIndex)
		args = append(args, filter.Category.String())
		argIndex++
	}
	if filter.IsActive != nil {
		baseQuery += fmt.Sprintf(` AND is_active = $%d`, argIndex)
		args = append(args, *filter.IsActive)
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
	dataQuery := `SELECT parameter_code, parameter_name, parameter_category, data_type,
	              uom, min_value, max_value, allowed_values, is_mandatory,
	              description, is_active, created_at, created_by, updated_at, updated_by ` + baseQuery +
		fmt.Sprintf(` ORDER BY parameter_code LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	args = append(args, filter.Limit(), filter.Offset())

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []*parameter.Parameter
	for rows.Next() {
		var (
			paramCode        string
			paramName        string
			paramCategory    string
			dataType         string
			uom              sql.NullString
			minValue         sql.NullFloat64
			maxValue         sql.NullFloat64
			allowedValuesRaw []byte
			isMandatory      bool
			description      sql.NullString
			isActive         bool
			createdAt        time.Time
			createdBy        string
			updatedAt        sql.NullTime
			updatedBy        sql.NullString
		)

		if err := rows.Scan(
			&paramCode,
			&paramName,
			&paramCategory,
			&dataType,
			&uom,
			&minValue,
			&maxValue,
			&allowedValuesRaw,
			&isMandatory,
			&description,
			&isActive,
			&createdAt,
			&createdBy,
			&updatedAt,
			&updatedBy,
		); err != nil {
			return nil, 0, err
		}

		var allowedValues []string
		if len(allowedValuesRaw) > 0 {
			json.Unmarshal(allowedValuesRaw, &allowedValues)
		}

		codeVO, _ := parameter.NewParameterCode(paramCode)
		categoryVO, _ := parameter.NewCategory(paramCategory)
		dataTypeVO, _ := parameter.NewDataType(dataType)

		var uomPtr, descPtr, updatedByPtr *string
		var minPtr, maxPtr *float64
		var updatedAtPtr *time.Time

		if uom.Valid {
			uomPtr = &uom.String
		}
		if minValue.Valid {
			minPtr = &minValue.Float64
		}
		if maxValue.Valid {
			maxPtr = &maxValue.Float64
		}
		if description.Valid {
			descPtr = &description.String
		}
		if updatedAt.Valid {
			updatedAtPtr = &updatedAt.Time
		}
		if updatedBy.Valid {
			updatedByPtr = &updatedBy.String
		}

		entity := parameter.Reconstitute(
			codeVO,
			paramName,
			categoryVO,
			dataTypeVO,
			uomPtr,
			minPtr,
			maxPtr,
			allowedValues,
			isMandatory,
			descPtr,
			isActive,
			createdAt,
			createdBy,
			updatedAtPtr,
			updatedByPtr,
		)
		result = append(result, entity)
	}

	return result, total, rows.Err()
}

// Update persists changes to an existing Parameter
func (r *ParameterRepository) Update(ctx context.Context, entity *parameter.Parameter) error {
	var allowedValuesJSON []byte
	var err error
	if len(entity.AllowedValues()) > 0 {
		allowedValuesJSON, err = json.Marshal(entity.AllowedValues())
		if err != nil {
			return fmt.Errorf("failed to marshal allowed_values: %w", err)
		}
	}

	query := `
		UPDATE mst_parameter
		SET parameter_name = $2, parameter_category = $3, data_type = $4,
		    uom = $5, min_value = $6, max_value = $7, allowed_values = $8,
		    is_mandatory = $9, description = $10, is_active = $11,
		    updated_at = $12, updated_by = $13
		WHERE parameter_code = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		entity.Code().String(),
		entity.Name(),
		entity.Category().String(),
		entity.DataType().String(),
		entity.UOM(),
		entity.MinValue(),
		entity.MaxValue(),
		allowedValuesJSON,
		entity.IsMandatory(),
		entity.Description(),
		entity.IsActive(),
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
		return parameter.ErrNotFound
	}

	return nil
}

// Delete removes a Parameter by its code
func (r *ParameterRepository) Delete(ctx context.Context, code parameter.ParameterCode) error {
	query := `DELETE FROM mst_parameter WHERE parameter_code = $1`

	result, err := r.db.ExecContext(ctx, query, code.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return parameter.ErrNotFound
	}

	return nil
}

// ExistsByCode checks if a Parameter with the given code exists
func (r *ParameterRepository) ExistsByCode(ctx context.Context, code parameter.ParameterCode) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM mst_parameter WHERE parameter_code = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, code.String()).Scan(&exists)
	return exists, err
}
