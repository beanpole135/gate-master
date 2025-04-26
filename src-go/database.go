package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3" //SQLite database driver
)

const YearsRetentionPolicy = 1

type Database struct {
	filepath string
	db       *sql.DB
}

var blankdatabase bool = false

func NewDatabase(fpath string) (*Database, error) {
	// Make sure the containing folder exists first
	os.MkdirAll(filepath.Dir(fpath), 0700)
	// Niw open/create the database
	file, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	file.Close()
	D := Database{
		filepath: fpath,
	}
	D.db, err = sql.Open("sqlite3", fpath)
	if err != nil {
		return nil, err
	}
	if !D.TablesExist() {
		err = D.CreateTables()
	}
	return &D, err
}

func (D *Database) Close() {
	if D.db != nil {
		D.db.Close()
	}
}

func (D *Database) TablesExist() bool {
	list, err := D.AccountsSelectAll()
	if err == nil && len(list) > 0 {
		blankdatabase = false
		return true
	}
	return false
}

func (D *Database) CreateTables() error {
	fmt.Println("Creating tables")
	blankdatabase = true
	err := D.CreateAccTable()
	if err != nil {
		return err
	}
	err = D.CreateACTable()
	if err != nil {
		return err
	}
	err = D.CreateGateLogTable()
	if err != nil {
		return err
	}
	err = D.CreateContactTable()
	if err != nil {
		return err
	}
	return nil
}

func (D *Database) PruneTables() {
	//This is designed to be started as a background goroutine from main.go ONLY
	for range time.Tick(24 * time.Hour) {
		ya := time.Now().AddDate(-YearsRetentionPolicy, 0, 0) //years ago
		err := D.PruneGateLogs(ya)
		if err != nil {
			fmt.Println("Got error pruning GateLogs before %v: %v", ya, err)
		}
		err = D.PruneContacts(ya)
		if err != nil {
			fmt.Println("Got error pruning Contacts before %v: %v", ya, err)
		}
		err = D.PruneAccountCodes(ya)
		if err != nil {
			fmt.Println("Got error pruning AccountCodes before %v: %v", ya, err)
		}
		err = D.PruneAccounts(ya)
		if err != nil {
			fmt.Println("Got error pruning Accounts before %v: %v", ya, err)
		}
	}
}

func (D *Database) ExecSql(query string, args ...any) (sql.Result, error) {
	return D.db.Exec(query, args...)
}

func (D *Database) QuerySql(query string, args ...any) (*sql.Rows, error) {
	//Make sure you "defer rows.Close()" the rows returned!!
	return D.db.Query(query, args...)
}

// Time conversion functions
func (D *Database) TimeNow() int64 {
	return time.Now().Unix()
}

func (D *Database) ToTime(t time.Time) int64 {
	return t.Unix()
}

func (D *Database) ParseTime(secs int64) time.Time {
	return time.Unix(secs, 0)
}
