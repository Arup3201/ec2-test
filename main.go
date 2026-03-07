package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func checkHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"message": "healthy",
	})
}

type Todo struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
}

func getTodos(w http.ResponseWriter, r *http.Request) {

	var todos []Todo
	result := db.Find(&todos)
	if result.Error != nil {
		if errors.Is(result.Error, sql.ErrNoRows) {
			json.NewEncoder(w).Encode(map[string]any{
				"todos": []Todo{},
			})
			return
		}

		http.Error(w, "Failed to fetch all todos from database", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"todos": todos,
	})
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	type createTodoRequest struct {
		Title       string `json:"title"`
		Description string `json:"description,omitempty"`
	}

	var payload createTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	todo := Todo{
		Title:       payload.Title,
		Description: payload.Description,
	}
	err := gorm.G[Todo](db).Create(context.Background(), &todo)
	if err != nil {
		http.Error(w, "Failed to create todo in database", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Todo created",
	})
}

func main() {
	var err error

	DB_HOST := os.Getenv("DB_HOST")
	DB_USER := os.Getenv("DB_USER")
	DB_PORT := os.Getenv("DB_PORT")
	DB_PASS := os.Getenv("DB_PASS")
	DB_NAME := os.Getenv("DB_NAME")

	if DB_HOST == "" || DB_USER == "" || DB_PORT == "" || DB_PASS == "" || DB_NAME == "" {
		log.Fatalf("[ERROR] missing database connection values\n")
	}

	dsn := fmt.Sprintf("host=%s user=%s port=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_USER, DB_PORT, DB_PASS, DB_NAME)

	db, err = gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{})
	if err != nil {
		log.Fatalf("[ERROR] gorm open postgres database: %s\n", err)
	}

	err = db.AutoMigrate(&Todo{})
	if err != nil {
		log.Fatalf("[ERROR] database todo table migrate: %s\n", err)
	}

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}

	server := &http.Server{
		Addr:    "0.0.0.0:" + PORT,
		Handler: nil,
	}
	http.HandleFunc("GET /health", checkHealth)
	http.HandleFunc("GET /", getTodos)
	http.HandleFunc("POST /", createTodo)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("[ERROR] server listen and serve: %s\n", err)
	}
}
