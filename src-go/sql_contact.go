package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Contact struct {
	ContactID int64
	AccountID int32
	Email string
	PhoneNum string
	CellType string
	IsPrimary bool
	IsActive bool
	IsUtility bool
	IsDelivery bool
	//Internal audit fields
	TimeCreated time.Time
	TimeModified time.Time
}

func (C Contact) ContactEmail() string {
	if C.Email != "" {
		return C.Email
	}
	eml, _ := phoneToEmail(C.PhoneNum, C.CellType)
	return eml
}

func (C Contact) Display() string {
	if C.Email != "" {
		return C.Email
	}
	return fmt.Sprintf("%s (%s)", C.PhoneNum, strings.ToUpper(C.CellType))
}

func (C Contact) StatusValue() string {
	if C.IsActive {
		return "active"
	}
	return "inactive"
}

func (C Contact) ContactType() string {
	if C.Email != "" {
		return "email"
	}
	return C.CellType
}

func (C Contact) ContactValue() string {
	if C.Email != "" {
		return C.Email
	}
	return C.PhoneNum
}

func (D *Database) CreateContactTable() error {
	q := `create table if not exists contact (
contact_id integer primary key autoincrement,
account_id integer not null,
email text,
phone_num text,
cell_type text,
is_primary boolean default false,
is_active boolean default false,
is_utility boolean default false,
is_delivery boolean default false,
time_created integer not null,
time_modified integer not null
	);`
	_, err := D.ExecSql(q)
	return err
}

const contactquery = `select contact_id, account_id, email, phone_num, cell_type, is_primary, is_active, is_utility, is_delivery, time_created, time_modified from contact`
// internal function to read the rows from the table
func (D *Database) parseContactRows(rows *sql.Rows, with_picture bool) ([]Contact, error) {
	defer rows.Close()
	var list []Contact
	var t_create, t_mod int64
	for rows.Next() {
		var c Contact
		var err error
		if err = rows.Scan(&c.ContactID, &c.AccountID, &c.Email, &c.PhoneNum, &c.CellType, &c.IsPrimary, &c.IsActive, &c.IsUtility, &c.IsDelivery, &t_create, &t_mod); err != nil {
			return list, err
		}
		c.TimeCreated = D.ParseTime(t_create)
		c.TimeModified = D.ParseTime(t_mod)
		list = append(list, c)
	}
	return list, nil
}

func (D *Database) ContactInsert(c *Contact) (*Contact, error) {
	q := `insert into contact (account_id, email, phone_num, cell_type, is_primary, is_active, is_utility, is_delivery, time_created, time_modified) values
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		returning contact_id;`
	rslt, err := D.ExecSql(q, c.AccountID, c.Email, c.PhoneNum, c.CellType, c.IsPrimary, c.IsActive, c.IsUtility, c.IsDelivery, D.TimeNow(), D.TimeNow())
	if err != nil {
		return nil, err
	}
	recordId, err := rslt.LastInsertId()
	c.ContactID = int64(recordId)
	return c, err
}

func (D *Database) ContactUpdate(c *Contact) (*Contact, error) {
	if c.ContactID < 1 {
		return nil, fmt.Errorf("Missing Contact ID for ContactUpdate")
	}
	c.TimeModified = time.Now()
	q := `update contact set
		email = ?,
		phone_num = ?,
		cell_type = ?,
		is_primary = ?,
		is_active = ?,
		is_utility = ?,
		is_delivery = ?,
		time_modified = ?
		where contact_id = ? and account_id = ?;`
	_, err := D.ExecSql(q, 
		c.Email,
		c.PhoneNum,
		c.CellType,
		c.IsPrimary,
		c.IsActive,
		c.IsUtility,
		c.IsDelivery,
		D.TimeNow(),
		c.ContactID,
		c.AccountID,
	)
	if err != nil {
		return nil, err
	}
	return c, err
}

func (D *Database) ContactSelectAll() ([]Contact, error) {
	q := contactquery + ` order by time_created desc limit 1000;`
	rows, err := D.QuerySql(q)
	if err != nil {
		return nil, err
	}
	return D.parseContactRows(rows, false)
}

func (D *Database) ContactsForAccount(accId int64) ([]Contact, error) {
	q := contactquery + ` where account_id = ?;`
	rows, err := D.QuerySql(q, accId)
	if err != nil {
		return nil, err
	}
	return D.parseContactRows(rows, true)
}

func (D *Database) ContactFromID(contactId int64) (*Contact, error) {
	q := contactquery + ` where contact_id = ?;`
	rows, err := D.QuerySql(q, contactId)
	if err != nil {
		return nil, err
	}
	list, err := D.parseContactRows(rows, true)
	if len(list) >= 1 {
		return &list[0], err
	}
	return nil, err
}

func (D *Database) PruneContacts(before time.Time) error {
	q := `DELETE from contact where is_active = false and time_modified < ?;`
	_, err := D.ExecSql(q, D.ToTime(before))
	return err
}