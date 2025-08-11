# Manual Swagger Types Tutorial

## Overview

This guide explains how to manually define Swagger types for API documentation when automatic generation from database models isn't suitable. This approach is commonly needed when using database-specific types that don't translate well to API schemas.

## When to Use Manual Swagger Types

### Common Use Cases

1. **Database-Specific Types**: When using libraries like `pgx` or `sqlc` that generate database-specific types
2. **Complex Type Mappings**: When your database schema differs significantly from your API response structure
3. **Validation Requirements**: When you need specific validation rules that aren't present in your database models
4. **API Versioning**: When you want to maintain stable API contracts independent of database changes
5. **Custom Field Transformations**: When you need to transform field names, types, or structure for the API

### Our Specific Case: pgx/sqlc Types

In our project, we use `sqlc` to generate database models from PostgreSQL schemas. These generated types:
- Use PostgreSQL-specific types (e.g., `pgtype.Text`, `pgtype.Int4`)
- Include internal database fields not meant for API exposure
- Don't include validation tags needed for Swagger documentation
- May have different field names than desired for the API

## Creating Manual Swagger Types

### 1. File Organization

Create dedicated files for Swagger types in each module:

```
internal/
├── exercise/
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── swagger_types.go  ← Manual Swagger types
└── workout/
    ├── handler.go
    ├── service.go
    ├── repository.go
    └── swagger_types.go   ← Manual Swagger types
```

### 2. Basic Type Definition

```go
package exercise

import "time"

// ExerciseResponse represents an exercise response for swagger documentation
type ExerciseResponse struct {
    ID        int32     `json:"id" validate:"required" example:"1"`
    Name      string    `json:"name" validate:"required" example:"Bench Press"`
    CreatedAt time.Time `json:"created_at" validate:"required" example:"2023-01-01T15:04:05Z"`
    UpdatedAt time.Time `json:"updated_at" validate:"required" example:"2023-01-01T15:04:05Z"`
    UserID    string    `json:"user_id" validate:"required" example:"user-123"`
}
```

### 3. Required Struct Tags

#### JSON Tags
```go
`json:"field_name"` // Controls field name in JSON output
`json:"field_name,omitempty"` // Omits field if empty/nil
```

#### Validation Tags
```go
`validate:"required"` // Marks field as required in OpenAPI spec
`validate:"min=1"` // Minimum length/value validation
`validate:"max=256"` // Maximum length/value validation
`validate:"gte=0"` // Greater than or equal to
```

#### Documentation Tags
```go
`example:"Sample Value"` // Provides example value in Swagger UI
```

### 4. Advanced Type Patterns

#### Optional Fields
```go
type WorkoutResponse struct {
    ID    int32   `json:"id" validate:"required" example:"1"`
    Notes *string `json:"notes,omitempty" example:"Great workout today"`
}
```

#### Nested Types
```go
type CreateWorkoutRequest struct {
    Date      string          `json:"date" validate:"required"`
    Exercises []ExerciseInput `json:"exercises" validate:"required,min=1"`
    Notes     string          `json:"notes,omitempty" validate:"max=256"`
}

type ExerciseInput struct {
    Name string     `json:"name" validate:"required,min=1,max=256"`
    Sets []SetInput `json:"sets" validate:"required,min=1"`
}

type SetInput struct {
    Weight  *int32 `json:"weight,omitempty" validate:"gte=0"`
    Reps    int32  `json:"reps" validate:"required,min=1"`
    SetType string `json:"setType" validate:"required,oneof=warmup working"`
}
```

#### Enums with Validation
```go
type SetInput struct {
    SetType string `json:"setType" validate:"required,oneof=warmup working"`
    //                                              ^^^^^
    //                                              Restricts to specific values
}
```

### 5. Handler Integration

Link your Swagger types to handler endpoints using Swagger annotations:

```go
// ListExercises godoc
// @Summary List exercises
// @Description Get all exercises for the authenticated user
// @Tags exercises
// @Accept json
// @Produce json
// @Security StackAuth
// @Success 200 {array} exercise.ExerciseResponse  ← References our manual type
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /exercises [get]
func (h *ExerciseHandler) ListExercises(w http.ResponseWriter, r *http.Request) {
    // Implementation...
}
```

### 6. Type Conversion

Since manual types differ from database models, create conversion functions:

```go
// Convert database model to API response type
func dbExerciseToResponse(dbExercise db.Exercise) ExerciseResponse {
    return ExerciseResponse{
        ID:        dbExercise.ID,
        Name:      dbExercise.Name,
        CreatedAt: dbExercise.CreatedAt.Time, // Convert pgtype.Timestamp to time.Time
        UpdatedAt: dbExercise.UpdatedAt.Time,
        UserID:    dbExercise.UserID,
    }
}

// Convert multiple database models
func dbExercisesToResponse(dbExercises []db.Exercise) []ExerciseResponse {
    responses := make([]ExerciseResponse, len(dbExercises))
    for i, dbExercise := range dbExercises {
        responses[i] = dbExerciseToResponse(dbExercise)
    }
    return responses
}
```

### 7. Service Layer Integration

Use conversion functions in your service layer:

```go
func (s *ExerciseService) ListExercises(ctx context.Context) ([]ExerciseResponse, error) {
    // Get data from repository (returns db types)
    dbExercises, err := s.repo.ListExercises(ctx)
    if err != nil {
        return nil, err
    }
    
    // Convert to API response types
    return dbExercisesToResponse(dbExercises), nil
}
```

## Generation Workflow

### 1. Update Swagger Documentation
After creating or modifying manual types, regenerate Swagger documentation:

```bash
make swagger  # or: swag init -g cmd/api/main.go -o docs
```

### 2. Frontend Type Generation
The updated `swagger.json` can then be used to generate frontend types:

```bash
# Example with swagger-typescript-api
npx swagger-typescript-api -p swagger.json -o ./src/lib/api --modular
```

## Best Practices

### 1. Consistent Naming
- Use consistent field naming conventions (snake_case, camelCase, etc.)
- Match your frontend expectations
- Be consistent with existing API patterns

### 2. Validation Rules
```go
// Good: Specific validation rules
`validate:"required,min=1,max=256"`

// Avoid: Generic or missing validation
`validate:"required"`
```

### 3. Documentation
```go
// Good: Descriptive examples and comments
type WorkoutResponse struct {
    // Unique identifier for the workout
    ID   int32     `json:"id" validate:"required" example:"1"`
    Date time.Time `json:"date" validate:"required" example:"2023-01-01T15:04:05Z"`
}
```

### 4. Type Safety
- Always create conversion functions between DB and API types
- Use pointer types (`*string`, `*int32`) for optional fields
- Validate converted data before sending responses

## Alternative Approaches

### When Manual Types Aren't Needed

1. **Simple CRUD APIs**: If using ORMs like GORM that generate clean types
2. **Direct JSON Marshaling**: When database types already marshal correctly
3. **GraphQL APIs**: Where schema-first approaches handle type generation

### Hybrid Approaches

1. **Embedded Types**: Embed database types and add Swagger-specific fields
2. **Interface-Based**: Use interfaces for common operations while maintaining type flexibility
3. **Code Generation**: Create custom tools to generate Swagger types from database schemas

## Troubleshooting

### Common Issues

1. **Missing Required Fields**: Ensure `validate:"required"` tags are present
2. **Incorrect JSON Names**: Check `json` tag formatting
3. **Type Conversion Errors**: Verify conversion functions handle all edge cases
4. **Swagger Generation Fails**: Ensure all referenced types are properly defined

### Validation

Test your types by:
1. Generating Swagger documentation (`make swagger`)
2. Checking the resulting `swagger.json` for correct schema
3. Testing API endpoints with various payloads
4. Verifying frontend type generation produces expected results

## Conclusion

Manual Swagger types provide precise control over your API documentation and client generation. While they require additional maintenance, they offer:

- **Type Safety**: Clear separation between database and API concerns
- **Flexibility**: Easy customization of field names, validation, and structure
- **Documentation**: Rich examples and descriptions for better developer experience
- **Evolution**: API contracts can evolve independently of database schema

This approach is particularly valuable in complex applications where the database schema doesn't directly map to desired API structure, which is common when using database-specific libraries or when maintaining backward compatibility.
