package flummbot

import (
	"database/sql"
)

type Tells struct {
	Config Config
}

func (t *Tells) DbSetup(db *sql.DB) {
	db.Exec(`
		CREATE TABLE IF NOT EXISTS tells (
			"id"      integer NOT NULL PRIMARY KEY,
			"from"    text,
			"to"      text,
			"channel" text,
			"body"    text,
			"date"    text
		);
	`)
}
