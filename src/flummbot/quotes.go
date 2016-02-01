package flummbot

import (
	"database/sql"
)

type Quotes struct {
	Config Config
	Db     *sql.DB
}

func (q *Quotes) DbSetup() {
	q.Db.Exec(`
		CREATE TABLE IF NOT EXISTS quotes (
			"id"      integer NOT NULL PRIMARY KEY,
			"nick"    text,
			"quote"   text,
			"channel" text,
			"date"    text
		);
	`)
}
