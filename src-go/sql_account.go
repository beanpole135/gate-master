package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// NOTE: "username" is expected to be an email address
// so we always lowercase it before we save/read it from the database

const (
	Account_Active   = 1
	Account_Inactive = 2
	Account_Admin    = 3
)

func hashPassword(pw string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}

func validatePasswordFormat(pw string) (bool, string) {
	var reasons []string
	if len(pw) < 3 {
		reasons = append(reasons, "3+ character minimum")
	}
	if strings.Contains(pw, " ") {
		reasons = append(reasons, "no spaces")
	}
	if len(reasons) > 0 {
		return false, strings.Join(reasons, ", ")
	}
	return true, ""
}

type Account struct {
	AccountID     int32
	FirstName     string
	LastName      string
	Username      string
	PwHash        string
	TempPwHash    string
	AccountStatus int
	TimeCreated   time.Time
	TimeModified  time.Time
}

func (A Account) Status() string {
	switch A.AccountStatus {
	case Account_Active:
		return "Active"
	case Account_Inactive:
		return "Inactive"
	case Account_Admin:
		return "Admin"
	default:
		return "Unknown"
	}
}

func (A Account) StatusValue() string {
	return strings.ToLower(A.Status())
}

func DefaultAdminAccount() *Account {
	return &Account{
		AccountID:     -1,
		FirstName:     "admin",
		LastName:      "admin",
		Username:      "admin",
		AccountStatus: Account_Admin,
		TimeCreated:   time.Now(),
		TimeModified:  time.Now(),
	}
}

func (D *Database) CreateAccTable() error {
	q := `create table if not exists account (
account_id integer primary key autoincrement,
first_name text not null,
last_name text not null,
username text not null unique,
pw_hash text not null,
temp_pw_hash text not null,
account_status integer not null,
time_created integer not null,
time_modified integer not null
	);`
	_, err := D.ExecSql(q)
	return err
}

// internal function to read the rows from the account table
// NOTE: pw_hash is never returned!!
var accountSelect = `select account_id, first_name, last_name, username, account_status, time_created, time_modified
	from account`

var fullaccountSelect = `select account_id, first_name, last_name, username, account_status, time_created, time_modified, pw_hash, temp_pw_hash
	from account`

func (D *Database) parseAccountRows(rows *sql.Rows) ([]Account, error) {
	defer rows.Close()
	var accounts []Account
	var t_created, t_mod int64
	for rows.Next() {
		var acc Account
		if err := rows.Scan(&acc.AccountID, &acc.FirstName, &acc.LastName, &acc.Username, &acc.AccountStatus, &t_created, &t_mod); err != nil {
			return accounts, err
		}
		acc.TimeCreated = D.ParseTime(t_created)
		acc.TimeModified = D.ParseTime(t_mod)
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (D *Database) parseFullAccountRows(rows *sql.Rows) ([]Account, error) {
	defer rows.Close()
	var accounts []Account
	var t_created, t_mod int64
	for rows.Next() {
		var acc Account
		if err := rows.Scan(&acc.AccountID, &acc.FirstName, &acc.LastName, &acc.Username, &acc.AccountStatus, &t_created, &t_mod, &acc.PwHash, &acc.TempPwHash); err != nil {
			return accounts, err
		}
		acc.TimeCreated = D.ParseTime(t_created)
		acc.TimeModified = D.ParseTime(t_mod)
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (D *Database) AccountInsert(acc *Account) (*Account, error) {
	if acc.AccountStatus < 1 {
		acc.AccountStatus = Account_Active
	}

	q := `insert into account (first_name, last_name, username, pw_hash, temp_pw_hash, account_status, time_created, time_modified) values
		(?, ?, ?, ?, ?, ?, ?, ?)
		returning account_id;`
	rslt, err := D.ExecSql(q, acc.FirstName, acc.LastName, strings.ToLower(acc.Username), acc.PwHash, acc.TempPwHash, acc.AccountStatus, D.TimeNow(), D.TimeNow())
	if err != nil {
		fmt.Println("Error Inserting Account:", err)
		return nil, err
	}
	recordId, err := rslt.LastInsertId()
	acc.AccountID = int32(recordId)
	return acc, err
}

func (D *Database) AccountUpdate(acc *Account) (*Account, error) {
	if acc.AccountID < 1 {
		return nil, fmt.Errorf("Missing Account ID for UpdateAccount")
	}
	acc.TimeModified = time.Now()
	var err error
	if acc.PwHash != "" {
		// Setting a new password - clear temporary password if one is set
		q := `update account set 
		first_name = ?,
		last_name = ?,
		username = ?,
		pw_hash = ?,
		temp_pw_hash = '',
		account_status = ?,
		time_modified = ?
		where account_id = ?;`
		_, err = D.ExecSql(q, acc.FirstName, acc.LastName, strings.ToLower(acc.Username), acc.PwHash, acc.AccountStatus, D.TimeNow(), acc.AccountID)

	} else if acc.TempPwHash != "" {
		// Adding a temporary password (do not change current password hash!)
		q := `update account set 
		first_name = ?,
		last_name = ?,
		username = ?,
		temp_pw_hash = ?,
		account_status = ?,
		time_modified = ?
		where account_id = ?;`
		_, err = D.ExecSql(q, acc.FirstName, acc.LastName, strings.ToLower(acc.Username), acc.TempPwHash, acc.AccountStatus, D.TimeNow(), acc.AccountID)

	} else {
		// Do not update password hashes (regular updates)
		q := `update account set 
		first_name = ?,
		last_name = ?,
		username = ?,
		account_status = ?,
		time_modified = ?
		where account_id = ?;`
		_, err = D.ExecSql(q, acc.FirstName, acc.LastName, strings.ToLower(acc.Username), acc.AccountStatus, D.TimeNow(), acc.AccountID)
	}
	if err != nil {
		fmt.Println("Error Updating Account:", err)
		return nil, err
	}
	return acc, nil
}

func (D *Database) AccountsSelectAll() ([]Account, error) {
	q := accountSelect + ";"
	rows, err := D.QuerySql(q)
	if err != nil {
		fmt.Println("Error Selecting all Accounts:", err)
		return nil, err
	}
	return D.parseAccountRows(rows)
}

func (D *Database) AccountFromID(accountId int32) (*Account, error) {
	if accountId == -1 {
		return DefaultAdminAccount(), nil
	}
	q := accountSelect + " where account_id = ?;"
	rows, err := D.QuerySql(q, accountId)
	if err != nil {
		fmt.Println("Error Selecting Account from ID:", err)
		return nil, err
	}
	list, err := D.parseAccountRows(rows)
	if len(list) >= 1 {
		return &list[0], err
	}
	return nil, err
}

func (D *Database) AccountFromUser(username string) (*Account, error) {
	q := accountSelect + " where username = ?;"
	rows, err := D.QuerySql(q, strings.ToLower(username))
	if err != nil {
		fmt.Println("Error Selecting Account from username:", err)
		return nil, err
	}
	list, err := D.parseAccountRows(rows)
	if len(list) >= 1 {
		return &list[0], err
	}
	return nil, err
}

func (D *Database) AccountExists(username string) bool {
	q := "select account_id from account where username = ?;"
	rows, err := D.QuerySql(q, strings.ToLower(username))
	if err != nil {
		fmt.Println("Error Selecting Account from username:", err)
		return false
	}
	defer rows.Close()
	return rows.Next()
}

func (D *Database) AccountForUsernamePassword(u string, passw string) (*Account, error) {
	if blankdatabase && u == "admin" {
		return DefaultAdminAccount(), nil
	}
	q := fullaccountSelect + " where username = ?;"
	rows, err := D.QuerySql(q, strings.ToLower(u))
	if err != nil {
		fmt.Println("Error Selecting Account from username for pwcheck:", err)
		return nil, err
	}
	accounts, err2 := D.parseFullAccountRows(rows)
	if err2 != nil {
		return nil, err2
	}
	if len(accounts) != 1 {
		return nil, fmt.Errorf("Invalid Username/Password: %v", accounts)
	}
	//Verify the account is active
	if accounts[0].AccountStatus == Account_Inactive {
		return nil, fmt.Errorf("Invalid Account Status")
	}
	//Now validate the password hash
	err = accounts[0].validatePassword(passw)
	if err != nil {
		return nil, err
	}
	//Got a valid account - return it
	return &accounts[0], nil
}

func (D *Database) PruneAccounts(before time.Time) error {
	q := `DELETE from account where is_active = false and time_modified < ?;`
	_, err := D.ExecSql(q, D.ToTime(before))
	return err
}

func (A Account) validatePassword(pw string) error {
	//This handles the check for primary/temporary passwords associated with the account
	var err error = fmt.Errorf("Invalid Account")
	// Check primary password first (if one exists)
	if A.PwHash != "" {
		err = bcrypt.CompareHashAndPassword([]byte(A.PwHash), []byte(pw))
	}
	// Check temporary password next (if necessary and exists)
	if A.TempPwHash != "" && err != nil {
		err = bcrypt.CompareHashAndPassword([]byte(A.TempPwHash), []byte(pw))
	}
	return err
}
