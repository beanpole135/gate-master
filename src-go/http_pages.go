package main

import (
	"fmt"
	"net/http"
	"time"
	"strconv"
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
	http.HandleFunc("/gate-open", checkToken(performGateOpen,true, false))
	// Accounts Tab
	http.HandleFunc("/page-accounts", checkToken(tab_accountsHandler, true, true))
	http.HandleFunc("/page-account-new", checkToken(tab_accountNewHandler, true, true))
	http.HandleFunc("/page-account-view", checkToken(tab_accountViewHandler, true, true))
	http.HandleFunc("/account-create", checkToken(performAccountCreate, true, true))
	// AccountCodes Tab
	http.HandleFunc("/page-accountcodes", checkToken(tab_accountcodesHandler, true, false))
	http.HandleFunc("/page-accountcode-new", checkToken(tab_accountcodeNewHandler, true, false))
	http.HandleFunc("/page-accountcode-view", checkToken(tab_accountcodeViewHandler, true, false))
	http.HandleFunc("/accountcode-create", checkToken(performAccountcodeCreate, true, false))
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
	acct, err := DB.AccountForUsernamePassword(user, passw)
	if err != nil {
		fmt.Println("Invalid Credentials:", err, user, passw)
		time.Sleep(2 * time.Second) //Delay a small amount on fails (help prevent brute-force attacks)
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
	uname := r.Form.Get("uname")
	fname := r.Form.Get("fname")
	lname := r.Form.Get("lname")
	newpw := r.Form.Get("newpw")
	newpw2 := r.Form.Get("newpw2")
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
	if newpw != newpw2 || newpw == "" {
		returnError(w, "New passwords do not match")
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
		FirstName: fname,
		LastName: lname,
		Username: uname,
		AccountStatus: accstatus,
		PwHash: hashPassword(newpw),
	}
	_, err := DB.AccountInsert(&acc)
	if err != nil {
		//Error
		returnError(w, "Internal error creating account")
		return
	}
	//Now reload the accounts page
	tab_accountsHandler(w, r, p)
}

func performAccountcodeCreate(w http.ResponseWriter, r *http.Request, p *Page) {
	// Parse the form
	r.ParseForm()
	code := r.Form.Get("code")
	label := r.Form.Get("label")
	is_utility := r.Form.Get("isutility") == formChecked
	is_delivery := r.Form.Get("isdelivery") == formChecked
	date_start := parseFormDate(r.Form.Get("dstart"))
	date_end := parseFormDate(r.Form.Get("dend"))
	time_start := parseFormTime(r.Form.Get("tstart"))
	time_end := parseFormTime(r.Form.Get("tend"))
	day_Su := r.Form.Get("d_sunday") == formChecked
	day_Mo := r.Form.Get("d_monday") == formChecked
	day_Tu := r.Form.Get("d_tuesday") == formChecked
	day_We := r.Form.Get("d_wednesday") == formChecked
	day_Th := r.Form.Get("d_thursday") == formChecked
	day_Fr := r.Form.Get("d_friday") == formChecked
	day_Sa := r.Form.Get("d_saturday") == formChecked

	// Validate the inputs
	if code == "" {
		returnError(w, "Missing PIN Code")
		return
	}
	if !validatePinCodeFormat(code) {
		returnError(w, "PIN code must be 4 or more numbers")
		return		
	}
	if label == "" {
		returnError(w, "Missing Description")
		return
	}

	if ac, _ := DB.AccountCodeMatch(code) ; ac != nil {
		returnError(w, "Invalid PIN - pick another one")
		return
	}

	if (time_start != nil && time_end == nil) || (time_start == nil && time_end != nil) {
		returnError(w, "Both start and end times must be provided, or neither of them")
	}
	// Create the new accountcode
	acc := AccountCode{
		IsActive: true,
		AccountID: int64(p.Token.UserId),
		Code: code,
		Label: label,
		IsUtility: is_utility,
		IsDelivery: is_delivery,
	}
	if day_Su { acc.ValidDays = append(acc.ValidDays, "su") }
	if day_Mo { acc.ValidDays = append(acc.ValidDays, "mo") }
	if day_Tu { acc.ValidDays = append(acc.ValidDays, "tu") }
	if day_We { acc.ValidDays = append(acc.ValidDays, "we") }
	if day_Th { acc.ValidDays = append(acc.ValidDays, "th") }
	if day_Fr { acc.ValidDays = append(acc.ValidDays, "fr") }
	if day_Sa { acc.ValidDays = append(acc.ValidDays, "sa") }
	if date_start != nil {
		acc.DateStart = *date_start
	}
	if date_end != nil {
		acc.DateEnd = *date_end
	}
	if time_start != nil {
		acc.TimeStart = *time_start
	}
	if time_end != nil {
		acc.TimeEnd = *time_end
	}
	_, err := DB.AccountCodeInsert(&acc)
	if err != nil {
		//Error
		returnError(w, "Internal error creating account")
		return
	}
	//Now reload the accountcodes page
	tab_accountcodesHandler(w, r, p)
}