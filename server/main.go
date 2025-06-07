package main

import (
	"encoding/json"
	"io"
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

	router.HandleFunc("POST /api/workouts", func(w http.ResponseWriter, r *http.Request) {
		// Read the raw JSON body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Parse the JSON body
		// var workout map[string]interface{}
		// err = json.Unmarshal(body, &workout)
		// if err != nil {
		// 	http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		// 	return
		// }

		log.Printf("Received work JSON: %s\n", string(body))

		// Optionally, respond to the client
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"received"}`))
	})

	fileServer := http.FileServer(http.Dir("./dist"))
	router.Handle("/", fileServer)

	log.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
