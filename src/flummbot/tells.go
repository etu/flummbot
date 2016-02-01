package flummbot

import (
	"database/sql"
)

type Tells struct {
	Config Config
	Db     *sql.DB
}

func (t *Tells) DbSetup() {
	t.Db.Exec(`
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
