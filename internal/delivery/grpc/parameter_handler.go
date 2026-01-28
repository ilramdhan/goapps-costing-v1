package grpc

import (
	"context"
	"errors"

	pb "github.com/homindolenern/goapps-costing-v1/gen/go/costing/v1"
	appparam "github.com/homindolenern/goapps-costing-v1/internal/application/parameter"
	"github.com/homindolenern/goapps-costing-v1/internal/domain/parameter"
)

// ParameterHandler implements the gRPC ParameterService
type ParameterHandler struct {
	pb.UnimplementedParameterServiceServer
	createHandler *appparam.CreateHandler
	updateHandler *appparam.UpdateHandler
	deleteHandler *appparam.DeleteHandler
	getHandler    *appparam.GetHandler
	listHandler   *appparam.ListHandler
	validator     *ValidationHelper
}

// NewParameterHandler creates a new Parameter handler
func NewParameterHandler(
	createHandler *appparam.CreateHandler,
	updateHandler *appparam.UpdateHandler,
	deleteHandler *appparam.DeleteHandler,
	getHandler *appparam.GetHandler,
	listHandler *appparam.ListHandler,
	validator *ValidationHelper,
) *ParameterHandler {
	return &ParameterHandler{
		createHandler: createHandler,
		updateHandler: updateHandler,
		deleteHandler: deleteHandler,
		getHandler:    getHandler,
		listHandler:   listHandler,
		validator:     validator,
	}
}

// CreateParameter creates a new Parameter
func (h *ParameterHandler) CreateParameter(ctx context.Context, req *pb.CreateParameterRequest) (*pb.CreateParameterResponse, error) {
	// Validate request
	if validationResp := h.validator.Validate(ctx, req); validationResp != nil {
		return &pb.CreateParameterResponse{Base: validationResp}, nil
	}

	cmd := appparam.CreateCommand{
		ParameterCode: req.ParameterCode,
		ParameterName: req.ParameterName,
		Category:      pbParamCategoryToString(req.ParameterCategory),
		DataType:      pbDataTypeToString(req.DataType),
		UOM:           req.Uom,
		MinValue:      req.MinValue,
		MaxValue:      req.MaxValue,
		AllowedValues: req.AllowedValues,
		IsMandatory:   req.IsMandatory,
		Description:   req.Description,
		CreatedBy:     "system", // TODO: Extract from context/auth
	}

	entity, err := h.createHandler.Handle(ctx, cmd)
	if err != nil {
		return &pb.CreateParameterResponse{
			Base: paramErrorToBaseResponse(err),
		}, nil
	}

	return &pb.CreateParameterResponse{
		Base: paramSuccessResponse("Parameter created successfully"),
		Data: paramEntityToProto(entity),
	}, nil
}

// GetParameter retrieves a Parameter by code
func (h *ParameterHandler) GetParameter(ctx context.Context, req *pb.GetParameterRequest) (*pb.GetParameterResponse, error) {
	query := appparam.GetQuery{ParameterCode: req.ParameterCode}

	entity, err := h.getHandler.Handle(ctx, query)
	if err != nil {
		return &pb.GetParameterResponse{
			Base: paramErrorToBaseResponse(err),
		}, nil
	}

	return &pb.GetParameterResponse{
		Base: paramSuccessResponse("Parameter retrieved successfully"),
		Data: paramEntityToProto(entity),
	}, nil
}

// ListParameters retrieves a paginated list of Parameters
func (h *ParameterHandler) ListParameters(ctx context.Context, req *pb.ListParametersRequest) (*pb.ListParametersResponse, error) {
	query := appparam.ListQuery{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	if req.Category != nil && *req.Category != pb.ParameterCategory_PARAMETER_CATEGORY_UNSPECIFIED {
		cat := pbParamCategoryToString(*req.Category)
		query.Category = &cat
	}

	if req.IsActive != nil {
		query.IsActive = req.IsActive
	}

	result, err := h.listHandler.Handle(ctx, query)
	if err != nil {
		return &pb.ListParametersResponse{
			Base: paramErrorToBaseResponse(err),
		}, nil
	}

	data := make([]*pb.Parameter, len(result.Parameters))
	for i, entity := range result.Parameters {
		data[i] = paramEntityToProto(entity)
	}

	totalPages := int32(result.Total) / req.PageSize
	if int32(result.Total)%req.PageSize > 0 {
		totalPages++
	}

	return &pb.ListParametersResponse{
		Base: paramSuccessResponse("Parameters retrieved successfully"),
		Data: data,
		Pagination: &pb.PaginationMeta{
			CurrentPage: req.Page,
			PageSize:    req.PageSize,
			TotalItems:  result.Total,
			TotalPages:  totalPages,
		},
	}, nil
}

// UpdateParameter updates an existing Parameter
func (h *ParameterHandler) UpdateParameter(ctx context.Context, req *pb.UpdateParameterRequest) (*pb.UpdateParameterResponse, error) {
	cmd := appparam.UpdateCommand{
		ParameterCode: req.ParameterCode,
		ParameterName: req.ParameterName,
		Category:      pbParamCategoryToString(req.ParameterCategory),
		DataType:      pbDataTypeToString(req.DataType),
		UOM:           req.Uom,
		MinValue:      req.MinValue,
		MaxValue:      req.MaxValue,
		AllowedValues: req.AllowedValues,
		IsMandatory:   req.IsMandatory,
		Description:   req.Description,
		IsActive:      req.IsActive,
		UpdatedBy:     "system", // TODO: Extract from context/auth
	}

	entity, err := h.updateHandler.Handle(ctx, cmd)
	if err != nil {
		return &pb.UpdateParameterResponse{
			Base: paramErrorToBaseResponse(err),
		}, nil
	}

	return &pb.UpdateParameterResponse{
		Base: paramSuccessResponse("Parameter updated successfully"),
		Data: paramEntityToProto(entity),
	}, nil
}

// DeleteParameter deletes a Parameter by code
func (h *ParameterHandler) DeleteParameter(ctx context.Context, req *pb.DeleteParameterRequest) (*pb.DeleteParameterResponse, error) {
	cmd := appparam.DeleteCommand{ParameterCode: req.ParameterCode}

	err := h.deleteHandler.Handle(ctx, cmd)
	if err != nil {
		return &pb.DeleteParameterResponse{
			Base: paramErrorToBaseResponse(err),
		}, nil
	}

	return &pb.DeleteParameterResponse{
		Base: paramSuccessResponse("Parameter deleted successfully"),
	}, nil
}

// Helper functions

func pbParamCategoryToString(cat pb.ParameterCategory) string {
	switch cat {
	case pb.ParameterCategory_PARAMETER_CATEGORY_MACHINE:
		return "MACHINE"
	case pb.ParameterCategory_PARAMETER_CATEGORY_MATERIAL:
		return "MATERIAL"
	case pb.ParameterCategory_PARAMETER_CATEGORY_QUALITY:
		return "QUALITY"
	case pb.ParameterCategory_PARAMETER_CATEGORY_OUTPUT:
		return "OUTPUT"
	case pb.ParameterCategory_PARAMETER_CATEGORY_PROCESS:
		return "PROCESS"
	default:
		return ""
	}
}

func stringToPbParamCategory(cat string) pb.ParameterCategory {
	switch cat {
	case "MACHINE":
		return pb.ParameterCategory_PARAMETER_CATEGORY_MACHINE
	case "MATERIAL":
		return pb.ParameterCategory_PARAMETER_CATEGORY_MATERIAL
	case "QUALITY":
		return pb.ParameterCategory_PARAMETER_CATEGORY_QUALITY
	case "OUTPUT":
		return pb.ParameterCategory_PARAMETER_CATEGORY_OUTPUT
	case "PROCESS":
		return pb.ParameterCategory_PARAMETER_CATEGORY_PROCESS
	default:
		return pb.ParameterCategory_PARAMETER_CATEGORY_UNSPECIFIED
	}
}

func pbDataTypeToString(dt pb.ParameterDataType) string {
	switch dt {
	case pb.ParameterDataType_PARAMETER_DATA_TYPE_NUMERIC:
		return "NUMERIC"
	case pb.ParameterDataType_PARAMETER_DATA_TYPE_TEXT:
		return "TEXT"
	case pb.ParameterDataType_PARAMETER_DATA_TYPE_BOOLEAN:
		return "BOOLEAN"
	case pb.ParameterDataType_PARAMETER_DATA_TYPE_DROPDOWN:
		return "DROPDOWN"
	default:
		return ""
	}
}

func stringToPbDataType(dt string) pb.ParameterDataType {
	switch dt {
	case "NUMERIC":
		return pb.ParameterDataType_PARAMETER_DATA_TYPE_NUMERIC
	case "TEXT":
		return pb.ParameterDataType_PARAMETER_DATA_TYPE_TEXT
	case "BOOLEAN":
		return pb.ParameterDataType_PARAMETER_DATA_TYPE_BOOLEAN
	case "DROPDOWN":
		return pb.ParameterDataType_PARAMETER_DATA_TYPE_DROPDOWN
	default:
		return pb.ParameterDataType_PARAMETER_DATA_TYPE_UNSPECIFIED
	}
}

func paramEntityToProto(entity *parameter.Parameter) *pb.Parameter {
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

	return &pb.Parameter{
		ParameterCode:     entity.Code().String(),
		ParameterName:     entity.Name(),
		ParameterCategory: stringToPbParamCategory(entity.Category().String()),
		DataType:          stringToPbDataType(entity.DataType().String()),
		Uom:               entity.UOM(),
		MinValue:          entity.MinValue(),
		MaxValue:          entity.MaxValue(),
		AllowedValues:     entity.AllowedValues(),
		IsMandatory:       entity.IsMandatory(),
		Description:       entity.Description(),
		IsActive:          entity.IsActive(),
		Audit:             audit,
	}
}

func paramSuccessResponse(message string) *pb.BaseResponse {
	return &pb.BaseResponse{
		StatusCode: "200",
		IsSuccess:  true,
		Message:    message,
	}
}

func paramErrorToBaseResponse(err error) *pb.BaseResponse {
	statusCode := "500"
	message := "Internal server error"

	switch {
	case errors.Is(err, parameter.ErrNotFound):
		statusCode = "404"
		message = err.Error()
	case errors.Is(err, parameter.ErrAlreadyExists):
		statusCode = "409"
		message = err.Error()
	case errors.Is(err, parameter.ErrInvalidCode),
		errors.Is(err, parameter.ErrInvalidCategory),
		errors.Is(err, parameter.ErrInvalidDataType),
		errors.Is(err, parameter.ErrEmptyName),
		errors.Is(err, parameter.ErrMinGreaterThanMax),
		errors.Is(err, parameter.ErrDropdownNoOptions):
		statusCode = "400"
		message = err.Error()
	}

	return &pb.BaseResponse{
		StatusCode: statusCode,
		IsSuccess:  false,
		Message:    message,
	}
}
