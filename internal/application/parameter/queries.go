package parameter

import (
	"context"

	"github.com/homindolenern/goapps-costing-v1/internal/domain/parameter"
)

// GetQuery represents the get Parameter query.
type GetQuery struct {
	ParameterCode string
}

// GetHandler handles the GetParameter query.
type GetHandler struct {
	repo parameter.Repository
}

// NewGetHandler creates a new get handler.
func NewGetHandler(repo parameter.Repository) *GetHandler {
	return &GetHandler{repo: repo}
}

// Handle executes the get query.
func (h *GetHandler) Handle(ctx context.Context, query GetQuery) (*parameter.Parameter, error) {
	code, err := parameter.NewParameterCode(query.ParameterCode)
	if err != nil {
		return nil, err
	}

	return h.repo.GetByCode(ctx, code)
}

// ListQuery represents the list Parameters query.
type ListQuery struct {
	Category *string
	IsActive *bool
	Page     int
	PageSize int
}

// ListResult contains the list result with pagination.
type ListResult struct {
	Parameters []*parameter.Parameter
	Total      int64
}

// ListHandler handles the ListParameters query.
type ListHandler struct {
	repo parameter.Repository
}

// NewListHandler creates a new list handler.
func NewListHandler(repo parameter.Repository) *ListHandler {
	return &ListHandler{repo: repo}
}

// Handle executes the list query.
func (h *ListHandler) Handle(ctx context.Context, query ListQuery) (*ListResult, error) {
	filter := parameter.ListFilter{
		Page:     query.Page,
		PageSize: query.PageSize,
		IsActive: query.IsActive,
	}

	if query.Category != nil {
		cat, err := parameter.NewCategory(*query.Category)
		if err != nil {
			return nil, err
		}
		filter.Category = &cat
	}

	params, total, err := h.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Parameters: params,
		Total:      total,
	}, nil
}
