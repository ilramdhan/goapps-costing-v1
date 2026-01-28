package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/homindolenern/goapps-costing-v1/gen/go/costing/v1"
	appuom "github.com/homindolenern/goapps-costing-v1/internal/application/uom"
	"github.com/homindolenern/goapps-costing-v1/internal/domain/uom"
)

// UOMHandler implements the gRPC UOMService
type UOMHandler struct {
	pb.UnimplementedUOMServiceServer
	createHandler *appuom.CreateHandler
	updateHandler *appuom.UpdateHandler
	deleteHandler *appuom.DeleteHandler
	getHandler    *appuom.GetHandler
	listHandler   *appuom.ListHandler
}

// NewUOMHandler creates a new UOM handler
func NewUOMHandler(
	createHandler *appuom.CreateHandler,
	updateHandler *appuom.UpdateHandler,
	deleteHandler *appuom.DeleteHandler,
	getHandler *appuom.GetHandler,
	listHandler *appuom.ListHandler,
) *UOMHandler {
	return &UOMHandler{
		createHandler: createHandler,
		updateHandler: updateHandler,
		deleteHandler: deleteHandler,
		getHandler:    getHandler,
		listHandler:   listHandler,
	}
}

// CreateUOM creates a new Unit of Measure
func (h *UOMHandler) CreateUOM(ctx context.Context, req *pb.CreateUOMRequest) (*pb.CreateUOMResponse, error) {
	cmd := appuom.CreateCommand{
		UOMCode:   req.UomCode,
		UOMName:   req.UomName,
		Category:  pbCategoryToString(req.UomCategory),
		IsBaseUOM: req.IsBaseUom,
		CreatedBy: "system", // TODO: Extract from context/auth
	}

	entity, err := h.createHandler.Handle(ctx, cmd)
	if err != nil {
		return &pb.CreateUOMResponse{
			Base: errorToBaseResponse(err),
		}, nil
	}

	return &pb.CreateUOMResponse{
		Base: successResponse("UOM created successfully"),
		Data: entityToProto(entity),
	}, nil
}

// GetUOM retrieves a Unit of Measure by code
func (h *UOMHandler) GetUOM(ctx context.Context, req *pb.GetUOMRequest) (*pb.GetUOMResponse, error) {
	query := appuom.GetQuery{UOMCode: req.UomCode}

	entity, err := h.getHandler.Handle(ctx, query)
	if err != nil {
		return &pb.GetUOMResponse{
			Base: errorToBaseResponse(err),
		}, nil
	}

	return &pb.GetUOMResponse{
		Base: successResponse("UOM retrieved successfully"),
		Data: entityToProto(entity),
	}, nil
}

// ListUOMs retrieves a paginated list of Units of Measure
func (h *UOMHandler) ListUOMs(ctx context.Context, req *pb.ListUOMsRequest) (*pb.ListUOMsResponse, error) {
	query := appuom.ListQuery{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	if req.Category != nil && *req.Category != pb.UOMCategory_UOM_CATEGORY_UNSPECIFIED {
		cat := pbCategoryToString(*req.Category)
		query.Category = &cat
	}

	result, err := h.listHandler.Handle(ctx, query)
	if err != nil {
		return &pb.ListUOMsResponse{
			Base: errorToBaseResponse(err),
		}, nil
	}

	data := make([]*pb.UOM, len(result.UOMs))
	for i, entity := range result.UOMs {
		data[i] = entityToProto(entity)
	}

	totalPages := int32(result.Total) / req.PageSize
	if int32(result.Total)%req.PageSize > 0 {
		totalPages++
	}

	return &pb.ListUOMsResponse{
		Base: successResponse("UOMs retrieved successfully"),
		Data: data,
		Pagination: &pb.PaginationMeta{
			CurrentPage: req.Page,
			PageSize:    req.PageSize,
			TotalItems:  result.Total,
			TotalPages:  totalPages,
		},
	}, nil
}

// UpdateUOM updates an existing Unit of Measure
func (h *UOMHandler) UpdateUOM(ctx context.Context, req *pb.UpdateUOMRequest) (*pb.UpdateUOMResponse, error) {
	cmd := appuom.UpdateCommand{
		UOMCode:   req.UomCode,
		UOMName:   req.UomName,
		Category:  pbCategoryToString(req.UomCategory),
		IsBaseUOM: req.IsBaseUom,
		UpdatedBy: "system", // TODO: Extract from context/auth
	}

	entity, err := h.updateHandler.Handle(ctx, cmd)
	if err != nil {
		return &pb.UpdateUOMResponse{
			Base: errorToBaseResponse(err),
		}, nil
	}

	return &pb.UpdateUOMResponse{
		Base: successResponse("UOM updated successfully"),
		Data: entityToProto(entity),
	}, nil
}

// DeleteUOM deletes a Unit of Measure by code
func (h *UOMHandler) DeleteUOM(ctx context.Context, req *pb.DeleteUOMRequest) (*pb.DeleteUOMResponse, error) {
	cmd := appuom.DeleteCommand{UOMCode: req.UomCode}

	err := h.deleteHandler.Handle(ctx, cmd)
	if err != nil {
		return &pb.DeleteUOMResponse{
			Base: errorToBaseResponse(err),
		}, nil
	}

	return &pb.DeleteUOMResponse{
		Base: successResponse("UOM deleted successfully"),
	}, nil
}

// Helper functions

func pbCategoryToString(cat pb.UOMCategory) string {
	switch cat {
	case pb.UOMCategory_UOM_CATEGORY_WEIGHT:
		return "WEIGHT"
	case pb.UOMCategory_UOM_CATEGORY_VOLUME:
		return "VOLUME"
	case pb.UOMCategory_UOM_CATEGORY_QUANTITY:
		return "QUANTITY"
	case pb.UOMCategory_UOM_CATEGORY_LENGTH:
		return "LENGTH"
	default:
		return ""
	}
}

func stringToPbCategory(cat string) pb.UOMCategory {
	switch cat {
	case "WEIGHT":
		return pb.UOMCategory_UOM_CATEGORY_WEIGHT
	case "VOLUME":
		return pb.UOMCategory_UOM_CATEGORY_VOLUME
	case "QUANTITY":
		return pb.UOMCategory_UOM_CATEGORY_QUANTITY
	case "LENGTH":
		return pb.UOMCategory_UOM_CATEGORY_LENGTH
	default:
		return pb.UOMCategory_UOM_CATEGORY_UNSPECIFIED
	}
}

func entityToProto(entity *uom.UOM) *pb.UOM {
	audit := &pb.AuditInfo{
		CreatedAt: entity.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy: entity.CreatedBy(),
	}
	if entity.UpdatedAt() != nil {
		updatedAt := entity.UpdatedAt().Format("2006-01-02T15:04:05Z07:00")
		audit.UpdatedAt = &updatedAt
	}
	if entity.UpdatedBy() != nil {
		audit.UpdatedBy = entity.UpdatedBy()
	}

	return &pb.UOM{
		UomCode:     entity.Code().String(),
		UomName:     entity.Name(),
		UomCategory: stringToPbCategory(entity.Category().String()),
		IsBaseUom:   entity.IsBaseUOM(),
		Audit:       audit,
	}
}

func successResponse(message string) *pb.BaseResponse {
	return &pb.BaseResponse{
		StatusCode: "200",
		IsSuccess:  true,
		Message:    message,
	}
}

func errorToBaseResponse(err error) *pb.BaseResponse {
	statusCode := "500"
	message := "Internal server error"

	switch {
	case errors.Is(err, uom.ErrNotFound):
		statusCode = "404"
		message = err.Error()
	case errors.Is(err, uom.ErrAlreadyExists):
		statusCode = "409"
		message = err.Error()
	case errors.Is(err, uom.ErrInvalidUOMCode),
		errors.Is(err, uom.ErrInvalidCategory),
		errors.Is(err, uom.ErrEmptyName):
		statusCode = "400"
		message = err.Error()
	}

	return &pb.BaseResponse{
		StatusCode: statusCode,
		IsSuccess:  false,
		Message:    message,
	}
}

func mapDomainErrorToGRPC(err error) error {
	switch {
	case errors.Is(err, uom.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, uom.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, uom.ErrInvalidUOMCode),
		errors.Is(err, uom.ErrInvalidCategory),
		errors.Is(err, uom.ErrEmptyName):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
