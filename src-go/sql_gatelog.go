package main

import (
	"database/sql"
	"encoding/base64"
	"time"
)

type GateLog struct {
	LogID       int64
	AccountID   int32
	OpenedName  string
	UsedCode    string
	UsedWeb     bool
	CodeTags    string
	GatePicture []byte
	TimeOpened  time.Time
	Success     bool
}

func (G GateLog) OpenedBy() string {
	if G.UsedWeb {
		return "Web"
	}
	return "PIN"
}

func (G GateLog) ShowPIN(accid int32) string {
	if accid == G.AccountID {
		return G.UsedCode
	}
	return "****"
}

func (G GateLog) HasImage() bool {
	return len(G.GatePicture) > 0
}

func (G GateLog) ImageBase64() string {
	if !G.HasImage() {
		return ""
	}
	return base64.StdEncoding.EncodeToString(G.GatePicture)
}

func (D *Database) CreateGateLogTable() error {
	q := `create table if not exists gatelog (
log_id integer primary key autoincrement,
account_id integer,
opened_name text,
used_code text,
used_web boolean,
code_tags text,
gate_picture_bytes blob,
time_opened integer not null,
success boolean
	);`
	_, err := D.ExecSql(q)
	return err
}

// internal function to read the rows from the table
func (D *Database) parseGatelogRows(rows *sql.Rows, with_picture bool) ([]GateLog, error) {
	defer rows.Close()
	var list []GateLog
	var t_opened int64
	for rows.Next() {
		var gl GateLog
		var err error
		if with_picture {
			if err = rows.Scan(&gl.LogID, &gl.AccountID, &gl.OpenedName, &gl.UsedCode, &gl.UsedWeb, &gl.CodeTags, &gl.GatePicture, &t_opened, &gl.Success); err != nil {
				return list, err
			}
		} else {
			if err = rows.Scan(&gl.LogID, &gl.AccountID, &gl.OpenedName, &gl.UsedCode, &gl.UsedWeb, &gl.CodeTags, &t_opened, &gl.Success); err != nil {
				return list, err
			}
		}
		gl.TimeOpened = D.ParseTime(t_opened)
		list = append(list, gl)
	}
	return list, nil
}

func (D *Database) GateLogInsert(gl *GateLog) (*GateLog, error) {
	q := `insert into gatelog (account_id, opened_name, used_code, used_web, code_tags, gate_picture_bytes, time_opened, success) values
		(?, ?, ?, ?, ?, ?, ?, ?)
		returning log_id;`
	rslt, err := D.ExecSql(q, gl.AccountID, gl.OpenedName, gl.UsedCode, gl.UsedWeb, gl.CodeTags, gl.GatePicture, D.TimeNow(), gl.Success)
	if err != nil {
		return nil, err
	}
	recordId, err := rslt.LastInsertId()
	gl.LogID = int64(recordId)
	return gl, err
}

func (D *Database) GatelogSelectAll() ([]GateLog, error) {
	q := `select log_id, account_id, opened_name, used_code, used_web, code_tags, time_opened, success
	from gatelog order by time_opened desc limit 1000;`
	rows, err := D.QuerySql(q)
	if err != nil {
		return nil, err
	}
	return D.parseGatelogRows(rows, false)
}

func (D *Database) GatelogSelectAccount(account int32) ([]GateLog, error) {
	q := `select log_id, account_id, opened_name, used_code, used_web, code_tags, time_opened, success
	from gatelog where account_id = ? order by time_opened desc limit 1000;`
	rows, err := D.QuerySql(q, account)
	if err != nil {
		return nil, err
	}
	return D.parseGatelogRows(rows, false)
}

func (D *Database) GateLogFromID(logId int64) (*GateLog, error) {
	q := `select log_id, account_id, opened_name, used_code, used_web, code_tags, gate_picture_bytes, time_opened, success
	from gatelog where log_id = ?;`
	rows, err := D.QuerySql(q, logId)
	if err != nil {
		return nil, err
	}
	list, err := D.parseGatelogRows(rows, true)
	if len(list) >= 1 {
		return &list[0], err
	}
	return nil, err
}

func (D *Database) PruneGateLogs(before time.Time) error {
	q := `DELETE from gatelog where time_opened < ?;`
	_, err := D.ExecSql(q, D.ToTime(before))
	return err
}
