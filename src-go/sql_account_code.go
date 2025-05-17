package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func validatePinCodeFormat(p string) bool {
	for _, v := range p {
		//This checks each character for a numeric value (0-9)
		if v < '0' || v > '9' {
			return false
		}
	}
	return len(p) >= 4
}

type AccountCode struct {
	AccountCodeID int64
	AccountID     int32
	Code          string
	CodeLength    int //Not stored in database - temporary variable
	Label         string
	IsActive      bool
	IsUtility     bool
	IsDelivery    bool
	IsContractor  bool
	IsMail        bool
	DateStart     time.Time
	DateEnd       time.Time
	TimeStart     time.Time
	TimeEnd       time.Time
	ValidDays     []string //2-character abbreviations for days (su, tu, th)

	//Internal audit fields
	TimeCreated  time.Time
	TimeModified time.Time

	//Internal pass-through field (not stored in DB)
	AccountName string
}

func (A AccountCode) Status() string {
	if A.IsActive {
		return "Active"
	}
	return "Inactive"
}

func (A AccountCode) TagsString() string {
	var tags []string
	if A.IsUtility {
		tags = append(tags, "Utility")
	}
	if A.IsDelivery {
		tags = append(tags, "Delivery")
	}
	if A.IsContractor {
		tags = append(tags, "Contractor")
	}
	if A.IsMail {
		tags = append(tags, "Mail")
	}
	return strings.Join(tags, ", ")
}

func (A AccountCode) HasDay(d string) bool {
	if len(A.ValidDays) < 1 || len(A.ValidDays) > 6 {
		return true
	}
	for _, v := range A.ValidDays {
		if v == d {
			return true
		}
	}
	return false
}

func (AC AccountCode) WhenValidString() string {
	var lines []string
	datefmt := "Jan _2, 2006"
	timefmt := "3:04PM"
	//Valid dates
	if !AC.DateStart.IsZero() && !AC.DateEnd.IsZero() {
		lines = append(lines, AC.DateStart.Format(datefmt)+" to "+AC.DateEnd.Format(datefmt))
	} else if !AC.DateStart.IsZero() {
		lines = append(lines, "After "+AC.DateStart.Format(datefmt))
	} else if !AC.DateEnd.IsZero() {
		lines = append(lines, "Until "+AC.DateEnd.Format(datefmt))
	}
	//Valid times
	if !AC.TimeStart.IsZero() && !AC.TimeEnd.IsZero() {
		lines = append(lines, AC.TimeStart.Format(timefmt)+" to "+AC.TimeEnd.Format(timefmt))
	} else if !AC.TimeStart.IsZero() {
		lines = append(lines, "After "+AC.TimeStart.Format(timefmt))
	} else if !AC.TimeEnd.IsZero() {
		lines = append(lines, "Until "+AC.TimeEnd.Format(timefmt))
	}
	//Valid days of the week
	if len(AC.ValidDays) > 0 && len(AC.ValidDays) < 7 {
		lines = append(lines, combineVDays(AC.ValidDays))
	}
	if len(lines) == 0 {
		lines = append(lines, "Always Valid")
	}
	return strings.Join(lines, "\n")
}

func (AC *AccountCode) IsValid() bool {
	//General active flag first
	if !AC.IsActive {
		return false
	}
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
is_contractor boolean default false,
is_mail boolean default false,
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
var accountCodeSelect = `select account_code_id, account_id, code, label, is_active, is_utility, is_delivery, is_contractor, is_mail, date_start, date_end, time_start, time_end, valid_days, time_created, time_modified
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
			&acc.IsContractor,
			&acc.IsMail,
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
	q := `insert into account_code (
		account_id,
		code,
		label,
		is_active,
		is_utility,
		is_delivery,
		is_contractor,
		is_mail,
		date_start,
		date_end,
		time_start,
		time_end,
		valid_days,
		time_created,
		time_modified) values
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		returning account_code_id;`
	rslt, err := D.ExecSql(q,
		acc.AccountID,
		acc.Code,
		acc.Label,
		acc.IsActive,
		acc.IsUtility,
		acc.IsDelivery,
		acc.IsContractor,
		acc.IsMail,
		D.ToTime(acc.DateStart),
		D.ToTime(acc.DateEnd),
		D.ToTime(acc.TimeStart),
		D.ToTime(acc.TimeEnd),
		combineVDays(acc.ValidDays),
		D.TimeNow(),
		D.TimeNow(),
	)
	if err != nil {
		fmt.Println("Error Inserting AccountCode:", err)
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
	q := `update account_code set
		account_id = ?,
		code = ?,
		label = ?,
		is_active = ?,
		is_utility = ?,
		is_delivery = ?,
		is_contractor = ?,
		is_mail = ?,
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
		acc.IsContractor,
		acc.IsMail,
		D.ToTime(acc.DateStart),
		D.ToTime(acc.DateEnd),
		D.ToTime(acc.TimeStart),
		D.ToTime(acc.TimeEnd),
		combineVDays(acc.ValidDays),
		D.TimeNow(),
		acc.AccountCodeID,
	)
	if err != nil {
		fmt.Println("Error Updating AccountCode:", err)
		return nil, err
	}
	return acc, err
}

func (D *Database) AccountCodeSelectAll(accountid int32, accCodeID int64) ([]AccountCode, error) {
	//accountid = 0 means return everything
	q := accountCodeSelect
	var conditions []string
	var args []interface{}
	if accountid > 0 {
		conditions = append(conditions, "account_id = ?")
		args = append(args, accountid)
	}
	if accCodeID > 0 {
		conditions = append(conditions, "account_code_id = ?")
		args = append(args, accCodeID)
	}
	if len(conditions) > 0 {
		q += " where " + strings.Join(conditions, " and ")
	}
	rows, err := D.QuerySql(q+";", args...)
	if err != nil {
		fmt.Println("Error Selecting All AccountCode:", err)
		return nil, err
	}
	return D.parseAccountCodeRows(rows)
}

func (D *Database) AccountCodeMatch(code string) (*AccountCode, error) {
	q := accountCodeSelect + " where code = ?;"
	rows, err := D.QuerySql(q, code)
	if err != nil {
		fmt.Println("Error Selecting AccountCode from code:", err)
		return nil, err
	}
	accounts, err2 := D.parseAccountCodeRows(rows)
	if err2 != nil {
		return nil, err2
	}
	if len(accounts) != 1 {
		return nil, nil
	}
	//Got a valid account - return it
	return &accounts[0], nil
}

func (D *Database) PruneAccountCodes(before time.Time) error {
	q := `DELETE from account_code where is_active = false and time_modified < ?;`
	_, err := D.ExecSql(q, D.ToTime(before))
	if err != nil {
		fmt.Println("Error Deleting AccountCodes:", err)
	}
	return err
}
