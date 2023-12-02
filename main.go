package main

import (
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

func DBTodoGet() ([]Todo, error) {
	query := "SELECT * FROM Todo"
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Todo, &todo.Done, &todo.CreatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}
func TodoGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		println("bad", r.Method)
		return
	}
	todos, err := DBTodoGet()
	if err != nil {
		log.Panic(err)
	}
	println("good", r.Method)
	tmpl := template.Must(template.ParseFiles("index.html"))

	data := map[string][]Todo{
		"Todos": todos,
	}
	tmpl.Execute(w, data)
}
