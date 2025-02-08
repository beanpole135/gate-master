package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const formChecked = "on"

func setupPages() {
	// Authentication Endpoints
	http.HandleFunc("/login", checkToken(loginPageHandler, false, false))
	http.HandleFunc("/auth-login", checkToken(performLoginHandler, false, false))
	http.HandleFunc("/auth-logout", checkToken(performLogoutHandler, true, false))
	// Main Page (parent of all tabs)
	http.HandleFunc("/gate", checkToken(gatePageHandler, true, false))
	// View Tab
	http.HandleFunc("/page-view", checkToken(tab_gateHandler, true, false))
	http.HandleFunc("/gate-open", checkToken(performGateOpen, true, false))
	// Accounts Tab
	http.HandleFunc("/page-accounts", checkToken(tab_accountsHandler, true, true))
	http.HandleFunc("/page-account-new", checkToken(tab_accountNewHandler, true, true))
	http.HandleFunc("/page-account-view", checkToken(tab_accountViewHandler, true, true))
	http.HandleFunc("/account-create", checkToken(performAccountCreate, true, true))
	http.HandleFunc("/account-update", checkToken(performAccountUpdate, true, true))
	// AccountCodes Tab
	http.HandleFunc("/page-accountcodes", checkToken(tab_accountcodesHandler, true, false))
	http.HandleFunc("/page-accountcode-new", checkToken(tab_accountcodeNewHandler, true, false))
	http.HandleFunc("/page-accountcode-view", checkToken(tab_accountcodeViewHandler, true, false))
	http.HandleFunc("/accountcode-create", checkToken(performAccountcodeCreate, true, false))
	http.HandleFunc("/accountcode-update", checkToken(performAccountcodeUpdate, true, false))
	// Contacts Tab
	http.HandleFunc("/page-contacts", checkToken(tab_contactsHandler, true, false))
	http.HandleFunc("/page-contact-new", checkToken(tab_contactNewHandler, true, false))
	http.HandleFunc("/page-contact-view", checkToken(tab_contactViewHandler, true, false))
	http.HandleFunc("/contact-create", checkToken(performContactCreate, true, false))
	http.HandleFunc("/contact-update", checkToken(performContactUpdate, true, false))
	http.HandleFunc("/contact-test", checkToken(performContactTest, true, false))

	// Profile Tab
	http.HandleFunc("/page-profile", checkToken(tab_profileHandler, true, false))
	http.HandleFunc("/profile-update", checkToken(performProfileUpdate, true, false))
	http.HandleFunc("/profile-pwupdate", checkToken(performProfilePWChange, true, false))
	// View Tab
	http.HandleFunc("/page-logs", checkToken(tab_logsHandler, true, false))
	http.HandleFunc("/page-log-view", checkToken(tab_logViewHandler, true, false))

}

// Page Handlers
// Make sure to add new pages to setupPages() above
func loginPageHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	p.Title = fmt.Sprintf("%s Login", CONFIG.SiteName)
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
	acct, err := DB.AccountForUsernamePassword(user, passw)
	if err != nil {
		fmt.Println("Invalid Credentials:", err, user, passw)
		time.Sleep(2 * time.Second) //Delay a small amount on fails (help prevent brute-force attacks)
		returnError(w, "Invalid Credentials")
		return
	}
	// Create Token
	at := AuthToken{
		UserId:  int32(acct.AccountID),
		IsAdmin: (acct.AccountStatus == Account_Admin),
	}
	tok, err := CreateSignedToken(at, CONFIG.Auth.JwtSecret, CONFIG.Auth.JwtTokenSecs)
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
	renderTemplate(w, "post-login", p)
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
	p.Contacts, err = DB.ContactsForAccount(int64(id))
	// Now render the page
	renderTemplate(w, "tab_account_view", p)
}

func tab_accountcodesHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	p.AccountCodes, _ = DB.AccountCodeSelectAll(p.Token.UserId, 0) //all codes for current user
	renderTemplate(w, "tab_accountcodes", p)
}

func tab_accountcodeNewHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	renderTemplate(w, "tab_accountcode_new", p)
}

func tab_accountcodeViewHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	//Parse the form
	r.ParseForm()
	accid := r.Form.Get("accid")
	id, err := strconv.Atoi(accid)
	if err != nil {
		//Invalid account ID
		returnError(w, "Invalid Account Code")
		return
	}
	// Load the accountcode from the DB
	list, err := DB.AccountCodeSelectAll(p.Token.UserId, int64(id))
	if err != nil || len(list) < 1 {
		//Invalid account ID
		returnError(w, "Invalid Account Code")
		return
	}
	p.AccountCode = list[0]
	// Load additional account info here
	renderTemplate(w, "tab_accountcode_view", p)
}

func tab_profileHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	// Load the info about the current profile into the page struct
	var err error
	p.Profile, err = DB.AccountFromID(p.Token.UserId)
	if err != nil || p.Profile == nil {
		tab_accountsHandler(w, r, p)
		return
	}
	// Load additional account info here
	p.Contacts, err = DB.ContactsForAccount(int64(p.Token.UserId))

	// Now render the page
	renderTemplate(w, "tab_profile", p)
}

func tab_logsHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	p.GateLogs, _ = DB.GatelogSelectAll()
	renderTemplate(w, "tab_logs", p)
}

func tab_logViewHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	//Parse the form
	r.ParseForm()
	logid := r.Form.Get("logid")
	id, err := strconv.Atoi(logid)
	if err != nil {
		//Invalid log ID
		returnError(w, "Invalid log Code")
		return
	}
	// Load the log from the DB
	p.GateLog, err = DB.GateLogFromID(int64(id))
	if err != nil || p.GateLog == nil {
		//Invalid account ID
		returnError(w, "Invalid Log ID")
		return
	}
	renderTemplate(w, "tab_log_view", p)
}

func tab_contactsHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	var err error
	p.Contacts, err = DB.ContactsForAccount(int64(p.Token.UserId)) //all contacts for current user
	if err != nil {
		fmt.Println("Got error reading contacts for Account:", err)
	}
	renderTemplate(w, "tab_contacts", p)
}

func tab_contactNewHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	renderTemplate(w, "tab_contact_new", p)
}

func tab_contactViewHandler(w http.ResponseWriter, r *http.Request, p *Page) {
	//Parse the form
	r.ParseForm()
	cid := r.Form.Get("contactid")
	id, err := strconv.Atoi(cid)
	if err != nil {
		//Invalid account ID
		returnError(w, "Invalid Contact ID")
		return
	}
	//Grab the contact
	p.Contact, err = DB.ContactFromID(int64(id))
	//Verify contact ID matches the current user
	if err != nil || p.Contact.AccountID != p.Token.UserId {
		returnError(w, "Invalid Contact ID")
		return
	}
	renderTemplate(w, "tab_contact_view", p)
}

func performGateOpen(w http.ResponseWriter, r *http.Request, p *Page) {
	fmt.Println("Gate Opening!")
	acc, err := DB.AccountFromID(p.Token.UserId)
	if err != nil {
		// Current user no longer exists?
		handleError(w, r)
		return
	}
	err = OpenGateAndNotify(acc, nil)
	returnSuccess(w, "Gate Opening!")
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

func performProfilePWChange(w http.ResponseWriter, r *http.Request, p *Page) {
	//Parse the form
	r.ParseForm()
	oldpw := r.Form.Get("oldpw")
	newpw := r.Form.Get("newpw")
	newpw2 := r.Form.Get("newpw2")

	if ok, reasons := validatePasswordFormat(newpw); !ok {
		returnError(w, fmt.Sprintf("Invalid Password format: %s", reasons))
		return
	}
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
	acc2, err := DB.AccountForUsernamePassword(acc.Username, oldpw)
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

func performAccountCreate(w http.ResponseWriter, r *http.Request, p *Page) {
	// Parse the form
	r.ParseForm()
	uname := r.Form.Get("uname") //This is expected to be an email address
	fname := r.Form.Get("fname")
	lname := r.Form.Get("lname")
	newpw := "TEST1234" //randomize this later
	//newpw2 := r.Form.Get("newpw2")
	isadmin := r.Form.Get("isadmin") == formChecked
	// Validate the inputs
	if fname == "" || lname == "" {
		returnError(w, "Missing first/last name(s)")
		return
	}
	if ok, reasons := validatePasswordFormat(newpw); !ok {
		returnError(w, fmt.Sprintf("Invalid Password format: %s", reasons))
		return
	}
	if uname == "" || DB.AccountExists(uname) {
		returnError(w, "Invalid username")
		return
	}
	accstatus := Account_Active
	if isadmin || p.Token.UserId == -1 {
		//First user created must be an admin
		accstatus = Account_Admin
	}
	// Create the new account
	acc := Account{
		FirstName:     fname,
		LastName:      lname,
		Username:      uname,
		AccountStatus: accstatus,
		PwHash:        hashPassword(newpw),
	}
	nacc, err := DB.AccountInsert(&acc)
	if err != nil {
		//Error
		returnError(w, "Internal error creating account")
		return
	}
	//Create the primary contact for this account
	ct := Contact{
		AccountID: nacc.AccountID,
		Email:     uname,
		IsPrimary: true,
		IsActive:  true,
	}
	_, err = DB.ContactInsert(&ct)
	if err != nil {
		returnError(w, "Internal error creating account")
		return
	}
	//Send an email containing the new password to the account holder
	CONFIG.Email.SendEmail(
		uname,
		"New "+CONFIG.SiteName+" Account",
		fmt.Sprintf("%s has just created an account for you at %s.\nPlease login and change your password as soon as possible.\n\nYour temporary password is:  %s",
			nacc.FirstName+" "+nacc.LastName,
			CONFIG.Host,
			newpw,
		),
		true,
	)
	//Now reload the accounts page
	tab_accountsHandler(w, r, p)
}

func performAccountUpdate(w http.ResponseWriter, r *http.Request, p *Page) {
	// Parse the form
	r.ParseForm()
	accid := r.Form.Get("accid")
	fname := r.Form.Get("fname")
	lname := r.Form.Get("lname")
	status := r.Form.Get("status")
	// Validate the inputs
	if fname == "" || lname == "" {
		returnError(w, "Missing first/last name(s)")
		return
	}
	accnum, err := strconv.Atoi(accid)
	if accid == "" || err != nil {
		returnError(w, "Invalid Account")
		return
	}
	//Fetch the account from the DB
	acc, err := DB.AccountFromID(int32(accnum))
	if err != nil {
		returnError(w, "Invalid Account")
		return
	}
	// Update the account fields
	switch status {
	case "active":
		acc.AccountStatus = Account_Active
	case "inactive":
		acc.AccountStatus = Account_Inactive
	case "admin":
		acc.AccountStatus = Account_Admin
	default:
		returnError(w, "Invalid account status")
		return
	}
	acc.FirstName = fname
	acc.LastName = lname

	_, err = DB.AccountUpdate(acc)
	if err != nil {
		//Error
		returnError(w, "Internal error updating account")
		return
	}
	//Now reload the accounts page
	tab_accountsHandler(w, r, p)
}

func performAccountcodeCreate(w http.ResponseWriter, r *http.Request, p *Page) {
	// Load the form input into the AccountCode
	acc, err := LoadAccountCodeFromForm(r)
	if err != nil {
		returnError(w, err.Error())
		return
	}
	acc.AccountID = p.Token.UserId //Always associate new PIN with current user account
	acc.IsActive = true            //new PINs are always active initially

	// New Code Validation
	if ac, _ := DB.AccountCodeMatch(acc.Code); ac != nil {
		returnError(w, "Invalid PIN - pick another one")
		return
	}

	// Create the new code
	_, err = DB.AccountCodeInsert(&acc)
	if err != nil {
		//Error
		returnError(w, "Internal error creating account")
		return
	}
	//Now reload the accountcodes page
	tab_accountcodesHandler(w, r, p)
}

func performAccountcodeUpdate(w http.ResponseWriter, r *http.Request, p *Page) {

	// Load the form input into the AccountCode
	acc, err := LoadAccountCodeFromForm(r)
	if err != nil {
		returnError(w, err.Error())
		return
	}
	acc.AccountID = p.Token.UserId

	// Create the new code
	_, err = DB.AccountCodeUpdate(&acc)
	if err != nil {
		//Error
		returnError(w, "Internal error updating PIN")
		return
	}
	//Now reload the accountcodes page
	tab_accountcodesHandler(w, r, p)
}

func performContactCreate(w http.ResponseWriter, r *http.Request, p *Page) {
	// Load the form input into the AccountCode
	ct, err := LoadContactFromForm(r)
	if err != nil {
		returnError(w, err.Error())
		return
	}
	ct.AccountID = p.Token.UserId //Always associate new Contact with current user account
	ct.IsActive = true            //new contacts are always active initially

	// Create the new code
	_, err = DB.ContactInsert(&ct)
	if err != nil {
		//Error
		returnError(w, "Internal error creating contact")
		return
	}
	//Now reload the accountcodes page
	tab_contactsHandler(w, r, p)
}

func performContactUpdate(w http.ResponseWriter, r *http.Request, p *Page) {
	// Load the form input into the AccountCode
	ct, err := LoadContactFromForm(r)
	if err != nil {
		returnError(w, err.Error())
		return
	}
	ct.AccountID = p.Token.UserId

	// Create the new code
	_, err = DB.ContactUpdate(&ct)
	if err != nil {
		//Error
		returnError(w, "Internal error updating contact")
		return
	}
	//Now reload the accountcodes page
	tab_contactsHandler(w, r, p)
}

func performContactTest(w http.ResponseWriter, r *http.Request, p *Page) {
	//Parse the form
	r.ParseForm()
	cid := r.Form.Get("contactid")
	id, err := strconv.Atoi(cid)
	if err != nil {
		//Invalid account ID
		returnError(w, "Invalid Contact ID")
		return
	}
	//Grab the contact
	p.Contact, err = DB.ContactFromID(int64(id))
	//Verify contact ID matches the current user
	if err != nil || p.Contact.AccountID != p.Token.UserId {
		returnError(w, "Invalid Contact ID")
		return
	}
	//Now send a test email to that contact
	err = CONFIG.Email.SendEmail(
		p.Contact.ContactEmail(),
		"Notification Test",
		fmt.Sprintf("This is a test of the %s notification system", CONFIG.SiteName),
		false,
	)
	if err == nil {
		returnSuccess(w, "Test Sent")
	} else {
		returnError(w, "Test Failed to Send: "+err.Error())
	}
}
