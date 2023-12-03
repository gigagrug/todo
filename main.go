package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func main() {
	openDB()
	defer closeDB()
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	mux.HandleFunc("/", TodoGet)
	mux.HandleFunc("/postTodo", TodoPost)
	mux.HandleFunc("/updateTodo", TodoUpdate)
	mux.HandleFunc("/deleteTodo", TodoDelete)

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

	query := "SELECT * FROM Todo"
	rows, err := DB.Query(query)
	if err != nil {
		http.Error(w, "Error querying todos", http.StatusInternalServerError)
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

	tmpl := template.Must(template.ParseFiles("index.html"))
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

	insertTodo := "INSERT INTO Todo (todo, done) VALUES (?, ?)"
	_, err := DB.Exec(insertTodo, todo, done)
	if err != nil {
		http.Error(w, "Error inserting todo", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Todo added successfully!")
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

	todoID := r.FormValue("todoID")
	todo := r.FormValue("todo")
	done := r.FormValue("done") == "on" // Checkbox value will be "on" if checked

	updateTodo := "UPDATE Todo SET todo = ?, done = ? WHERE id = ?"
	_, err = DB.Exec(updateTodo, todo, done, todoID)
	if err != nil {
		http.Error(w, "Error updating todo", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Todo updated successfully!")
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

	todoID := r.FormValue("todoID") // Assuming todoID is passed as a hidden field in the form

	deleteTodo := "DELETE FROM Todo WHERE id = ?"
	_, err = DB.Exec(deleteTodo, todoID)
	if err != nil {
		http.Error(w, "Error deleting todo", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Todo deleted successfully!")
}
