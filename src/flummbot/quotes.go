package flummbot

import (
	"database/sql"
)

type Quotes struct {
	Config Config
}

func (q *Quotes) DbSetup(db *sql.DB) {
	db.Exec(`
		CREATE TABLE IF NOT EXISTS quotes (
			"id"    integer NOT NULL PRIMARY KEY,
			"nick"  text,
			"quote" text,
			"date"  text
		);
	`)
}
