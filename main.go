package main

import (
	"fmt"
	"net/http"
	"html/template"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", root)
	server := &http.Server{
		Addr: "127.0.0.1:8080",
		Handler: mux,
	}
	fmt.Println("Server started.")
	server.ListenAndServe()
}

func root(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./templates/root.html")
	t.Execute(w, nil)
}
