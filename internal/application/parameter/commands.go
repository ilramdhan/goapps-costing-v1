package parameter

import (
	"context"

	"github.com/homindolenern/goapps-costing-v1/internal/domain/parameter"
)

// CreateCommand represents the create Parameter command.
type CreateCommand struct {
	ParameterCode string
	ParameterName string
	Category      string
	DataType      string
	UOM           *string
	MinValue      *float64
	MaxValue      *float64
	AllowedValues []string
	IsMandatory   bool
	Description   *string
	CreatedBy     string
}

// CreateHandler handles the CreateParameter command.
type CreateHandler struct {
	repo parameter.Repository
}

// NewCreateHandler creates a new create handler.
func NewCreateHandler(repo parameter.Repository) *CreateHandler {
	return &CreateHandler{repo: repo}
}

// Handle executes the create command.
func (h *CreateHandler) Handle(ctx context.Context, cmd CreateCommand) (*parameter.Parameter, error) {
	// 1. Create and validate value objects
	code, err := parameter.NewParameterCode(cmd.ParameterCode)
	if err != nil {
		return nil, err
	}

	category, err := parameter.NewCategory(cmd.Category)
	if err != nil {
		return nil, err
	}

	dataType, err := parameter.NewDataType(cmd.DataType)
	if err != nil {
		return nil, err
	}

	// 2. Check for duplicates
	exists, err := h.repo.ExistsByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, parameter.ErrAlreadyExists
	}

	// 3. Create domain entity
	entity, err := parameter.NewParameter(code, cmd.ParameterName, category, dataType, cmd.CreatedBy)
	if err != nil {
		return nil, err
	}

	// 4. Set optional fields
	entity.SetUOM(cmd.UOM)
	entity.SetDescription(cmd.Description)
	entity.SetMandatory(cmd.IsMandatory)

	if err := entity.SetNumericConstraints(cmd.MinValue, cmd.MaxValue); err != nil {
		return nil, err
	}
	if err := entity.SetAllowedValues(cmd.AllowedValues); err != nil {
		return nil, err
	}

	// 5. Persist
	if err := h.repo.Create(ctx, entity); err != nil {
		return nil, err
	}

	return entity, nil
}

// UpdateCommand represents the update Parameter command.
type UpdateCommand struct {
	ParameterCode string
	ParameterName string
	Category      string
	DataType      string
	UOM           *string
	MinValue      *float64
	MaxValue      *float64
	AllowedValues []string
	IsMandatory   bool
	Description   *string
	IsActive      bool
	UpdatedBy     string
}

// UpdateHandler handles the UpdateParameter command.
type UpdateHandler struct {
	repo parameter.Repository
}

// NewUpdateHandler creates a new update handler.
func NewUpdateHandler(repo parameter.Repository) *UpdateHandler {
	return &UpdateHandler{repo: repo}
}

// Handle executes the update command.
func (h *UpdateHandler) Handle(ctx context.Context, cmd UpdateCommand) (*parameter.Parameter, error) {
	// 1. Create value objects
	code, err := parameter.NewParameterCode(cmd.ParameterCode)
	if err != nil {
		return nil, err
	}

	category, err := parameter.NewCategory(cmd.Category)
	if err != nil {
		return nil, err
	}

	dataType, err := parameter.NewDataType(cmd.DataType)
	if err != nil {
		return nil, err
	}

	// 2. Get existing entity
	entity, err := h.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 3. Update entity
	if err := entity.Update(cmd.ParameterName, category, dataType, cmd.UpdatedBy); err != nil {
		return nil, err
	}

	entity.SetUOM(cmd.UOM)
	entity.SetDescription(cmd.Description)
	entity.SetMandatory(cmd.IsMandatory)

	if err := entity.SetNumericConstraints(cmd.MinValue, cmd.MaxValue); err != nil {
		return nil, err
	}
	if err := entity.SetAllowedValues(cmd.AllowedValues); err != nil {
		return nil, err
	}

	if cmd.IsActive {
		entity.Activate()
	} else {
		entity.Deactivate()
	}

	// 4. Persist
	if err := h.repo.Update(ctx, entity); err != nil {
		return nil, err
	}

	return entity, nil
}

// DeleteCommand represents the delete Parameter command.
type DeleteCommand struct {
	ParameterCode string
}

// DeleteHandler handles the DeleteParameter command.
type DeleteHandler struct {
	repo parameter.Repository
}

// NewDeleteHandler creates a new delete handler.
func NewDeleteHandler(repo parameter.Repository) *DeleteHandler {
	return &DeleteHandler{repo: repo}
}

// Handle executes the delete command.
func (h *DeleteHandler) Handle(ctx context.Context, cmd DeleteCommand) error {
	code, err := parameter.NewParameterCode(cmd.ParameterCode)
	if err != nil {
		return err
	}

	return h.repo.Delete(ctx, code)
}
