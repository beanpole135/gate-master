package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

// staticFS is our static web server content.
//
//go:embed static/*
var staticFS embed.FS

// htmlFS is our static html file templates
//
//go:embed html/*
var htmlFS embed.FS

type Page struct {
	Title        string
	Token        *AuthToken
	Profile      *Account
	Accounts     []Account
	AccountCodes []AccountCode
	AccountCode  AccountCode
	GateLogs     []GateLog
	GateLog      *GateLog
	Contacts     []Contact
	Contact      *Contact
}

var templates *template.Template
var CAM *Camera
var DB *Database
var CONFIG *Config

func exitErr(err error, msg string) {
	if err != nil {
		fmt.Println(fmt.Sprintf(msg, err))
		os.Exit(1)
	}
}

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
	var err error
	//Load the config file
	conffile := "config.json"
	if len(os.Args) > 1 {
		conffile = os.Args[1]
	}
	CONFIG, err = LoadConfig(conffile)
	if err != nil {
		fmt.Println("Could not read config - using defaults:", err)
	}
	//Load the HTML templates (built-in)
	templates, err = template.ParseFS(htmlFS, "html/*.html")
	exitErr(err, "Could not load Templates: %v")

	//Setup the Camera
	CAM, err = NewCamera()
	exitErr(err, "Could not create Camera: %v")
	defer CAM.Close()

	//Setup the Database
	DB, err = NewDatabase(CONFIG.DbFile)
	exitErr(err, "Could not create database: %v")
	defer DB.Close()

	CONFIG.Keypad.StartWatching()

	//Setup the HTTP auth system
	setupSecureCookies()

	//Setup the pages / endpoints
	http.HandleFunc("/favicon.ico", favicon)
	http.Handle("/static/", http.StripPrefix("/", http.FileServer(http.FS(staticFS))))
	http.HandleFunc("/stream", checkToken(CAM.ServeImages, true, false))

	// Individual Pages
	setupPages()
	// Final setup
	go DB.PruneTables() //Runs the pruning checks every day

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
			Title: CONFIG.SiteName,
		}
		if validateToken {
			toks := ReadSecureCookieTokens(w, r)
			if toks == "" {
				fmt.Println("Cannot read token from cookie")
				handleError(w, r)
				return
			}
			tok, err := ReadSignedToken(toks, CONFIG.Auth.JwtSecret, true)
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
	msg = strings.ReplaceAll(msg, "\"", "\\\"")
	w.Header().Add("HX-Trigger", fmt.Sprintf(`{"showError": "%s"}`, msg))
	http.Error(w, msg, http.StatusBadRequest)
}

func returnSuccess(w http.ResponseWriter, msg string) {
	msg = strings.ReplaceAll(msg, "\"", "\\\"")
	w.Header().Add("HX-Trigger", fmt.Sprintf(`{"showSuccess": "%s"}`, msg))
	http.Error(w, msg, http.StatusBadRequest)
}

// Simple randomization functions
const letterBytes = "abcdefghikmnopqrstuvwxyzABCDEFGHJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	fmt.Println("Generated Random String:", string(b))
	return string(b)
}

const numberBytes = "0123456789"

func RandomPIN(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = numberBytes[rand.Intn(len(numberBytes))]
	}
	fmt.Println("Generated Random PIN:", string(b))
	return string(b)
}
