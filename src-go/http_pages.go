package main

import (
	"fmt"
	"net/http"
)

func setupPages() {
	http.HandleFunc("/login", checkToken(loginPageHandler, false))
	http.HandleFunc("/auth-login", checkToken(performLoginHandler, false))
	http.HandleFunc("/auth-logout", checkToken(performLogoutHandler, true))
	http.HandleFunc("/gate", checkToken(gatePageHandler, true))
	http.HandleFunc("/page-view", checkToken(tab_gateHandler, true))
	http.HandleFunc("/page-accounts", checkToken(tab_accountsHandler, true))
	http.HandleFunc("/page-profile", checkToken(tab_profileHandler, true))
}

// Page Handlers
// Make sure to add new pages to setupPages() above
func loginPageHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	p.Title = fmt.Sprintf("%s Login", PageTitle)
	triggerPageRefreshOnLoad(w)
	renderTemplate(w, "login", p)
}

func performLoginHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	// This performs the actual login process and then redirects to the main page when done
	// Read username/password from request
	r.ParseForm()
	user := r.Form.Get("uname")
	passw := r.Form.Get("passw")	
	// Verify validity
	pwhash := hashPassword(passw)
	acct, err := DB.AccountForUsernamePassword(user, pwhash)
	if err != nil {
		fmt.Println("Invalid Credentials:", err)
		handleError(w, r)
		return
	}
	// Create Token
	at := AuthToken{
		UserId: int32(acct.AccountID),
		IsAdmin: (acct.AccountStatus == Account_Admin),
	}
	tok, err := CreateSignedToken(at, jwtsecretkey, tokenLifeS)
	if err != nil {
		fmt.Println("Cannot create signed token:", err)
		handleError(w, r)
		return
	}
	fmt.Println("Successful login:", user)
	SetTokenCookie(w, r, tok)
	// Redirect over to the main page
	http.Redirect(w, r, "/gate", http.StatusSeeOther)
}

func performLogoutHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	// Invalidate and clear out the cookies/token
	DeleteTokenCookie(w, r)
	// Redirect back to the login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func gatePageHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	renderTemplate(w, "gate", p)
}

func tab_gateHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	fmt.Println("Gate Tab Handler")
	renderTemplate(w, "tab_gate", p)
}

func tab_accountsHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	fmt.Println("Gate Tab Handler")
	renderTemplate(w, "tab_accounts", p)
}

func tab_profileHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	fmt.Println("Gate Tab Handler")
	renderTemplate(w, "tab_profile", p)
}