package uom

import (
	"context"

	"github.com/homindolenern/goapps-costing-v1/internal/domain/uom"
)

// GetQuery represents the get UOM query.
type GetQuery struct {
	UOMCode string
}

// GetHandler handles the GetUOM query.
type GetHandler struct {
	repo uom.Repository
}

// NewGetHandler creates a new get handler.
func NewGetHandler(repo uom.Repository) *GetHandler {
	return &GetHandler{repo: repo}
}

// Handle executes the get query.
func (h *GetHandler) Handle(ctx context.Context, query GetQuery) (*uom.UOM, error) {
	code, err := uom.NewUOMCode(query.UOMCode)
	if err != nil {
		return nil, err
	}

	return h.repo.GetByCode(ctx, code)
}

// ListQuery represents the list UOMs query.
type ListQuery struct {
	Category *string
	Page     int
	PageSize int
}

// ListResult contains the list result with pagination.
type ListResult struct {
	UOMs  []*uom.UOM
	Total int64
}

// ListHandler handles the ListUOMs query.
type ListHandler struct {
	repo uom.Repository
}

// NewListHandler creates a new list handler.
func NewListHandler(repo uom.Repository) *ListHandler {
	return &ListHandler{repo: repo}
}

// Handle executes the list query.
func (h *ListHandler) Handle(ctx context.Context, query ListQuery) (*ListResult, error) {
	filter := uom.ListFilter{
		Page:     query.Page,
		PageSize: query.PageSize,
	}

	if query.Category != nil {
		cat, err := uom.NewCategory(*query.Category)
		if err != nil {
			return nil, err
		}
		filter.Category = &cat
	}

	uoms, total, err := h.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		UOMs:  uoms,
		Total: total,
	}, nil
}
