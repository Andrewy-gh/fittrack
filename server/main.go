package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("GET /api/hello", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]string{
			"message": "Hello World",
		}

		// Set the Content-Type header to application/json
		w.Header().Set("Content-Type", "application/json")

		// Encode the data to JSON and write it to the response
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		}
	})

	fileServer := http.FileServer(http.Dir("./dist"))
	router.Handle("/", fileServer)

	log.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
