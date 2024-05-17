package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func main() {
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	defer DB.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS "Todo" (
			id SERIAL PRIMARY KEY,
			todo TEXT,
			done BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
		`)
	if err != nil {
		slog.Info(err.Error())
		panic(err)
	}
	slog.Info("Table created")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /getTodos/{$}", TodoGet)
	mux.HandleFunc("POST /createTodo/{$}", TodoPost)
	mux.HandleFunc("PUT /updateTodo/{todoId}/{$}", TodoUpdate)
	mux.HandleFunc("DELETE /deleteTodo/{todoId}/{$}", TodoDelete)

	if err := http.ListenAndServe(":8000", addCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func addCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})
}

type Todo struct {
	ID        int    `json:"id"`
	Todo      string `json:"todo"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"createdAt"`
}

func TodoGet(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query(`SELECT * FROM "Todo" ORDER BY id ASC`)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Todo, &todo.Done, &todo.CreatedAt); err != nil {
			slog.Error(err.Error())
			http.Error(w, "Error scanning todos", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating through todos", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(todos)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Unable to marshal JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonData)
}

func TodoPost(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	json.NewDecoder(r.Body).Decode(&todo)

	_, err := DB.Exec(`INSERT INTO "Todo" (todo, done) VALUES ($1, $2)`, todo.Todo, todo.Done)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error inserting todo", http.StatusInternalServerError)
		return
	}
}

func TodoUpdate(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	json.NewDecoder(r.Body).Decode(&todo)

	id := r.PathValue("todoId")

	_, err := DB.Exec(`UPDATE "Todo" SET todo = $1, done = $2 WHERE id = $3`, todo.Todo, todo.Done, id)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error updating todo", http.StatusInternalServerError)
		return
	}
}

func TodoDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("todoId")

	_, err := DB.Exec(`DELETE FROM "Todo" WHERE id = $1`, id)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error deleting todo", http.StatusInternalServerError)
		return
	}
}
