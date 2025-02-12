package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Todo struct {
	ID        int    `json:"id"`
	Completed bool   `json:"completed"`
	Body      string `json:"body"`
}

func main() {
	// Getting data from env

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading env file", err)
	}

	// todos := &[]Todo{} donot take pointer on global variables
	todos := []Todo{}

	server := http.NewServeMux()

	server.HandleFunc("GET /v1/api/todos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(&todos)
	})

	server.HandleFunc("POST /v1/api/todos", func(w http.ResponseWriter, r *http.Request) {
		todo := &Todo{}

		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Access-Control-Allow-Methods", "*")
		w.Header().Add("Access-Control-Allow-Origins", "*")
		err := json.NewDecoder(r.Body).Decode(todo)

		fmt.Println(err)
		if err == io.EOF {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode("Data to upload is required")
			return
		}

		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode("Facing conversion error")
			return
		}

		todo.ID = len(todos) + 1
		todos = append(todos, *todo)
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(*todo)

	})

	server.HandleFunc("PATCH /v1/api/todos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Access-Control-Allow-Methods", "*")
		id := r.URL.Query().Get("id")
		for i, todo := range todos {
			if fmt.Sprint(todo.ID) == id {
				todos[i].Completed = !todos[i].Completed
				w.WriteHeader(200)
				json.NewEncoder(w).Encode(&todos[i])
				return
			}
		}

		w.WriteHeader(400)
		json.NewEncoder(w).Encode("updation for the requested todo failed")
	})

	server.HandleFunc("DELETE /v1/api/todos", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		for i, todo := range todos {
			if fmt.Sprint(todo.ID) == id {
				todos = append(todos[:i], todos[i+1:]...)
				w.WriteHeader(204)
				json.NewEncoder(w).Encode("Deleted the requested record")
				return
			}
		}

		w.WriteHeader(404)
		json.NewEncoder(w).Encode("Record not found")
	})
	slog.Info("message", "server listening at", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), server))
}
