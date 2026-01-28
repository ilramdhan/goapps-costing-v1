# Development Rules & Guidelines
## Yarn Costing gRPC Microservice

**Version**: 1.0  
**Last Updated**: 2026-01-28

---

## Table of Contents
1. [Architecture Rules](#1-architecture-rules)
2. [Project Structure](#2-project-structure)
3. [Proto & gRPC Rules](#3-proto--grpc-rules)
4. [Domain Layer Rules (DDD)](#4-domain-layer-rules-ddd)
5. [Application Layer Rules](#5-application-layer-rules)
6. [Infrastructure Layer Rules](#6-infrastructure-layer-rules)
7. [Error Handling](#7-error-handling)
8. [Response Format](#8-response-format)
9. [Database Rules](#9-database-rules)
10. [Testing Rules](#10-testing-rules)
11. [Git & Commit Rules](#11-git--commit-rules)
12. [Code Style](#12-code-style)

---

## 1. Architecture Rules

### 1.1 Layer Dependencies

```
┌─────────────────────────────────────────────────────────────────┐
│  Delivery Layer (gRPC Handlers, HTTP)                          │
│    ↓ Can depend on: Application, Domain, Proto                  │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│  Application Layer (Use Cases, Commands, Queries)               │
│    ↓ Can depend on: Domain only                                  │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│  Domain Layer (Entities, Value Objects, Repository Interfaces)  │
│    ✗ CANNOT depend on any other layer                           │
└─────────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────────┐
│  Infrastructure Layer (Database, External Services)             │
│    ↑ Implements interfaces from Domain Layer                     │
└─────────────────────────────────────────────────────────────────┘
```

**Rules:**
- ❌ Domain layer MUST NOT import from other layers
- ❌ Application layer MUST NOT import from Infrastructure or Delivery
- ✅ Infrastructure implements interfaces defined in Domain
- ✅ Delivery calls Application layer, never directly Infrastructure

### 1.2 Bounded Context Naming

Each bounded context follows this package structure:
```
internal/
├── domain/{context}/          # Domain layer
├── application/{context}/     # Application layer
├── infrastructure/{context}/  # Infrastructure layer (if specific)
└── delivery/grpc/             # Shared delivery layer
```

---

## 2. Project Structure

### 2.1 File Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Go files | `snake_case.go` | `uom_repository.go` |
| Proto files | `snake_case.proto` | `uom_service.proto` |
| Migration files | `{version}_{description}.{up/down}.sql` | `000001_create_uom.up.sql` |
| Test files | `{name}_test.go` | `uom_repository_test.go` |

### 2.2 Package Naming

- Use short, lowercase single-word names
- Avoid underscores or camelCase
- Bad: `uom_service`, `uomService`
- Good: `uom`, `parameter`

### 2.3 Interface Naming

- Suffix interfaces with `-er` when describing behavior
- Repository interfaces: `{Entity}Repository`
- Service interfaces: `{Entity}Service`

```go
// Good
type UOMRepository interface {
    Create(ctx context.Context, uom *UOM) error
}

// Good for behavior
type Validator interface {
    Validate() error
}
```

---

## 3. Proto & gRPC Rules

### 3.1 Proto File Organization

```protobuf
// File: proto/costing/v1/uom.proto

syntax = "proto3";

package costing.v1;

option go_package = "github.com/your-org/goapps-costing-v1/gen/go/costing/v1;costingv1";

import "buf/validate/validate.proto";
import "google/api/annotations.proto";
import "costing/v1/common.proto";

// 1. Service definition first
service UOMService {
  rpc CreateUOM(CreateUOMRequest) returns (CreateUOMResponse) {
    option (google.api.http) = {
      post: "/v1/uoms"
      body: "*"
    };
  }
}

// 2. Request messages
message CreateUOMRequest {
  string uom_code = 1 [(buf.validate.field).string = {
    min_len: 1,
    max_len: 20,
    pattern: "^[A-Z0-9_]+$"
  }];
  // ...
}

// 3. Response messages
message CreateUOMResponse {
  BaseResponse base = 1;
  UOM data = 2;
}

// 4. Entity messages
message UOM {
  string uom_code = 1;
  string uom_name = 2;
  // ...
}
```

### 3.2 Validation Rules (buf.validate)

**Always use buf.validate for:**
- Required fields
- String length/pattern
- Numeric ranges
- Enum values

```protobuf
message CreateParameterRequest {
  // Required string with pattern
  string parameter_code = 1 [(buf.validate.field).string = {
    min_len: 1,
    max_len: 50,
    pattern: "^[A-Z][A-Z0-9_]*$"
  }];

  // Required enum
  ParameterCategory category = 2 [(buf.validate.field).enum = {
    defined_only: true,
    not_in: [0]  // Exclude UNSPECIFIED
  }];

  // Optional numeric with range
  optional double min_value = 3 [(buf.validate.field).double = {
    gte: -999999999,
    lte: 999999999
  }];
}
```

### 3.3 Service Method Naming

| Operation | Method Name | HTTP | URL Pattern |
|-----------|-------------|------|-------------|
| Create | `Create{Entity}` | POST | `/v1/{entities}` |
| Get by ID | `Get{Entity}` | GET | `/v1/{entities}/{id}` |
| List | `List{Entities}` | GET | `/v1/{entities}` |
| Update | `Update{Entity}` | PUT | `/v1/{entities}/{id}` |
| Delete | `Delete{Entity}` | DELETE | `/v1/{entities}/{id}` |

### 3.4 Proto Generation Command

```bash
# Always use Makefile
make proto

# Never run protoc directly in development
# ❌ protoc --go_out=...
```

---

## 4. Domain Layer Rules (DDD)

### 4.1 Entity Structure

```go
// internal/domain/uom/entity.go

package uom

import (
    "time"
)

// UOM is the aggregate root for Unit of Measure
type UOM struct {
    code       UOMCode      // Value Object
    name       string
    category   UOMCategory  // Value Object
    isBaseUOM  bool
    createdAt  time.Time
    createdBy  string
    updatedAt  *time.Time
    updatedBy  *string
}

// NewUOM creates a new UOM with validation
func NewUOM(code UOMCode, name string, category UOMCategory, createdBy string) (*UOM, error) {
    if name == "" {
        return nil, ErrEmptyName
    }
    if createdBy == "" {
        return nil, ErrEmptyCreatedBy
    }

    return &UOM{
        code:      code,
        name:      name,
        category:  category,
        isBaseUOM: false,
        createdAt: time.Now(),
        createdBy: createdBy,
    }, nil
}

// Getters - expose internal state read-only
func (u *UOM) Code() UOMCode     { return u.code }
func (u *UOM) Name() string      { return u.name }
func (u *UOM) Category() UOMCategory { return u.category }

// Domain behavior methods
func (u *UOM) SetAsBaseUOM() {
    u.isBaseUOM = true
}

func (u *UOM) UpdateName(name string, updatedBy string) error {
    if name == "" {
        return ErrEmptyName
    }
    u.name = name
    now := time.Now()
    u.updatedAt = &now
    u.updatedBy = &updatedBy
    return nil
}
```

### 4.2 Value Object Rules

```go
// internal/domain/uom/value_object.go

package uom

import (
    "regexp"
)

// UOMCode is a value object for UOM identifier
type UOMCode string

var uomCodePattern = regexp.MustCompile(`^[A-Z0-9_]{1,20}$`)

func NewUOMCode(code string) (UOMCode, error) {
    if !uomCodePattern.MatchString(code) {
        return "", ErrInvalidUOMCode
    }
    return UOMCode(code), nil
}

func (c UOMCode) String() string {
    return string(c)
}

// UOMCategory represents the category of UOM
type UOMCategory string

const (
    CategoryWeight   UOMCategory = "WEIGHT"
    CategoryVolume   UOMCategory = "VOLUME"
    CategoryQuantity UOMCategory = "QUANTITY"
    CategoryLength   UOMCategory = "LENGTH"
)

func NewUOMCategory(category string) (UOMCategory, error) {
    switch UOMCategory(category) {
    case CategoryWeight, CategoryVolume, CategoryQuantity, CategoryLength:
        return UOMCategory(category), nil
    default:
        return "", ErrInvalidCategory
    }
}
```

### 4.3 Repository Interface in Domain

```go
// internal/domain/uom/repository.go

package uom

import "context"

// Repository defines the interface for UOM persistence
// This interface is defined in domain, implemented in infrastructure
type Repository interface {
    Create(ctx context.Context, uom *UOM) error
    GetByCode(ctx context.Context, code UOMCode) (*UOM, error)
    List(ctx context.Context, filter ListFilter) ([]*UOM, int64, error)
    Update(ctx context.Context, uom *UOM) error
    Delete(ctx context.Context, code UOMCode) error
    ExistsByCode(ctx context.Context, code UOMCode) (bool, error)
}

// ListFilter contains filtering options for listing UOMs
type ListFilter struct {
    Category *UOMCategory
    Page     int
    PageSize int
}
```

### 4.4 Domain Errors

```go
// internal/domain/uom/errors.go

package uom

import "errors"

var (
    ErrNotFound        = errors.New("uom not found")
    ErrAlreadyExists   = errors.New("uom already exists")
    ErrEmptyName       = errors.New("uom name cannot be empty")
    ErrEmptyCreatedBy  = errors.New("created_by cannot be empty")
    ErrInvalidUOMCode  = errors.New("invalid uom code format")
    ErrInvalidCategory = errors.New("invalid uom category")
)
```

---

## 5. Application Layer Rules

### 5.1 Command Handler Structure

```go
// internal/application/uom/commands.go

package uom

import (
    "context"

    "github.com/your-org/goapps-costing-v1/internal/domain/uom"
)

// CreateUOMCommand represents the create UOM command
type CreateUOMCommand struct {
    UOMCode   string
    UOMName   string
    Category  string
    CreatedBy string
}

// CreateUOMHandler handles the CreateUOM command
type CreateUOMHandler struct {
    repo uom.Repository
}

func NewCreateUOMHandler(repo uom.Repository) *CreateUOMHandler {
    return &CreateUOMHandler{repo: repo}
}

func (h *CreateUOMHandler) Handle(ctx context.Context, cmd CreateUOMCommand) (*uom.UOM, error) {
    // 1. Validate and create value objects
    code, err := uom.NewUOMCode(cmd.UOMCode)
    if err != nil {
        return nil, err
    }

    category, err := uom.NewUOMCategory(cmd.Category)
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

    // 4. Persist
    if err := h.repo.Create(ctx, entity); err != nil {
        return nil, err
    }

    return entity, nil
}
```

### 5.2 Query Handler Structure

```go
// internal/application/uom/queries.go

package uom

import (
    "context"

    "github.com/your-org/goapps-costing-v1/internal/domain/uom"
)

// ListUOMsQuery represents the list UOMs query
type ListUOMsQuery struct {
    Category *string
    Page     int
    PageSize int
}

// ListUOMsHandler handles the ListUOMs query
type ListUOMsHandler struct {
    repo uom.Repository
}

func NewListUOMsHandler(repo uom.Repository) *ListUOMsHandler {
    return &ListUOMsHandler{repo: repo}
}

type ListUOMsResult struct {
    UOMs  []*uom.UOM
    Total int64
}

func (h *ListUOMsHandler) Handle(ctx context.Context, query ListUOMsQuery) (*ListUOMsResult, error) {
    filter := uom.ListFilter{
        Page:     query.Page,
        PageSize: query.PageSize,
    }

    if query.Category != nil {
        cat, err := uom.NewUOMCategory(*query.Category)
        if err != nil {
            return nil, err
        }
        filter.Category = &cat
    }

    uoms, total, err := h.repo.List(ctx, filter)
    if err != nil {
        return nil, err
    }

    return &ListUOMsResult{
        UOMs:  uoms,
        Total: total,
    }, nil
}
```

---

## 6. Infrastructure Layer Rules

### 6.1 Repository Implementation

```go
// internal/infrastructure/postgres/uom_repository.go

package postgres

import (
    "context"
    "database/sql"
    "errors"

    "github.com/your-org/goapps-costing-v1/internal/domain/uom"
)

type UOMRepository struct {
    db *sql.DB
}

func NewUOMRepository(db *sql.DB) *UOMRepository {
    return &UOMRepository{db: db}
}

// Verify interface implementation at compile time
var _ uom.Repository = (*UOMRepository)(nil)

func (r *UOMRepository) Create(ctx context.Context, entity *uom.UOM) error {
    query := `
        INSERT INTO mst_uom (uom_code, uom_name, uom_category, is_base_uom, created_at, created_by)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

    _, err := r.db.ExecContext(ctx, query,
        entity.Code().String(),
        entity.Name(),
        string(entity.Category()),
        entity.IsBaseUOM(),
        entity.CreatedAt(),
        entity.CreatedBy(),
    )

    return err
}

func (r *UOMRepository) GetByCode(ctx context.Context, code uom.UOMCode) (*uom.UOM, error) {
    query := `
        SELECT uom_code, uom_name, uom_category, is_base_uom, created_at, created_by, updated_at, updated_by
        FROM mst_uom
        WHERE uom_code = $1
    `

    var dto uomDTO
    err := r.db.QueryRowContext(ctx, query, code.String()).Scan(
        &dto.UOMCode,
        &dto.UOMName,
        &dto.UOMCategory,
        &dto.IsBaseUOM,
        &dto.CreatedAt,
        &dto.CreatedBy,
        &dto.UpdatedAt,
        &dto.UpdatedBy,
    )

    if errors.Is(err, sql.ErrNoRows) {
        return nil, uom.ErrNotFound
    }
    if err != nil {
        return nil, err
    }

    return dto.ToEntity()
}
```

### 6.2 Database Connection Rules

```go
// Always use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Always use prepared statements or parameterized queries
// ❌ Bad
query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", id)

// ✅ Good
query := "SELECT * FROM users WHERE id = $1"
row := db.QueryRowContext(ctx, query, id)
```

---

## 7. Error Handling

### 7.1 Error Types

| Layer | Error Handling |
|-------|----------------|
| Domain | Return domain-specific errors |
| Application | Wrap with context, return domain errors |
| Infrastructure | Wrap DB errors, translate to domain errors |
| Delivery | Map to gRPC status codes |

### 7.2 gRPC Error Mapping

```go
// internal/delivery/grpc/errors.go

package grpc

import (
    "errors"

    "github.com/your-org/goapps-costing-v1/internal/domain/uom"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func mapDomainError(err error) error {
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
```

---

## 8. Response Format

### 8.1 Standard Response Structure

All responses MUST include `BaseResponse`:

```json
{
  "base": {
    "validation_errors": [],
    "status_code": "200",
    "is_success": true,
    "message": "UOM created successfully"
  },
  "data": {
    "uom_code": "KG",
    "uom_name": "Kilogram",
    "uom_category": "WEIGHT"
  }
}
```

### 8.2 Validation Error Response

```json
{
  "base": {
    "validation_errors": [
      {
        "field": "uom_code",
        "message": "value length must be at least 1 characters"
      },
      {
        "field": "uom_category",
        "message": "value must be one of [WEIGHT, VOLUME, QUANTITY, LENGTH]"
      }
    ],
    "status_code": "400",
    "is_success": false,
    "message": "Validation failed"
  }
}
```

### 8.3 Response Builder

```go
// pkg/response/wrapper.go

package response

import (
    commonv1 "github.com/your-org/goapps-costing-v1/gen/go/costing/v1"
)

func Success(message string) *commonv1.BaseResponse {
    return &commonv1.BaseResponse{
        ValidationErrors: nil,
        StatusCode:       "200",
        IsSuccess:        true,
        Message:          message,
    }
}

func Created(message string) *commonv1.BaseResponse {
    return &commonv1.BaseResponse{
        ValidationErrors: nil,
        StatusCode:       "201",
        IsSuccess:        true,
        Message:          message,
    }
}

func ValidationFailed(errors []*commonv1.ValidationError) *commonv1.BaseResponse {
    return &commonv1.BaseResponse{
        ValidationErrors: errors,
        StatusCode:       "400",
        IsSuccess:        false,
        Message:          "Validation failed",
    }
}

func NotFound(message string) *commonv1.BaseResponse {
    return &commonv1.BaseResponse{
        ValidationErrors: nil,
        StatusCode:       "404",
        IsSuccess:        false,
        Message:          message,
    }
}
```

---

## 9. Database Rules

### 9.1 Table Naming

- Use `snake_case`
- Prefix with module identifier: `mst_` (master), `cst_` (costing), `wfl_` (workflow)
- Examples: `mst_uom`, `mst_parameter`, `cst_product_cost`

### 9.2 Column Naming

- Use `snake_case`
- Boolean columns: prefix with `is_`, `has_`, `can_`
- Timestamp columns: suffix with `_at` (e.g., `created_at`, `updated_at`, `deleted_at`)
- User audit columns: `created_by`, `updated_by`, `deleted_by`

### 9.3 Migration File Naming

```
{version}_{description}.{direction}.sql

Examples:
000001_create_mst_uom.up.sql
000001_create_mst_uom.down.sql
000002_create_mst_parameter.up.sql
000002_create_mst_parameter.down.sql
```

### 9.4 Migration Rules

```sql
-- Always include IF NOT EXISTS for CREATE
CREATE TABLE IF NOT EXISTS mst_uom (
    -- ...
);

-- Always include IF EXISTS for DROP
DROP TABLE IF EXISTS mst_uom;

-- Always create indexes
CREATE INDEX IF NOT EXISTS idx_mst_uom_category ON mst_uom(uom_category);

-- Always include audit columns
created_at TIMESTAMPTZ DEFAULT NOW(),
created_by VARCHAR(100),
updated_at TIMESTAMPTZ,
updated_by VARCHAR(100)
```

---

## 10. Testing Rules

### 10.1 Test File Location

Tests are placed next to the code they test:
```
internal/domain/uom/
├── entity.go
├── entity_test.go       # Unit tests for entity
├── repository.go
└── repository_test.go   # Mock tests for repository interface
```

### 10.2 Test Naming

```go
func TestUOM_NewUOM_ValidInput_ReturnsUOM(t *testing.T) {}
func TestUOM_NewUOM_EmptyName_ReturnsError(t *testing.T) {}
func TestUOMRepository_Create_Success(t *testing.T) {}
func TestUOMRepository_GetByCode_NotFound_ReturnsError(t *testing.T) {}
```

Pattern: `Test{Type}_{Method}_{Scenario}_{Expected}`

### 10.3 Table-Driven Tests

```go
func TestNewUOMCode(t *testing.T) {
    tests := []struct {
        name    string
        code    string
        wantErr error
    }{
        {
            name:    "valid uppercase code",
            code:    "KG",
            wantErr: nil,
        },
        {
            name:    "valid code with underscore",
            code:    "CUBIC_M",
            wantErr: nil,
        },
        {
            name:    "invalid lowercase",
            code:    "kg",
            wantErr: ErrInvalidUOMCode,
        },
        {
            name:    "empty code",
            code:    "",
            wantErr: ErrInvalidUOMCode,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := NewUOMCode(tt.code)
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("NewUOMCode() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 10.4 Coverage Requirements

| Layer | Minimum Coverage |
|-------|------------------|
| Domain | 90% |
| Application | 85% |
| Infrastructure | 70% |
| Delivery | 60% |

---

## 11. Git & Commit Rules

### 11.1 Branch Naming

```
feature/{ticket-id}-{short-description}
bugfix/{ticket-id}-{short-description}
hotfix/{ticket-id}-{short-description}

Examples:
feature/COST-123-add-uom-service
bugfix/COST-456-fix-validation-error
```

### 11.2 Commit Message Format

```
{type}({scope}): {subject}

{body}

{footer}
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `refactor`: Refactoring
- `test`: Adding tests
- `chore`: Maintenance

**Examples:**
```
feat(uom): add CreateUOM gRPC endpoint

- Implement domain entity with validation
- Add repository interface and postgres implementation
- Add gRPC handler with protovalidate

Refs: COST-123

---

fix(parameter): correct validation for min_value

min_value was allowing negative values which is incorrect
for this business domain.

Fixes: COST-456
```

---

## 12. Code Style

### 12.1 Import Ordering

```go
import (
    // 1. Standard library
    "context"
    "errors"
    "time"

    // 2. Third-party packages
    "github.com/rs/zerolog"
    "google.golang.org/grpc"

    // 3. Internal packages
    "github.com/your-org/goapps-costing-v1/internal/domain/uom"
    "github.com/your-org/goapps-costing-v1/pkg/response"
)
```

### 12.2 Variable Naming

```go
// Use short variable names in short scopes
for i, u := range uoms {
    // ...
}

// Use descriptive names for longer scopes
func (h *CreateUOMHandler) Handle(ctx context.Context, cmd CreateUOMCommand) (*uom.UOM, error) {
    existingUOM, err := h.repo.GetByCode(ctx, cmd.UOMCode)
    // ...
}

// Avoid single-letter names in package scope
// ❌ Bad
var u *UOM

// ✅ Good
var defaultUOM *UOM
```

### 12.3 Constants

```go
// Group related constants
const (
    MaxUOMCodeLength      = 20
    MaxUOMNameLength      = 100
    MaxParameterCodeLength = 50
)

// Use typed constants for enums
type UOMCategory string

const (
    CategoryWeight   UOMCategory = "WEIGHT"
    CategoryVolume   UOMCategory = "VOLUME"
    CategoryQuantity UOMCategory = "QUANTITY"
    CategoryLength   UOMCategory = "LENGTH"
)
```

### 12.4 Error Wrapping

```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create UOM: %w", err)
}

// Use errors.Is for comparison
if errors.Is(err, uom.ErrNotFound) {
    // handle not found
}
```

---

## Quick Reference Checklist

Before submitting PR, ensure:

- [ ] All tests pass (`make test`)
- [ ] Code coverage meets requirements
- [ ] Proto changes regenerated (`make proto`)
- [ ] Migrations tested up and down
- [ ] No lint errors (`make lint`)
- [ ] Commit messages follow convention
- [ ] Documentation updated if needed
- [ ] No hardcoded values (use config)
- [ ] Context passed through all layers
- [ ] Error handling follows guidelines
- [ ] Response format matches standard

---

**Maintainers**: DevOps Team  
**Questions**: Create issue with label `question`
