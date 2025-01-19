package main

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3" //SQLite database driver
)

type Database struct {
	filepath string
	db *sql.DB
}

var blankdatabase bool = false

func NewDatabase(filepath string) (*Database, error) {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil { return nil, err }
	file.Close()
	D := Database{
		filepath: filepath,
	}
	D.db, err = sql.Open("sqlite3", filepath)
	if err != nil { return nil, err}
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
	return nil
}

func (D *Database) ExecSql(query string, args ...any) (sql.Result, error) {
	return D.db.Exec(query, args...)
}

func (D *Database) QuerySql(query string, args ...any) (*sql.Rows, error) {
	//Make sure you "defer rows.Close()" the rows returned!!
	return D.db.Query(query, args...)
}

//Time conversion functions
func (D *Database) TimeNow() int64 {
	return time.Now().Unix()
}

func (D *Database) ToTime(t time.Time) int64 {
	return t.Unix()
}

func (D *Database) ParseTime(secs int64) time.Time {
	return time.Unix(secs, 0)
}