package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: nil,
	}
	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"message": "healthy",
		})
	})

	err := server.ListenAndServe()
	if err != nil {
		log.Printf("[ERROR] server listen and serve failed with error: %s\n", err)
	}
}
