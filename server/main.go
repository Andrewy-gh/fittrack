package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Andrewy-gh/fittrack/server/internal/api"
	"github.com/Andrewy-gh/fittrack/server/internal/db"
	"github.com/jackc/pgx/v5"
)

func main() {
	dbURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer conn.Close(context.Background())

	// Create queries instance
	queries := db.New(conn)

	// Create API handler
	apiHandler := api.NewHandler(queries)

	// Setup routes
	router := http.NewServeMux()
	apiHandler.RegisterRoutes(router)

	// Serve static files
	fileServer := http.FileServer(http.Dir("./dist"))
	router.Handle("/", fileServer)

	log.Println("Starting server on port 8080...")
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
