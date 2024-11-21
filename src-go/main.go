package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
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

var templates *template.Template
var Cam *Camera
var DB *Database

func exitErr(err error, msg string) {
	if err != nil {
		fmt.Println(fmt.Sprintf(msg, err))
		os.Exit(1)
	}
}

func main() {
	var err error
	templates, err = template.ParseFS(htmlFS, "html/*.html")
	exitErr(err, "Could not load Templates: %v")

	//Setup the Camera
	Cam, err := NewCamera()
	exitErr(err, "Could not create Camera: %v")
	defer Cam.Close()

	//Setup the Database
	DB, err = NewDatabase("test.sqlite")
	exitErr(err, "Could not create database: %v")
	defer DB.Close()

	//Sample to test the database system
	/*A := Account{
		FirstName: "FTest",
		LastName: "LTest",
		Username: "TESTUSER",
		PwHash: "password",
	}
	acc, err := DB.AccountInsert(&A)
	exitErr(err, "Cannot insert DB record: %v")

	a, err := DB.AccountsSelectAll()
	exitErr(err, "Could not load accounts: %v")
	if len(a) < 1 {
		fmt.Println("Account not loaded")
		os.Exit(1)
	}
	fmt.Println(fmt.Sprintf("Got Accounts: %v", a))
	a[0].FirstName = "FName2"
	acc, err = DB.AccountUpdate(&a[0])
	exitErr(err, "Could not update account: %v")

	acc, err = DB.AccountForUsernamePassword("TESTUSER","password")
	exitErr(err, "Could not find username with password: %v")
	fmt.Println(fmt.Sprintf("Got account %v", acc))
	*/

	//Setup the pages / endpoints
	http.Handle("/static/", http.StripPrefix("/", http.FileServer(http.FS(staticFS))))
	http.HandleFunc("/login", makeHandler(loginHandler))
	http.HandleFunc("/stream", Cam.ServeImages)
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