# API Documentation & Type Generation Decision

**Date:** 2025-08-10  
**Status:** Approved  
**Context:** Implementing PUT /api/workouts/:id and establishing type safety between Go backend and TypeScript frontend

## Problem Statement

We need to:
1. Document the API contract for PUT /api/workouts/:id endpoint
2. Establish automatic type generation to prevent drift between backend and frontend
3. Ensure both FE & BE teams are aligned on API contracts

## Current State Analysis

**Go Backend:**
- ✅ Structured handlers with validation using `validator/v10`
- ✅ JSON struct tags already in place
- ✅ sqlc-generated database models
- ✅ Clean separation of concerns (handler/service/repository)

**TypeScript Frontend:**
- ✅ TanStack Router + Query setup
- ✅ Vite build system
- ❌ Manual type definitions with drift issues
- ❌ `WorkoutFormValues` vs `CreateWorkoutRequest` misalignment

## Decision: OpenAPI + Swaggo + oapi-codegen

We chose **Option 1** from the original analysis with enhanced tooling:

### Selected Tools
- **Backend:** `swaggo/swag` for OpenAPI spec generation from Go annotations
- **Frontend:** `oapi-codegen` for TypeScript client + types generation
- **Standard:** OpenAPI 3.0 specification

### Why This Approach

**Technical Fit:**
- Leverages existing Go struct tags and validation rules
- Generates both types AND client methods
- Better Go → TypeScript mapping than alternatives
- Industry standard (OpenAPI) for long-term maintainability

**Rejected Alternatives:**
- ❌ **Option 2 (JSON inference):** Too brittle, requires live API calls
- ❌ **Option 3 (go2ts direct):** Less accurate type mapping, no client generation  
- ❌ **Option 4 (Runtime validation):** Unnecessary overhead for our use case

## Implementation Plan

### Phase 1: Setup Swaggo in Go Backend
```bash
# Install tools
go install github.com/swaggo/swag/cmd/swag@latest
go get github.com/swaggo/http-swagger@latest

# Add swagger annotations to handlers
# Generate spec
swag init -g cmd/api/main.go -o ./docs
```

### Phase 2: Setup TypeScript Generation
```bash
# Install frontend tools
npm install -D @apidevtools/swagger-cli oapi-codegen

# Generate types (to be automated in CI)
npx swagger-cli bundle ../server/docs/swagger.json --outfile api-spec.json
npx oapi-codegen -package=api -generate=types,client api-spec.json > src/lib/generated-api.ts
```

### Phase 3: Integration & Automation
- Add type generation to build scripts
- Update CI/CD to regenerate types on Go changes
- Migrate existing manual types to generated ones

## PUT /api/workouts/:id Specification

**Request:**
- Method: `PUT`
- Path: `/api/workouts/{id}`
- Body: `WorkoutFormValues` (matches existing client structure)
- Headers: `x-stack-access-token`

**Responses:**
- `200`: Updated workout with `workout_id`
- `204`: No content (alternative success)
- `400`: Validation errors
- `403`: RLS/ownership violations  
- `404`: Workout not found
- `500`: Internal server error

## Success Criteria

- [ ] Zero type drift between Go backend and TypeScript frontend
- [ ] Automatic regeneration of types on backend changes
- [ ] Full type safety for API calls
- [ ] Documentation serves both internal teams and external consumers
- [ ] PUT endpoint properly documented and implemented

## Next Steps

1. Implement swagger annotations on existing endpoints
2. Create PUT /api/workouts/:id endpoint with full documentation
3. Generate initial TypeScript types
4. Set up automated regeneration pipeline
5. Migrate frontend to use generated types

## Future Considerations

- Consider adding request/response validation middleware using generated schemas
- Evaluate adding API versioning strategy
- Monitor for performance impact of swagger middleware in production

---

**Decision Maker:** Development Team  
**Implementation Owner:** Backend Team (swagger) + Frontend Team (integration)  
**Review Date:** 2025-09-10 (1 month post-implementation)
