package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

const PageTitle = "Shadow Mountain"
const jwtsecretkey = "testkey" //not critical since we encrypt over the top of this as well
const tokenLifeS = 3600 // 1 hour lifetime

// staticFS is our static web server content.
//go:embed static/*
var staticFS embed.FS

// htmlFS is our static html file templates
//go:embed html/*
var htmlFS embed.FS

type Page struct {
	Title string
	Token *AuthToken
	Profile *Account
	Accounts []Account
	AccountCodes []AccountCode
	AccountCode AccountCode
	GateLogs []GateLog
	GateLog *GateLog
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

	//Setup the HTTP auth system
	setupSecureCookies()

	//Setup the pages / endpoints
	http.HandleFunc("/favicon.ico", favicon)
	http.Handle("/static/", http.StripPrefix("/", http.FileServer(http.FS(staticFS))))
	http.HandleFunc("/stream", checkToken(Cam.ServeImages, true, false))
	
	// Individual Pages
	setupPages()
	// Final setup
	http.HandleFunc("/", handleError)
	fmt.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func favicon(w http.ResponseWriter, r *http.Request) {
	data, err := staticFS.ReadFile("static/favicon.ico")
    if err != nil { 
    	fmt.Println("favicon.ico loading error", err)
    }
    http.ServeContent(w, r, "favicon.ico", time.Now(), bytes.NewReader(data))
}

// Internal functions for loading pages
func triggerPageRefreshOnLoad(w http.ResponseWriter) {
	w.Header().Add("HX-Refresh", "true")
}
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleError(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func checkToken(fn func(http.ResponseWriter, *http.Request, *Page), validateToken bool, adminonly bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := &Page{
			Title: PageTitle,
		}
		if validateToken {
			toks := ReadSecureCookieTokens(w, r)
			if toks == "" {
				fmt.Println("Cannot read token from cookie")
				handleError(w, r)
				return
			}
			tok, err := ReadSignedToken(toks, jwtsecretkey, true)
			if err != nil {
				fmt.Println("Got Token verify error:", err)
				handleError(w, r)
				return
			}
			if adminonly && !tok.IsAdmin {
				handleError(w, r)
				return
			}
			p.Token = &tok
		}
		//Now load the page
		fn(w, r, p)
	}
}

func returnError(w http.ResponseWriter, msg string) {
	fmt.Println("Got error to return:", msg)
	w.Header().Add("HX-Trigger", fmt.Sprintf(`{"showError": "%s"}`, msg))
	http.Error(w, msg, http.StatusBadRequest)
}
