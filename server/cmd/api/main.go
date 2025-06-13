package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/jackc/pgx/v5"
)

type api struct {
	logger  *slog.Logger
	queries *db.Queries
	conn    *pgx.Conn
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// Initialize dependencies
	queries := db.New(conn)
	workoutService := workout.NewService(logger, queries)
	workoutHandler := workout.NewHandler(workoutService)
	exerciseService := exercise.NewService(logger, queries)
	exerciseHandler := exercise.NewHandler(exerciseService)

	api := &api{
		logger:  logger,
		queries: queries,
		conn:    conn,
	}

	logger.Info("starting server", "addr", ":8080")

	err = http.ListenAndServe(":8080", api.routes(workoutHandler, exerciseHandler))
	logger.Error(err.Error())
	os.Exit(1)
}
