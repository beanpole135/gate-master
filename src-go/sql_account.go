package main

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	Account_Active = 1
	Account_Inactive = 2
)

type Account struct {
	AccountID int64
	FirstName string
	LastName string
	Username string
	PwHash string
	AccountStatus int
	TimeCreated time.Time
	TimeModified time.Time
}

func (D *Database) CreateAHTable() error {
	q := `create table if not exists account (
account_id integer primary key autoincrement,
first_name text not null,
last_name text not null,
username text not null unique,
pw_hash text not null,
account_status integer not null,
time_created integer not null,
time_modified integer
	);`
	_, err := D.ExecSql(q)
	return err
}

// internal function to read the rows from the account table
// NOTE: pw_hash is never returned!!
var accountSelect = `select account_id, first_name, last_name, username, account_status, time_created, time_modified
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

func (D *Database) AccountInsert(acc *Account) (*Account, error) {
	if acc.AccountStatus < 1 {
		acc.AccountStatus = Account_Active
	}

	q := `insert into account (first_name, last_name, username, pw_hash, account_status, time_created) values
		(?, ?, ?, ?, ?, ?)
		returning account_id;`
	rslt, err := D.ExecSql(q, acc.FirstName, acc.LastName, acc.Username, acc.PwHash, acc.AccountStatus, D.TimeNow())
	if err != nil {
		return nil, err
	}
	acc.AccountID, err = rslt.LastInsertId()
	return acc, err
}

func (D *Database) AccountUpdate(acc *Account) (*Account, error) {
	if acc.AccountID < 1 {
		return nil, fmt.Errorf("Missing Account ID for UpdateAccount")
	}
	acc.TimeModified = time.Now()
	q := `update account set 
	first_name = ?,
	last_name = ?,
	username = ?,
	pw_hash = ?,
	account_status = ?,
	time_modified = ?
	where account_id = ?;`
	_, err := D.ExecSql(q, acc.FirstName, acc.LastName, acc.Username, acc.PwHash, acc.AccountStatus, D.TimeNow(), acc.AccountID)
	if err != nil {
		return nil, err
	}
	return acc, err
}

func (D *Database) AccountsSelectAll() ([]Account, error) {
	q := accountSelect + ";"
	rows, err := D.QuerySql(q)
	if err != nil {
		return nil, err
	}
	return D.parseAccountRows(rows)
}

func (D *Database) AccountForUsernamePassword(u string, phash string) (*Account, error) {
	q := accountSelect + " where username = ? and pw_hash = ?;"
	rows, err := D.QuerySql(q, u, phash)
	if err != nil {
		return nil, err
	}
	accounts, err2 := D.parseAccountRows(rows)
	if err2 != nil {
		return nil, err2
	}
	if len(accounts) != 1 {
		return nil, fmt.Errorf("Invalid Username/Password")
	}
	//Got a valid account - return it
	return &accounts[0], nil
}