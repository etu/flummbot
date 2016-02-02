package flummbot

import (
	"database/sql"
	"github.com/fluffle/goirc/client"
	"strings"
)

type Quotes struct {
	Config *Config
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

func (q *Quotes) RegisterCallbacks(c *client.Conn) *client.Conn {
	c.HandleFunc(
		client.PRIVMSG,
		func(conn *client.Conn, line *client.Line) {
			go q.handle(conn, line)
		},
	)

	return c
}

func (q *Quotes) handle(conn *client.Conn, line *client.Line) {
	cmd := strings.Split(line.Args[1], " ")[0]

	if cmd == q.Config.Quotes.Command {
		msg := strings.Replace(line.Args[1], q.Config.Quotes.Command, "", 1)
		msg = strings.Trim(msg, " ")

		if len(msg) == 0 { // No message given: Fetch random quote
			// Prepare query
			stmt, _ := q.Db.Query(
				"SELECT nick, quote, date FROM quotes "+
					"WHERE `channel` = ? "+
					"ORDER BY RANDOM() LIMIT 1",
				line.Args[0],
			)
			defer stmt.Close()

			stmt.Next()

			var qNick string
			var qQuote string
			var qDate string

			stmt.Scan(&qNick, &qQuote, &qDate)

			// Remove the milliseconds from date
			qDate = strings.Split(qDate, ".")[0]

			// Return quote
			conn.Privmsg(
				line.Args[0],
				"Quote added by "+qNick+" @ "+qDate+": "+qQuote,
			)

		} else { // Add quote to database
			// Prepare query
			stmt, _ := q.Db.Prepare(`
				INSERT INTO quotes("nick", "quote", "date", "channel")
				VALUES(?, ?, ?, ?)
			`)
			defer stmt.Close()

			// Exec query: nick, quote, date, channel
			stmt.Exec(line.Nick, msg, line.Time, line.Args[0])

			// Respond in channel
			conn.Privmsg(
				line.Args[0],
				"Quote added, use "+q.Config.Quotes.Command+
					" without params to get a random quote",
			)
		}
	}
}
