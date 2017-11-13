package main

import (
	"database/sql"
	"github.com/fluffle/goirc/client"
	"sort"
	"strings"
)

type Tells struct {
	Config *Config
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

func (t *Tells) RegisterCallbacks(c *client.Conn) {
	c.HandleFunc(
		client.JOIN,
		func(conn *client.Conn, line *client.Line) {
			go t.deliver(conn, line)
		},
	)

	c.HandleFunc(
		client.PRIVMSG,
		func(conn *client.Conn, line *client.Line) {
			go t.deliver(conn, line)
			go t.register(conn, line)
		},
	)
}

func (t *Tells) register(conn *client.Conn, line *client.Line) {
	cmd := strings.Split(line.Args[1], " ")[0]

	// Look up if this module is allowed in the channel where it's used
	channel := line.Args[0]
	channels := t.Config.Tells.AllowedChannels

	sort.Strings(channels)
	i := sort.SearchStrings(channels, channel)

	// If it's not, complain in channel and just return
	if i < len(channels) && channels[i] != channel {
		conn.Privmsg(channel, "Module not enabled for this channel.")
		return
	}

	if cmd == t.Config.Tells.Command {
		target := strings.Split(line.Args[1], " ")[1]
		msg := strings.Replace(
			line.Args[1],
			t.Config.Tells.Command+" "+target+" ",
			"",
			1,
		)

		// Prepare query
		stmt, _ := t.Db.Prepare(
			`INSERT INTO tells("from", "to", "body", "date", "channel")
				VALUES(?, ?, ?, ?, ?)`,
		)
		defer stmt.Close()

		// Exec query: nick, target, msg, time, channel
		stmt.Exec(
			line.Nick,
			strings.ToLower(target),
			msg,
			line.Time,
			line.Args[0],
		)

		// Respond in channel
		conn.Privmsg(
			line.Args[0],
			"Alright, I'm going to tell "+target+": "+msg,
		)
	}
}

func (t *Tells) deliver(conn *client.Conn, line *client.Line) {
	// Check for messages to this person who joined
	rows, _ := t.Db.Query(
		"SELECT `id`, `from`, `body`, `date` "+
			"FROM tells WHERE `to` = ? AND `channel` = ?",
		strings.ToLower(line.Nick),
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
