package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type AccountCode struct {
	AccountCodeID int64
	AccountID int64
	Code string
	Label string
	IsActive bool
	IsUtility bool
	IsDelivery bool
	DateStart time.Time
	DateEnd time.Time
	TimeStart time.Time
	TimeEnd time.Time
	ValidDays []string //2-character abbreviations for days (su, tu, th)

	//Internal audit fields
	TimeCreated time.Time
	TimeModified time.Time
}

func (AC *AccountCode) IsValid() bool {
	//General active flag first
	if !AC.IsActive { return false }
	//Check valid dates (if either date is set - both optional)
	now := time.Now()
	if !AC.DateStart.IsZero() {
		if now.Before(AC.DateStart) {
			return false
		}
	}
	if !AC.DateEnd.IsZero() {
		if now.After(AC.DateEnd) {
			return false
		}
	}
	//Check valid day of the week (if set)
	if !nowValidWeekday(AC.ValidDays) {
		return false
	}

	//Now check valid time of day (if BOTH are set)
	if !nowBetweenTimes(AC.TimeStart, AC.TimeEnd) {
		return false
	}
	return true //All validity checks passed
}

func (D *Database) CreateACTable() error {
	q := `create table if not exists account_code (
account_code_id integer primary key autoincrement,
account_id not null,
code text not null unique,
label text not null,
is_active boolean default false,
is_utility boolean default false,
is_delivery boolean default false,
date_start integer,
date_end integer,
time_start integer,
time_end integer,
valid_days text,
time_created integer not null,
time_modified integer not null
	);`
	_, err := D.ExecSql(q)
	return err
}


// Quick internal functions for joining/splitting DB string
func combineVDays(days []string) string {
	//Ensure that strings in DB are all lowercase CSV
	return strings.ToLower(strings.Join(days, ","))
}

func splitVDays(days string) []string {
	return strings.Split(days, ",")
}


// internal function to read the rows from the account_code table
var accountCodeSelect = `select account_code_id, account_id, code, label, is_active, is_utility, is_delivery, date_start, date_end, time_start, time_end, valid_days, time_created, time_modified
	from account_code`

func (D *Database) parseAccountCodeRows(rows *sql.Rows) ([]AccountCode, error) {
	defer rows.Close()
	var accounts []AccountCode
	var t_created, t_mod, d_s, d_e, t_s, t_e int64
	var v_days string
	for rows.Next() {
		var acc AccountCode
		if err := rows.Scan(&acc.AccountCodeID, 
			&acc.AccountID,
			&acc.Code,
			&acc.Label,
			&acc.IsActive,
			&acc.IsUtility,
			&acc.IsDelivery,
			&d_s,
			&d_e,
			&t_s,
			&t_e,
			&v_days,
			&t_created,
			&t_mod); err != nil {
			return accounts, err
		}
		acc.TimeCreated = D.ParseTime(t_created)
		acc.TimeModified = D.ParseTime(t_mod)
		acc.DateStart = D.ParseTime(d_s)
		acc.DateEnd = D.ParseTime(d_e)
		acc.TimeStart = D.ParseTime(t_s)
		acc.TimeEnd = D.ParseTime(t_e)
		acc.ValidDays = splitVDays(v_days)
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (D *Database) AccountCodeInsert(acc *AccountCode) (*AccountCode, error) {
	q := `insert into account (
		account_id,
		code,
		label,
		is_active,
		is_utility,
		is_delivery,
		date_start,
		date_end,
		time_start,
		time_end,
		valid_days,
		time_created,
		time_modified) values
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		returning account_code_id;`
	rslt, err := D.ExecSql(q, 
		acc.AccountID,
		acc.Code,
		acc.Label,
		acc.IsActive,
		acc.IsUtility,
		acc.IsDelivery,
		D.ToTime(acc.DateStart),
		D.ToTime(acc.DateEnd),
		D.ToTime(acc.TimeStart),
		D.ToTime(acc.TimeEnd),
		combineVDays(acc.ValidDays),
		D.TimeNow(),
		D.TimeNow(),
	)
	if err != nil {
		return nil, err
	}
	acc.AccountCodeID, err = rslt.LastInsertId()
	return acc, err
}

func (D *Database) AccountCodeUpdate(acc *AccountCode) (*AccountCode, error) {
	if acc.AccountID < 1 {
		return nil, fmt.Errorf("Missing Account ID for AccountCodeUpdate")
	}
	acc.TimeModified = time.Now()
	q := `update account set
		account_id = ?,
		code = ?,
		label = ?,
		is_active = ?,
		is_utility = ?,
		is_delivery = ?,
		date_start = ?,
		date_end = ?,
		time_start = ?,
		time_end = ?,
		valid_days = ?,
		time_modified = ?
		where account_code_id = ?;`
	_, err := D.ExecSql(q, 
		acc.AccountID,
		acc.Code,
		acc.Label,
		acc.IsActive,
		acc.IsUtility,
		acc.IsDelivery,
		D.ToTime(acc.DateStart),
		D.ToTime(acc.DateEnd),
		D.ToTime(acc.TimeStart),
		D.ToTime(acc.TimeEnd),
		combineVDays(acc.ValidDays),
		D.TimeNow(),
		acc.AccountCodeID,
	)
	if err != nil {
		return nil, err
	}
	return acc, err
}

func (D *Database) AccountCodeSelectAll() ([]AccountCode, error) {
	q := accountCodeSelect + ";"
	rows, err := D.QuerySql(q)
	if err != nil {
		return nil, err
	}
	return D.parseAccountCodeRows(rows)
}

func (D *Database) AccountCodeMatch(code string) (*AccountCode, error) {
	q := accountCodeSelect + " where code = ?;"
	rows, err := D.QuerySql(q, code)
	if err != nil {
		return nil, err
	}
	accounts, err2 := D.parseAccountCodeRows(rows)
	if err2 != nil {
		return nil, err2
	}
	if len(accounts) != 1 {
		return nil, fmt.Errorf("Invalid Account Code")
	}
	//Got a valid account - return it
	return &accounts[0], nil
}