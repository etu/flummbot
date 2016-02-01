package flummbot

import (
	"database/sql"
	"github.com/fluffle/goirc/client"
	"strings"
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

func (t *Tells) RegisterCallbacks(c *client.Conn) *client.Conn {
	// fmt.Println(client.CONNECTED)

	c.HandleFunc(
		client.JOIN,
		func(conn *client.Conn, line *client.Line) {
			go t.deliverTells(conn, line)
		},
	)

	c.HandleFunc(
		client.PRIVMSG,
		func(conn *client.Conn, line *client.Line) {
			go t.deliverTells(conn, line)
		},
	)

	return c
}

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
// Wrapper to deliver !tell messages to the recipient. This will be used in //
// several locations.                                                       //
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
func (t *Tells) deliverTells(conn *client.Conn, line *client.Line) {
	// Check for messages to this person who joined
	rows, _ := t.Db.Query(
		"SELECT `id`, `from`, `body`, `date` "+
			"FROM tells WHERE `to` = ? AND `channel` = ?",
		line.Nick,
		line.Args[0],
	)

	// Make map with rows to delete from database
	toDelete := make(map[int]bool)

	for rows.Next() {
		var id int
		var from string
		var body string
		var date string

		// Fill vars with data
		_ = rows.Scan(&id, &from, &body, &date)

		// Remove the milliseconds from date
		date = strings.Split(date, ".")[0]

		// Print messages
		go conn.Privmsg(
			line.Args[0],
			line.Nick+": \""+body+"\" -- "+from+" @ "+date,
		)

		// Append to map to delete
		toDelete[id] = true
	}

	rows.Close()

	// Loop trough the map with ids to remove
	for id, _ := range toDelete {
		stmt, _ := t.Db.Prepare("DELETE FROM tells WHERE id = ?")
		stmt.Exec(id)

		stmt.Close()
	}
}
