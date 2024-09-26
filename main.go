package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func main() {
	db, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	DB = db
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS "Todo" (
			id SERIAL PRIMARY KEY,
			todo TEXT,
			done BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP);
		`)
	if err != nil {
		slog.Info(err.Error())
		panic(err)
	}
	slog.Info("Table created")

	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./frontend/assets/"))))

	mux.HandleFunc("GET /{$}", Home)

	mux.HandleFunc("POST /createTodo/{$}", TodoPost)
	mux.HandleFunc("PUT /updateTodo/{todoId}/{$}", TodoUpdate)
	mux.HandleFunc("DELETE /deleteTodo/{todoId}/{$}", TodoDelete)

	if err := http.ListenAndServe(":8000", middleware(mux)); err != nil {
		log.Fatal(err)
	}
}

func middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
			w.Header().Set("Content-Encoding", "br")
			br := brotli.NewWriter(w)
			defer br.Close()
			w = &responseWriter{
				ResponseWriter: w,
				Writer:         br,
			}
		}
		fmt.Println(r.Method, r.URL.Path, time.Since(start))
		h.ServeHTTP(w, r)
	})
}

type Todo struct {
	ID        int
	Todo      string
	Done      bool
	CreatedAt time.Time
}

func Home(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query(context.Background(), `SELECT * FROM "Todo" ORDER BY id ASC`)
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
			http.Error(w, "Error getting todos", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error getting todos", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("./frontend/index.html")
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error getting todos", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, todos)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error getting todos", http.StatusInternalServerError)
		return
	}
}

func TodoPost(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	todo.Todo = r.FormValue("todo")

	err := DB.QueryRow(context.Background(), `INSERT INTO "Todo" (todo) VALUES ($1) RETURNING id`, todo.Todo).Scan(&todo.ID)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error creating todo", http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.ParseFiles("./frontend/index.html"))
	tmpl.ExecuteTemplate(w, "todo", Todo{ID: todo.ID, Todo: todo.Todo, Done: todo.Done})
}

func TodoUpdate(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	id, _ := strconv.Atoi(r.PathValue("todoId"))
	todo.Todo = r.PostFormValue("todo")

	done := r.PostFormValue("done")
	if done == "on" {
		done = "true"
	} else {
		done = "false"
	}
	todo.Done, _ = strconv.ParseBool(done)

	_, err := DB.Exec(context.Background(), `UPDATE "Todo" SET todo = $1, done = $2 WHERE id = $3`, todo.Todo, todo.Done, id)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error updating todo", http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.ParseFiles("./frontend/index.html"))
	tmpl.ExecuteTemplate(w, "todo", Todo{ID: id, Todo: todo.Todo, Done: todo.Done})
}

func TodoDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("todoId")

	_, err := DB.Exec(context.Background(), `DELETE FROM "Todo" WHERE id = $1`, id)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error deleting todo", http.StatusInternalServerError)
		return
	}
}

type responseWriter struct {
	http.ResponseWriter
	Writer *brotli.Writer
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	return rw.Writer.Write(data)
}

func (rw *responseWriter) Flush() error {
	return rw.Writer.Flush()
}
