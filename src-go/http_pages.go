package main

import (
	"fmt"
	"net/http"
	"time"
	"strconv"
)

func setupPages() {
	// Authentication Endpoints
	http.HandleFunc("/login", checkToken(loginPageHandler, false, false))
	http.HandleFunc("/auth-login", checkToken(performLoginHandler, false, false))
	http.HandleFunc("/auth-logout", checkToken(performLogoutHandler, true, false))
	// Main Page (parent of all tabs)
	http.HandleFunc("/gate", checkToken(gatePageHandler, true, false))
	// View Tab
	http.HandleFunc("/page-view", checkToken(tab_gateHandler, true, false))
	http.HandleFunc("/gate-open", checkToken(performGateOpen,true, false))
	// Accounts Tab
	http.HandleFunc("/page-accounts", checkToken(tab_accountsHandler, true, true))
	http.HandleFunc("/page-account-new", checkToken(tab_accountNewHandler, true, true))
	http.HandleFunc("/page-account-view", checkToken(tab_accountViewHandler, true, true))
	// Profile Tab
	http.HandleFunc("/page-profile", checkToken(tab_profileHandler, true, false))
	http.HandleFunc("/profile-update", checkToken(performProfileUpdate,true, false))
	http.HandleFunc("/profile-pwreset", checkToken(performProfilePWReset,true, false))
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
		returnError(w, "Invalid Credentials")
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
		returnError(w, "Internal Error")
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
	p.Accounts, _ = DB.AccountsSelectAll()
	renderTemplate(w, "tab_accounts", p)
}

func tab_accountNewHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	renderTemplate(w, "tab_account_new", p)
}

func tab_accountViewHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	//Parse the form
	r.ParseForm()
	accid := r.Form.Get("accid")
	id, err := strconv.Atoi(accid)
	if err != nil {
		//Invalid account ID
		returnError(w, "Invalid Account")
		return		
	}
	// Load the account from the DB
	p.Profile, err = DB.AccountFromID(int32(id))
	if err != nil {
		//Invalid account ID
		returnError(w, "Invalid Account")
		return
	}
	// Load additional account info here
	renderTemplate(w, "tab_account_view", p)
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
	r.ParseForm()
	fname := r.Form.Get("fname")
	lname := r.Form.Get("lname")
	if lname == "" || fname == "" {
		returnError(w, "Missing name(s)")
		return
	}
	//Update the account in the DB
	acc, err := DB.AccountFromID(p.Token.UserId)
	if err != nil {
		// Current user no longer exists?
		handleError(w, r)
		return		
	}
	acc.FirstName = fname
	acc.LastName = lname
	_, err = DB.AccountUpdate(acc)
	if err != nil {
		//Error
		returnError(w, "Internal error updating profile")
		return
	}
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
		returnError(w, "New password does not match")
		return
	}

	//Verify old password provided is accurate
	acc, err := DB.AccountFromID(p.Token.UserId)
	if err != nil {
		// Current user no longer exists?
		handleError(w, r)
		return		
	}
	acc2, err := DB.AccountForUsernamePassword(acc.Username, hashPassword(oldpw))
	if acc.AccountID != acc2.AccountID {
		//Error
		returnError(w, "Incorrect Password")
		return
	}	

	//Change password
	acc.PwHash = hashPassword(newpw)
	_, err = DB.AccountUpdate(acc)
	if err != nil {
		//Error
		returnError(w, "Internal error updating password")
		return
	}
	//Now reload the profile page
	tab_profileHandler(w, r, p)
}