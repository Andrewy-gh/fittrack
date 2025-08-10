# Tutorial: How We Set Up Swagger

## Step 1: Install Dependencies
```bash
go get -u github.com/swaggo/swag/cmd/swag
go get -u github.com/swaggo/http-swagger
go get -u github.com/swaggo/files
```

## Step 2: Add API Metadata to main.go
```go
// @title FitTrack API
// @version 1.0
// @description A fitness tracking application API
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name x-stack-access-token
```
## Step 3: Create Response Types
```go
// internal/response/types.go
type ErrorResponse struct {
    Error   string `json:"error" example:"Invalid request"`
    Message string `json:"message" example:"The request body is malformed"`
}

type SuccessResponse struct {
    Message string `json:"message" example:"Operation completed successfully"`
}
```

## Step 4: Document Endpoints
```go
// CreateWorkout godoc
// @Summary Create a new workout
// @Description Create a new workout with exercises and sets
// @Tags workouts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body workout.CreateWorkoutRequest true "Workout data"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Router /workouts [post]
func (h *WorkoutHandler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
    // handler implementation
}
```
## Step 5: Add Swagger Route
```go
// cmd/api/routes.go
import httpSwagger "github.com/swaggo/http-swagger"

func (api *api) routes(...) *http.ServeMux {
    mux := http.NewServeMux()
    // ... other routes
    mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
    return mux
}
```

## Step 6: Generate Documentation
```bash
swag init -g cmd/api/main.go -o docs
```

## Step 7: Generate Documentation
```bash
swag init -g cmd/api/main.go -o docs
```

## Step 8: Access Documentation
•  Start your server: `go run cmd/api/main.go`
•  Visit: `http://localhost:8080/swagger/`

Key Tips:
1. **Always regenerate docs** after changing annotations: `swag init -g cmd/api/main.go -o docs`
2. **Use consistent response types** across your API
3. **Document all parameters, responses, and error codes**
4. **Keep swagger types separate** from internal models to avoid circular imports
5. **Test your documented endpoints** using the Swagger UI

The setup provides a fully interactive API documentation that developers can use to understand and test your endpoints!