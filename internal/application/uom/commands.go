package uom

import (
	"context"

	"github.com/homindolenern/goapps-costing-v1/internal/domain/uom"
)

// CreateCommand represents the create UOM command
type CreateCommand struct {
	UOMCode   string
	UOMName   string
	Category  string
	IsBaseUOM bool
	CreatedBy string
}

// CreateHandler handles the CreateUOM command
type CreateHandler struct {
	repo uom.Repository
}

// NewCreateHandler creates a new create handler
func NewCreateHandler(repo uom.Repository) *CreateHandler {
	return &CreateHandler{repo: repo}
}

// Handle executes the create command
func (h *CreateHandler) Handle(ctx context.Context, cmd CreateCommand) (*uom.UOM, error) {
	// 1. Create and validate value objects
	code, err := uom.NewUOMCode(cmd.UOMCode)
	if err != nil {
		return nil, err
	}

	category, err := uom.NewCategory(cmd.Category)
	if err != nil {
		return nil, err
	}

	// 2. Check for duplicates
	exists, err := h.repo.ExistsByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, uom.ErrAlreadyExists
	}

	// 3. Create domain entity
	entity, err := uom.NewUOM(code, cmd.UOMName, category, cmd.CreatedBy)
	if err != nil {
		return nil, err
	}

	if cmd.IsBaseUOM {
		entity.SetAsBaseUOM()
	}

	// 4. Persist
	if err := h.repo.Create(ctx, entity); err != nil {
		return nil, err
	}

	return entity, nil
}

// UpdateCommand represents the update UOM command
type UpdateCommand struct {
	UOMCode   string
	UOMName   string
	Category  string
	IsBaseUOM bool
	UpdatedBy string
}

// UpdateHandler handles the UpdateUOM command
type UpdateHandler struct {
	repo uom.Repository
}

// NewUpdateHandler creates a new update handler
func NewUpdateHandler(repo uom.Repository) *UpdateHandler {
	return &UpdateHandler{repo: repo}
}

// Handle executes the update command
func (h *UpdateHandler) Handle(ctx context.Context, cmd UpdateCommand) (*uom.UOM, error) {
	// 1. Create value objects
	code, err := uom.NewUOMCode(cmd.UOMCode)
	if err != nil {
		return nil, err
	}

	category, err := uom.NewCategory(cmd.Category)
	if err != nil {
		return nil, err
	}

	// 2. Get existing entity
	entity, err := h.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 3. Update entity
	if err := entity.Update(cmd.UOMName, category, cmd.IsBaseUOM, cmd.UpdatedBy); err != nil {
		return nil, err
	}

	// 4. Persist
	if err := h.repo.Update(ctx, entity); err != nil {
		return nil, err
	}

	return entity, nil
}

// DeleteCommand represents the delete UOM command
type DeleteCommand struct {
	UOMCode string
}

// DeleteHandler handles the DeleteUOM command
type DeleteHandler struct {
	repo uom.Repository
}

// NewDeleteHandler creates a new delete handler
func NewDeleteHandler(repo uom.Repository) *DeleteHandler {
	return &DeleteHandler{repo: repo}
}

// Handle executes the delete command
func (h *DeleteHandler) Handle(ctx context.Context, cmd DeleteCommand) error {
	code, err := uom.NewUOMCode(cmd.UOMCode)
	if err != nil {
		return err
	}

	return h.repo.Delete(ctx, code)
}
