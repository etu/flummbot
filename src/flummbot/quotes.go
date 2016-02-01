package flummbot

import (
	"database/sql"
	"github.com/fluffle/goirc/client"
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

func (t *Quotes) RegisterCallbacks(c *client.Conn) *client.Conn {
	return c
}
