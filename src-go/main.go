package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

const PageTitle = "Shadow Mountain"

// staticFS is our static web server content.
//go:embed static/*
var staticFS embed.FS

// htmlFS is our static html file templates
//go:embed html/*
var htmlFS embed.FS

type Page struct {
	Title string
}

/*func (p *Page) save() error {
	filename := p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}*/

var templates *template.Template

func main() {
	var err error
	templates, err = template.ParseFS(htmlFS, "html/*.html")
	if err != nil {
		fmt.Println("Could not load Templates:", err)
		return
	}
	http.Handle("/static/", http.StripPrefix("/", http.FileServer(http.FS(staticFS))))
	http.HandleFunc("/login", makeHandler(loginHandler))
	http.HandleFunc("/favicon.ico", iconHandler)
	fmt.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Internal functions for loading pages
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Put the JWT check and load here as needed
		// - if error, redirect back to /login page
		//Now load the page
		fn(w, r, PageTitle)
	}
}

// Individual Page Handlers
func iconHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/static/favicon.png", http.StatusSeeOther)
}
func loginHandler(w http.ResponseWriter, r *http.Request, title string) {
	p := &Page{Title: fmt.Sprintf("%s Login", title)}
	renderTemplate(w, "login", p)
}