package main

import (
	"html/template"
	"net/http"
)

type LayoutData struct {
    Title string
    Data  any
}

var tpl *template.Template

func main() {
	tpl, _ = template.ParseGlob("../../web/templates/*.html")
	http.HandleFunc("/chores", choresHandleFunc)
	http.HandleFunc("/groceries", groceriesHandleFunc)
	http.HandleFunc("/", indexHandleFunc)
	http.HandleFunc("/login", loginHandleFunc)
	http.HandleFunc("/logout", loginHandleFunc)
	http.HandleFunc("/recipes", recipesHandleFunc)
	fs := http.FileServer(http.Dir("../../web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.ListenAndServe(":8080", nil)
}

func choresHandleFunc(w http.ResponseWriter, r *http.Request) {
	data := LayoutData{Title: "Chores", Data: nil}
	tpl.ExecuteTemplate(w, "chores.html", data)
}
func groceriesHandleFunc(w http.ResponseWriter, r *http.Request) {
	data := LayoutData{Title: "Groceries", Data: nil}
	tpl.ExecuteTemplate(w, "groceries.html", data)
}
func indexHandleFunc(w http.ResponseWriter, r *http.Request) {
	data := LayoutData{Title: "Home", Data: nil}
	tpl.ExecuteTemplate(w, "index.html", data)
}
func loginHandleFunc(w http.ResponseWriter, r *http.Request) {
	data := LayoutData{Title: "login", Data: nil}
	tpl.ExecuteTemplate(w, "login.html", data)
}
func recipesHandleFunc(w http.ResponseWriter, r *http.Request) {
	data := LayoutData{Title: "Recipes", Data: nil}
	tpl.ExecuteTemplate(w, "recipes.html", data)
}
