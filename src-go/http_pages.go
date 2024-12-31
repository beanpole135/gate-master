package main

import (
	"fmt"
	"net/http"
	"time"
)

func setupPages() {
	http.HandleFunc("/login", checkToken(loginPageHandler, false))
	http.HandleFunc("/auth-login", checkToken(performLoginHandler, false))
	http.HandleFunc("/auth-logout", checkToken(performLogoutHandler, true))
	http.HandleFunc("/gate", checkToken(gatePageHandler, true))
	http.HandleFunc("/page-view", checkToken(tab_gateHandler, true))
	http.HandleFunc("/page-accounts", checkToken(tab_accountsHandler, true))
	http.HandleFunc("/page-profile", checkToken(tab_profileHandler, true))
	http.HandleFunc("/gate-open", checkToken(performGateOpen,true))
	http.HandleFunc("/profile-update", checkToken(performProfileUpdate,true))
	http.HandleFunc("/profile-pwreset", checkToken(performProfilePWReset,true))
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
		time.Sleep(3 * time.Second) //Delay a small amount on fails (help prevent brute-force attacks)
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
	renderTemplate(w, "tab_gate", p)
}

func tab_accountsHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	fmt.Println("Accounts Tab Handler")
	renderTemplate(w, "tab_accounts", p)
}

func tab_profileHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	// Load the info about the current profile into the page struct
	var err error
	p.Profile, err = DB.AccountFromID(p.Token.UserId)
	if err != nil || p.Profile == nil {
		tab_accountsHandler(w, r, p)
		return
	}
	// Now render the page
	renderTemplate(w, "tab_profile", p)
}

func performGateOpen(w http.ResponseWriter, r *http.Request, p *Page) {
	fmt.Println("Gate Opening!")
	renderTemplate(w, "button_gate_opening", p)
}

func performProfileUpdate(w http.ResponseWriter, r *http.Request, p *Page) {
	//Parse the form
	/*r.ParseForm()
	user := r.Form.Get("uname")
	passw := r.Form.Get("passw")*/
	//Update the account in the DB

	//Now reload the profile page
	tab_profileHandler(w, r, p)
}

func performProfilePWReset(w http.ResponseWriter, r *http.Request, p *Page) {
	//Parse the form
	r.ParseForm()
	oldpw := r.Form.Get("oldpw")
	newpw := r.Form.Get("newpw")
	newpw2 := r.Form.Get("newpw2")
	
	if newpw != newpw2 {
		//Error

	}

	//Verify old password provided is accurate
	acc, err := DB.AccountFromID(p.Token.UserId)
	if err != nil {

	}
	acc2, err := DB.AccountForUsernamePassword(acc.Username, hashPassword(oldpw))
	if acc.AccountID != acc2.AccountID {
		//Error

	}	

	//Change password
	acc.PwHash = hashPassword(newpw)
	_, err = DB.AccountUpdate(acc)
	if err != nil {
		//Error

	}
	//Now reload the profile page
	tab_profileHandler(w, r, p)
}