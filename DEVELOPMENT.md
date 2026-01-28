# Development Guide

Panduan lengkap untuk developer yang ingin berkontribusi di project **goapps-costing-v1**. Dokumen ini menjelaskan arsitektur, workflow, dan step-by-step untuk menambahkan service baru.

---

## Table of Contents
1. [Architecture Overview](#1-architecture-overview)
2. [Prerequisites & Setup](#2-prerequisites--setup)
3. [Project Structure](#3-project-structure)
4. [Development Workflow](#4-development-workflow)
5. [Step-by-Step: Creating a New Service](#5-step-by-step-creating-a-new-service)
6. [Testing Guide](#6-testing-guide)
7. [Code Standards](#7-code-standards)
8. [Git Workflow](#8-git-workflow)
9. [CI/CD & Deployment](#9-cicd--deployment)
10. [Troubleshooting](#10-troubleshooting)

---

## 1. Architecture Overview

Project ini menggunakan **Clean Architecture** dengan **Domain-Driven Design (DDD)** patterns:

```
┌─────────────────────────────────────────────────────────────┐
│                        DELIVERY LAYER                        │
│              (gRPC Handlers, HTTP Gateway)                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      APPLICATION LAYER                       │
│               (Use Cases, Command/Query Handlers)            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        DOMAIN LAYER                          │
│          (Entities, Value Objects, Repository Interface)     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    INFRASTRUCTURE LAYER                      │
│           (PostgreSQL, Redis, External Services)             │
└─────────────────────────────────────────────────────────────┘
```

### Key Principles
- **Domain layer is independent** - No external dependencies
- **Dependency Inversion** - Higher layers depend on abstractions
- **Command/Query Separation** - Write operations (Command) vs Read operations (Query)

---

## 2. Prerequisites & Setup

### Required Tools
```bash
# Go 1.24+
go version  # go1.24.x

# Buf CLI (for proto generation)
# Install: https://buf.build/docs/installation
buf --version

# Docker & Docker Compose
docker --version
docker compose version

# golangci-lint (for linting)
golangci-lint --version

# grpcurl (for gRPC testing)
grpcurl --version

# migrate (for database migrations)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Initial Setup
```bash
# 1. Clone repository
git clone https://github.com/ilramdhan/goapps-costing-v1.git
cd goapps-costing-v1

# 2. Install dependencies
go mod download

# 3. Start infrastructure (PostgreSQL, Redis)
docker compose -f deployments/docker-compose.yaml up -d

# 4. Run migrations
export DATABASE_URL="postgres://costing:costing123@localhost:5432/costing_master?sslmode=disable"
migrate -path migrations/postgres -database "$DATABASE_URL" up

# 5. Generate proto code (if needed)
buf generate

# 6. Run server
go run ./cmd/master-service
```

### Verify Setup
```bash
# Health check
curl http://localhost:8080/v1/health/live

# Swagger UI
open http://localhost:8080/swagger/
```

---

## 3. Project Structure

```
goapps-costing-v1/
├── cmd/
│   └── master-service/          # Entry point
│       └── main.go              # Wiring semua dependencies
│
├── proto/                       # Proto definitions
│   └── costing/v1/
│       ├── uom.proto            # UOM service messages
│       ├── parameter.proto      # Parameter service messages
│       └── common.proto         # Shared messages (BaseResponse, etc)
│
├── gen/
│   ├── go/costing/v1/           # Generated Go code from proto
│   └── openapi/                 # Generated Swagger spec
│
├── internal/                    # Private application code
│   ├── domain/                  # DOMAIN LAYER
│   │   ├── uom/
│   │   │   ├── entity.go        # UOM entity
│   │   │   ├── value_object.go  # UOMCode, Category value objects
│   │   │   ├── repository.go    # Repository interface
│   │   │   └── errors.go        # Domain errors
│   │   └── parameter/
│   │
│   ├── application/             # APPLICATION LAYER
│   │   ├── uom/
│   │   │   ├── create_handler.go
│   │   │   ├── update_handler.go
│   │   │   ├── delete_handler.go
│   │   │   ├── get_handler.go
│   │   │   └── list_handler.go
│   │   └── parameter/
│   │
│   ├── delivery/                # DELIVERY LAYER
│   │   ├── grpc/
│   │   │   ├── uom_handler.go
│   │   │   ├── parameter_handler.go
│   │   │   ├── health_handler.go
│   │   │   └── validation_helper.go
│   │   └── http/
│   │       └── error_handler.go
│   │
│   └── infrastructure/          # INFRASTRUCTURE LAYER
│       ├── postgres/
│       │   ├── connection.go
│       │   ├── uom_repository.go
│       │   └── parameter_repository.go
│       ├── redis/
│       └── cache/
│
├── pkg/                         # Public reusable packages
│   ├── errors/
│   ├── response/
│   ├── logger/
│   └── ratelimit/
│
├── migrations/
│   └── postgres/
│
├── tests/
│   └── integration/
│
└── deployments/
    ├── docker-compose.yaml
    └── k8s/
```

---

## 4. Development Workflow

### Alur Pengerjaan Service Baru

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  1. Proto   │───▶│  2. Domain  │───▶│ 3. Infra    │───▶│ 4. App      │
│  Definition │    │  Layer      │    │ (Repository)│    │ (Handlers)  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                                                                │
                                                                ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  8. Deploy  │◀───│  7. Test    │◀───│  6. Wire    │◀───│ 5. Delivery │
│  (CI/CD)    │    │             │    │  (main.go)  │    │ (gRPC)      │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

---

## 5. Step-by-Step: Creating a New Service

Contoh: Menambahkan **Material Service** untuk master data material.

### Step 1: Proto Definition

**File**: `proto/costing/v1/material.proto`

```protobuf
syntax = "proto3";

package costing.v1;

import "google/api/annotations.proto";
import "buf/validate/validate.proto";
import "costing/v1/common.proto";

option go_package = "github.com/homindolenern/goapps-costing-v1/gen/go/costing/v1;costingv1";

// ========================================
// MESSAGES
// ========================================

message Material {
  string material_code = 1;
  string material_name = 2;
  MaterialCategory category = 3;
  string uom_code = 4;  // Reference to UOM
  double unit_cost = 5;
  bool is_active = 6;
  AuditInfo audit = 7;
}

enum MaterialCategory {
  MATERIAL_CATEGORY_UNSPECIFIED = 0;
  MATERIAL_CATEGORY_RAW = 1;
  MATERIAL_CATEGORY_PACKAGING = 2;
  MATERIAL_CATEGORY_CONSUMABLE = 3;
}

// ========================================
// REQUEST/RESPONSE
// ========================================

message CreateMaterialRequest {
  string material_code = 1 [(buf.validate.field).string = {
    pattern: "^[A-Z][A-Z0-9_]*$",
    min_len: 2,
    max_len: 20
  }];
  string material_name = 2 [(buf.validate.field).string.min_len = 1];
  MaterialCategory category = 3 [(buf.validate.field).enum.not_in = [0]];
  string uom_code = 4 [(buf.validate.field).string.min_len = 1];
  double unit_cost = 5 [(buf.validate.field).double.gte = 0];
}

message CreateMaterialResponse {
  BaseResponse base = 1;
  Material data = 2;
}

// ... (GetMaterial, ListMaterials, UpdateMaterial, DeleteMaterial)

// ========================================
// SERVICE
// ========================================

service MaterialService {
  rpc CreateMaterial(CreateMaterialRequest) returns (CreateMaterialResponse) {
    option (google.api.http) = {
      post: "/v1/materials"
      body: "*"
    };
  }
  
  rpc GetMaterial(GetMaterialRequest) returns (GetMaterialResponse) {
    option (google.api.http) = {
      get: "/v1/materials/{material_code}"
    };
  }
  
  // ... other RPCs
}
```

**Generate Code**:
```bash
buf generate
# Output: gen/go/costing/v1/material.pb.go
#         gen/go/costing/v1/material_grpc.pb.go
#         gen/openapi/api.swagger.json (updated)
```

---

### Step 2: Domain Layer

Domain layer berisi business logic murni, tanpa dependency external.

#### 2a. Value Objects

**File**: `internal/domain/material/value_object.go`

```go
package material

import (
    "fmt"
    "regexp"
)

// MaterialCode adalah value object untuk material code
type MaterialCode struct {
    value string
}

var materialCodeRegex = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

func NewMaterialCode(code string) (MaterialCode, error) {
    if !materialCodeRegex.MatchString(code) {
        return MaterialCode{}, ErrInvalidMaterialCode
    }
    return MaterialCode{value: code}, nil
}

func (c MaterialCode) String() string { return c.value }

// Category adalah value object untuk kategori material
type Category struct {
    value string
}

var validCategories = map[string]bool{
    "RAW":        true,
    "PACKAGING":  true,
    "CONSUMABLE": true,
}

func NewCategory(cat string) (Category, error) {
    if !validCategories[cat] {
        return Category{}, ErrInvalidCategory
    }
    return Category{value: cat}, nil
}

func (c Category) String() string { return c.value }
```

#### 2b. Entity

**File**: `internal/domain/material/entity.go`

```go
package material

import "time"

// Material adalah aggregate root untuk master material
type Material struct {
    code      MaterialCode
    name      string
    category  Category
    uomCode   string
    unitCost  float64
    isActive  bool
    createdAt time.Time
    createdBy string
    updatedAt *time.Time
    updatedBy *string
}

// NewMaterial membuat entity Material baru
func NewMaterial(
    code MaterialCode,
    name string,
    category Category,
    uomCode string,
    unitCost float64,
    createdBy string,
) (*Material, error) {
    if name == "" {
        return nil, ErrEmptyName
    }
    if uomCode == "" {
        return nil, ErrEmptyUOMCode
    }
    if unitCost < 0 {
        return nil, ErrNegativeUnitCost
    }

    return &Material{
        code:      code,
        name:      name,
        category:  category,
        uomCode:   uomCode,
        unitCost:  unitCost,
        isActive:  true,
        createdAt: time.Now(),
        createdBy: createdBy,
    }, nil
}

// Getters
func (m *Material) Code() MaterialCode { return m.code }
func (m *Material) Name() string       { return m.name }
func (m *Material) Category() Category { return m.category }
func (m *Material) UOMCode() string    { return m.uomCode }
func (m *Material) UnitCost() float64  { return m.unitCost }
func (m *Material) IsActive() bool     { return m.isActive }
// ... other getters

// Update method
func (m *Material) Update(
    name string,
    category Category,
    uomCode string,
    unitCost float64,
    updatedBy string,
) error {
    if name == "" {
        return ErrEmptyName
    }
    
    m.name = name
    m.category = category
    m.uomCode = uomCode
    m.unitCost = unitCost
    now := time.Now()
    m.updatedAt = &now
    m.updatedBy = &updatedBy
    return nil
}
```

#### 2c. Repository Interface

**File**: `internal/domain/material/repository.go`

```go
package material

import "context"

// Repository adalah interface untuk persistence Material
// Implementasi ada di infrastructure layer
type Repository interface {
    Create(ctx context.Context, m *Material) error
    GetByCode(ctx context.Context, code MaterialCode) (*Material, error)
    List(ctx context.Context, filter ListFilter) ([]*Material, int32, error)
    Update(ctx context.Context, m *Material) error
    Delete(ctx context.Context, code MaterialCode) error
    ExistsByCode(ctx context.Context, code MaterialCode) (bool, error)
}

// ListFilter untuk filtering dan pagination
type ListFilter struct {
    Category *string
    IsActive *bool
    Page     int
    PageSize int
}
```

#### 2d. Domain Errors

**File**: `internal/domain/material/errors.go`

```go
package material

import "errors"

var (
    ErrNotFound            = errors.New("material not found")
    ErrAlreadyExists       = errors.New("material already exists")
    ErrInvalidMaterialCode = errors.New("invalid material code format")
    ErrInvalidCategory     = errors.New("invalid material category")
    ErrEmptyName           = errors.New("material name cannot be empty")
    ErrEmptyUOMCode        = errors.New("UOM code cannot be empty")
    ErrNegativeUnitCost    = errors.New("unit cost cannot be negative")
)
```

---

### Step 3: Infrastructure Layer (Repository Implementation)

**File**: `internal/infrastructure/postgres/material_repository.go`

```go
package postgres

import (
    "context"
    "database/sql"
    "errors"

    "github.com/homindolenern/goapps-costing-v1/internal/domain/material"
)

type MaterialRepository struct {
    db *sql.DB
}

func NewMaterialRepository(db *sql.DB) *MaterialRepository {
    return &MaterialRepository{db: db}
}

func (r *MaterialRepository) Create(ctx context.Context, m *material.Material) error {
    query := `
        INSERT INTO materials (
            material_code, material_name, category, 
            uom_code, unit_cost, is_active,
            created_at, created_by
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
    
    _, err := r.db.ExecContext(ctx, query,
        m.Code().String(),
        m.Name(),
        m.Category().String(),
        m.UOMCode(),
        m.UnitCost(),
        m.IsActive(),
        m.CreatedAt(),
        m.CreatedBy(),
    )
    
    if err != nil {
        if isUniqueViolation(err) {
            return material.ErrAlreadyExists
        }
        return err
    }
    
    return nil
}

func (r *MaterialRepository) GetByCode(ctx context.Context, code material.MaterialCode) (*material.Material, error) {
    query := `
        SELECT material_code, material_name, category,
               uom_code, unit_cost, is_active,
               created_at, created_by, updated_at, updated_by
        FROM materials 
        WHERE material_code = $1
    `
    
    row := r.db.QueryRowContext(ctx, query, code.String())
    return r.scanMaterial(row)
}

// ... other methods (List, Update, Delete, ExistsByCode)
```

**Migration**:

**File**: `migrations/postgres/000003_create_materials_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS materials (
    material_code VARCHAR(20) PRIMARY KEY,
    material_name VARCHAR(255) NOT NULL,
    category VARCHAR(50) NOT NULL,
    uom_code VARCHAR(10) NOT NULL REFERENCES uoms(uom_code),
    unit_cost DECIMAL(15,4) NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMPTZ,
    updated_by VARCHAR(100)
);

CREATE INDEX idx_materials_category ON materials(category);
CREATE INDEX idx_materials_uom_code ON materials(uom_code);
```

```bash
# Run migration
migrate -path migrations/postgres -database "$DATABASE_URL" up
```

---

### Step 4: Application Layer (Use Case Handlers)

Application layer mengimplementasikan use cases menggunakan Command/Query pattern.

**File**: `internal/application/material/create_handler.go`

```go
package material

import (
    "context"

    "github.com/homindolenern/goapps-costing-v1/internal/domain/material"
)

// CreateCommand adalah command untuk membuat material
type CreateCommand struct {
    MaterialCode string
    MaterialName string
    Category     string
    UOMCode      string
    UnitCost     float64
    CreatedBy    string
}

// CreateHandler menangani pembuatan material baru
type CreateHandler struct {
    repo material.Repository
}

func NewCreateHandler(repo material.Repository) *CreateHandler {
    return &CreateHandler{repo: repo}
}

func (h *CreateHandler) Handle(ctx context.Context, cmd CreateCommand) (*material.Material, error) {
    // 1. Validasi dan buat value objects
    code, err := material.NewMaterialCode(cmd.MaterialCode)
    if err != nil {
        return nil, err
    }

    category, err := material.NewCategory(cmd.Category)
    if err != nil {
        return nil, err
    }

    // 2. Cek apakah sudah exists
    exists, err := h.repo.ExistsByCode(ctx, code)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, material.ErrAlreadyExists
    }

    // 3. Buat entity
    entity, err := material.NewMaterial(
        code,
        cmd.MaterialName,
        category,
        cmd.UOMCode,
        cmd.UnitCost,
        cmd.CreatedBy,
    )
    if err != nil {
        return nil, err
    }

    // 4. Simpan ke repository
    if err := h.repo.Create(ctx, entity); err != nil {
        return nil, err
    }

    return entity, nil
}
```

**Buat handler lainnya**:
- `get_handler.go` - GetQuery + GetHandler
- `list_handler.go` - ListQuery + ListHandler + ListResult
- `update_handler.go` - UpdateCommand + UpdateHandler
- `delete_handler.go` - DeleteCommand + DeleteHandler

---

### Step 5: Delivery Layer (gRPC Handler)

**File**: `internal/delivery/grpc/material_handler.go`

```go
package grpc

import (
    "context"
    "errors"

    pb "github.com/homindolenern/goapps-costing-v1/gen/go/costing/v1"
    appmat "github.com/homindolenern/goapps-costing-v1/internal/application/material"
    "github.com/homindolenern/goapps-costing-v1/internal/domain/material"
)

type MaterialHandler struct {
    pb.UnimplementedMaterialServiceServer
    createHandler *appmat.CreateHandler
    updateHandler *appmat.UpdateHandler
    deleteHandler *appmat.DeleteHandler
    getHandler    *appmat.GetHandler
    listHandler   *appmat.ListHandler
    validator     *ValidationHelper
}

func NewMaterialHandler(
    createHandler *appmat.CreateHandler,
    updateHandler *appmat.UpdateHandler,
    deleteHandler *appmat.DeleteHandler,
    getHandler *appmat.GetHandler,
    listHandler *appmat.ListHandler,
    validator *ValidationHelper,
) *MaterialHandler {
    return &MaterialHandler{
        createHandler: createHandler,
        updateHandler: updateHandler,
        deleteHandler: deleteHandler,
        getHandler:    getHandler,
        listHandler:   listHandler,
        validator:     validator,
    }
}

func (h *MaterialHandler) CreateMaterial(
    ctx context.Context, 
    req *pb.CreateMaterialRequest,
) (*pb.CreateMaterialResponse, error) {
    // 1. Validate request (proto validation)
    if validationResp := h.validator.Validate(ctx, req); validationResp != nil {
        return &pb.CreateMaterialResponse{Base: validationResp}, nil
    }

    // 2. Build command
    cmd := appmat.CreateCommand{
        MaterialCode: req.MaterialCode,
        MaterialName: req.MaterialName,
        Category:     pbCategoryToString(req.Category),
        UOMCode:      req.UomCode,
        UnitCost:     req.UnitCost,
        CreatedBy:    "system", // TODO: dari context/auth
    }

    // 3. Execute handler
    entity, err := h.createHandler.Handle(ctx, cmd)
    if err != nil {
        return &pb.CreateMaterialResponse{
            Base: matErrorToBaseResponse(err),
        }, nil
    }

    // 4. Return success response
    return &pb.CreateMaterialResponse{
        Base: matSuccessResponse("Material created successfully"),
        Data: matEntityToProto(entity),
    }, nil
}

// Helper functions
func matSuccessResponse(message string) *pb.BaseResponse {
    return &pb.BaseResponse{
        StatusCode: "200",
        IsSuccess:  true,
        Message:    message,
    }
}

func matErrorToBaseResponse(err error) *pb.BaseResponse {
    statusCode := "500"
    message := "Internal server error"

    switch {
    case errors.Is(err, material.ErrNotFound):
        statusCode = "404"
        message = err.Error()
    case errors.Is(err, material.ErrAlreadyExists):
        statusCode = "409"
        message = err.Error()
    case errors.Is(err, material.ErrInvalidMaterialCode),
         errors.Is(err, material.ErrInvalidCategory),
         errors.Is(err, material.ErrEmptyName):
        statusCode = "400"
        message = err.Error()
    }

    return &pb.BaseResponse{
        StatusCode: statusCode,
        IsSuccess:  false,
        Message:    message,
    }
}
```

---

### Step 6: Wire Everything in main.go

**File**: `cmd/master-service/main.go`

```go
// Di function run()

// 1. Initialize repository
materialRepo := postgres.NewMaterialRepository(db)

// 2. Initialize application handlers
matCreateHandler := appmaterial.NewCreateHandler(materialRepo)
matUpdateHandler := appmaterial.NewUpdateHandler(materialRepo)
matDeleteHandler := appmaterial.NewDeleteHandler(materialRepo)
matGetHandler := appmaterial.NewGetHandler(materialRepo)
matListHandler := appmaterial.NewListHandler(materialRepo)

// 3. Initialize gRPC handler
materialHandler := grpcdelivery.NewMaterialHandler(
    matCreateHandler,
    matUpdateHandler,
    matDeleteHandler,
    matGetHandler,
    matListHandler,
    validationHelper, // shared validator
)

// Di function runGRPCServer()

// 4. Register service
pb.RegisterMaterialServiceServer(grpcServer, materialHandler)
```

---

### Step 7: Testing

#### Integration Test

**File**: `tests/integration/material_test.go`

```go
package integration

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/homindolenern/goapps-costing-v1/internal/domain/material"
)

func TestMaterialDomain_CreateValid(t *testing.T) {
    code, err := material.NewMaterialCode("MAT001")
    require.NoError(t, err)

    cat, err := material.NewCategory("RAW")
    require.NoError(t, err)

    entity, err := material.NewMaterial(
        code, "Raw Material 1", cat, "KG", 100.50, "test",
    )
    
    require.NoError(t, err)
    assert.Equal(t, "MAT001", entity.Code().String())
    assert.Equal(t, "Raw Material 1", entity.Name())
    assert.True(t, entity.IsActive())
}

func TestMaterialDomain_InvalidCode(t *testing.T) {
    _, err := material.NewMaterialCode("invalid")
    assert.ErrorIs(t, err, material.ErrInvalidMaterialCode)
}
```

```bash
# Run tests
go test -v ./tests/integration/...
go test -v ./internal/domain/material/...
```

#### E2E Test with grpcurl

```bash
# Create material
grpcurl -plaintext -d '{
  "material_code": "MAT001",
  "material_name": "Raw Material 1",
  "category": "MATERIAL_CATEGORY_RAW",
  "uom_code": "KG",
  "unit_cost": 100.50
}' localhost:9090 costing.v1.MaterialService/CreateMaterial

# Get material
grpcurl -plaintext -d '{"material_code": "MAT001"}' localhost:9090 costing.v1.MaterialService/GetMaterial

# List materials
grpcurl -plaintext -d '{"page": 1, "page_size": 10}' localhost:9090 costing.v1.MaterialService/ListMaterials
```

---

## 6. Testing Guide

### Test Structure
```
tests/
├── integration/        # Integration tests (database, etc)
│   ├── uom_test.go
│   ├── parameter_test.go
│   └── material_test.go
└── e2e/                # End-to-end tests
    └── api_test.go
```

### Run Tests
```bash
# All tests
go test -v ./...

# Specific package
go test -v ./internal/domain/material/...

# With coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Integration tests only
go test -v ./tests/integration/...
```

---

## 7. Code Standards

### Naming Conventions
| Type | Convention | Example |
|------|-----------|---------|
| Package | lowercase, single word | `material`, `uom` |
| Interface | `-er` suffix when possible | `Reader`, `Repository` |
| Struct | PascalCase | `MaterialHandler` |
| Function | PascalCase (exported), camelCase (private) | `NewMaterial`, `parseCode` |
| Constant | ALL_CAPS for public | `ErrNotFound` |
| Variable | camelCase | `materialCode` |

### Error Handling
```go
// ✅ DO: Return errors, let caller handle
func GetMaterial(code string) (*Material, error) {
    if code == "" {
        return nil, ErrEmptyCode
    }
    // ...
}

// ❌ DON'T: Panic in library code
func GetMaterial(code string) *Material {
    if code == "" {
        panic("empty code") // Bad!
    }
}
```

### Response Format
Semua response menggunakan `BaseResponse`:
```json
{
  "base": {
    "statusCode": "200",
    "isSuccess": true,
    "message": "Success",
    "validationErrors": []
  },
  "data": { ... }
}
```

### Lint Error Prevention

Project ini menggunakan **golangci-lint** dengan konfigurasi ketat. Berikut panduan untuk menghindari lint error yang umum terjadi:

#### Pre-Commit Checklist
```bash
# WAJIB jalankan sebelum commit:
go build ./...              # Pastikan build pass
golangci-lint run           # Pastikan 0 errors
go test -v ./...            # Pastikan tests pass
```

#### Install golangci-lint
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
```

#### Common Lint Errors & Solutions

##### 1. godot: Comment should end in a period
❌ **Error**:
```go
// CreateHandler handles the creation of UOM
type CreateHandler struct {}
```

✅ **Solution**: Semua exported comments HARUS diakhiri dengan titik.
```go
// CreateHandler handles the creation of UOM.
type CreateHandler struct {}
```

> **TIP**: Untuk multi-line comment, titik ada di akhir paragraf, bukan di setiap baris.

##### 2. revive: Type stuttering
❌ **Error**:
```go
package uom

type UOMCode struct { value string }  // BAD: "uom.UOMCode" stutters
```

✅ **Solution**: Gunakan nama yang singkat karena package name sudah menjelaskan context.
```go
package uom

type Code struct { value string }  // GOOD: "uom.Code"
```

##### 3. errorlint: Type assertion on error
❌ **Error**:
```go
if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
    return nil, err
}
```

✅ **Solution**: Gunakan `errors.As` untuk handle wrapped errors.
```go
var configFileNotFoundError viper.ConfigFileNotFoundError
if !errors.As(err, &configFileNotFoundError) {
    return nil, err
}
```

##### 4. exhaustive: Missing switch cases
❌ **Error**:
```go
switch cat {
case pb.UOMCategory_UOM_CATEGORY_WEIGHT:
    return "WEIGHT"
default:
    return ""
}  // Missing UNSPECIFIED case!
```

✅ **Solution**: Handle semua enum values termasuk UNSPECIFIED.
```go
switch cat {
case pb.UOMCategory_UOM_CATEGORY_WEIGHT:
    return "WEIGHT"
case pb.UOMCategory_UOM_CATEGORY_UNSPECIFIED:
    return ""
}
return ""  // fallback
```

##### 5. copylocks: Copying mutex value
❌ **Error**:
```go
response := map[string]interface{}{
    "base": pb.BaseResponse{...},  // BAD: copies protobuf struct with mutex
}
```

✅ **Solution**: Gunakan pointer untuk protobuf structs.
```go
response := map[string]interface{}{
    "base": &pb.BaseResponse{...},  // GOOD: pointer reference
}
```

##### 6. predeclared: Shadowing built-in identifiers
❌ **Error**:
```go
min := 100.0  // BAD: shadows built-in min()
max := 200.0  // BAD: shadows built-in max()
```

✅ **Solution**: Gunakan nama yang lebih spesifik.
```go
minVal := 100.0   // GOOD
maxVal := 200.0   // GOOD
minFloat := 100.0 // GOOD
```

##### 7. goimports: File not properly formatted
❌ **Error**: Import groups tidak benar atau tidak sorted.

✅ **Solution**: Jalankan goimports sebelum commit.
```bash
goimports -w .
# atau untuk file spesifik
goimports -w internal/domain/uom/entity.go
```

##### 8. gocritic: ifElseChain should be switch
❌ **Error**:
```go
if path == "/swagger/" {
    // ...
} else if path == "/swagger/api.json" {
    // ...
} else {
    // ...
}
```

✅ **Solution**: Gunakan switch statement.
```go
switch path {
case "/swagger/":
    // ...
case "/swagger/api.json":
    // ...
default:
    // ...
}
```

##### 9. errcheck: Unchecked error return
❌ **Error**:
```go
w.Write([]byte(html))  // BAD: ignoring error
json.NewEncoder(w).Encode(data)  // BAD: ignoring error
```

✅ **Solution**: Handle error atau explicitly ignore dengan `_ =`.
```go
_, _ = w.Write([]byte(html))  // Explicitly ignored (acceptable for http.ResponseWriter)
_ = json.NewEncoder(w).Encode(data)
```

> **NOTE**: Untuk http.ResponseWriter.Write, error biasanya tidak actionable, jadi explicitly ignore acceptable.

#### Automatic Fixes
```bash
# Fix import formatting
goimports -w .

# Fix general formatting
gofmt -w .

# Fix all godot issues (add periods to comments)
find . -name "*.go" -type f ! -path "./gen/*" ! -path "./vendor/*" \
  -exec sed -i 's/^\/\/ \([A-Z][a-zA-Z0-9_]*\) \(.*[a-zA-Z0-9)]\)$/\/\/ \1 \2./g' {} \;
```

#### IDE Setup (Recommended)
Configure IDE untuk auto-run golangci-lint on save:

**GoLand/IntelliJ**:
1. Settings → Tools → File Watchers
2. Add golangci-lint watcher

**VSCode**:
```json
// .vscode/settings.json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "editor.formatOnSave": true,
  "go.formatTool": "goimports"
}
```

---

## 8. Git Workflow

### Branch Strategy
```
main           ← Production-ready code
  └── develop  ← Development integration
       ├── feature/add-material-service
       ├── fix/validation-error-format
       └── refactor/cleanup-handlers
```

### Commit Format
```
type: description

Types:
- feat:     New feature
- fix:      Bug fix
- refactor: Code refactoring
- docs:     Documentation
- test:     Adding tests
- chore:    Build/tooling changes
```

### Pull Request Checklist
- [ ] Tests pass (`go test ./...`)
- [ ] Lint pass (`golangci-lint run`)
- [ ] Proto generated (`buf generate`)
- [ ] Migration added if needed
- [ ] Documentation updated

---

## 9. CI/CD & Deployment

### GitHub Actions Pipeline
```
Push → Lint → Test → Build → Docker → Deploy (Staging) → Deploy (Production)
```

### Local Docker Build
```bash
# Build image
docker build -t goapps-costing:latest .

# Run container
docker run -p 9090:9090 -p 8080:8080 \
  -e DATABASE_HOST=host.docker.internal \
  goapps-costing:latest
```

### Kubernetes Deploy
```bash
# Staging
kubectl apply -f deployments/k8s/staging/

# Production
kubectl apply -f deployments/k8s/production/
```

---

## 10. Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| `buf: command not found` | Install buf: `go install github.com/bufbuild/buf/cmd/buf@latest` |
| Proto compilation error | Run `buf lint` to check proto syntax |
| Database connection refused | Check docker compose: `docker compose ps` |
| Migration failed | Check DATABASE_URL and run `migrate version` |
| Validation not working | Ensure `buf.validate` imported in proto |

### Useful Commands
```bash
# Check proto syntax
buf lint

# Format proto files
buf format -w

# Check gRPC reflection
grpcurl -plaintext localhost:9090 list

# View migration status
migrate -path migrations/postgres -database "$DATABASE_URL" version

# Reset database
migrate -path migrations/postgres -database "$DATABASE_URL" drop
migrate -path migrations/postgres -database "$DATABASE_URL" up
```

---

## Summary Checklist

Ketika menambah service baru, ikuti checklist ini:

- [ ] **Proto**: Buat file `.proto` dengan messages dan service
- [ ] **Generate**: Jalankan `buf generate`
- [ ] **Domain**: Buat value objects, entity, repository interface, errors
- [ ] **Infrastructure**: Implementasi repository + migration
- [ ] **Application**: Buat handlers (Create, Get, List, Update, Delete)
- [ ] **Delivery**: Buat gRPC handler dengan validation
- [ ] **Wire**: Register di `main.go`
- [ ] **Test**: Tulis integration tests
- [ ] **Docs**: Update Swagger (auto dari proto)
- [ ] **Commit**: Push dengan format commit yang benar
