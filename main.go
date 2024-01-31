package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB
var path string

func openDB() error {
	db, err := sql.Open("postgres", os.Getenv("PRISMA_DB"))
	if err != nil {
		return err
	}
	DB = db
	return nil
}

func closeDB() error {
	return DB.Close()
}

func main() {
	if os.Getenv("PROD") != "true" {
		fmt.Println("dev")
		path = "./src"
	} else {
		fmt.Println("prod")
		path = "./bin/dist"
	}

	openDB()
	defer closeDB()
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(path+"/assets/"))))
	mux.HandleFunc("/", TodoGet)
	mux.HandleFunc("/postTodo", TodoPost)
	mux.HandleFunc("/updateTodo/", TodoUpdate)
	mux.HandleFunc("/deleteTodo/", TodoDelete)

	err := http.ListenAndServe(":8000", mux)
	if err != nil {
		log.Fatal(err)
	}
}

type Todo struct {
	ID        int
	Todo      string
	Done      bool
	CreatedAt string
}

func TodoGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := DB.Query(`SELECT * FROM "Todo"`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Todo, &todo.Done, &todo.CreatedAt); err != nil {
			http.Error(w, "Error scanning todos", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating through todos", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles(path + "/index.html"))
	data := map[string][]Todo{
		"Todos": todos,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func TodoPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	todo := r.FormValue("todo")
	done := r.FormValue("done") == "on"

	_, err := DB.Exec(`INSERT INTO "Todo" (todo, done) VALUES ($1, $2)`, todo, done)
	if err != nil {
		http.Error(w, "Error inserting todo", http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.ParseFiles(path + "/index.html"))
	tmpl.ExecuteTemplate(w, "todos", Todo{Todo: todo, Done: done})
}

func TodoUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	id := r.URL.Path[len("/updateTodo/"):]
	todo := r.FormValue("todo")
	done := r.FormValue("done") == "on" // Checkbox value will be "on" if checked

	_, err = DB.Exec(`UPDATE "Todo" SET todo = $1, done = $2 WHERE id = $3`, todo, done, id)
	if err != nil {
		http.Error(w, "Error updating todo", http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.ParseFiles(path + "/index.html"))
	tmpl.ExecuteTemplate(w, "todos", Todo{Todo: todo, Done: done})
}

func TodoDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	id := r.URL.Path[len("/deleteTodo/"):]

	_, err = DB.Exec(`DELETE FROM "Todo" WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Error deleting todo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
