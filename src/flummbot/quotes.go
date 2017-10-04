package flummbot

import (
	"database/sql"
	"github.com/fluffle/goirc/client"
	"sort"
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

func (q *Quotes) RegisterCallbacks(c *client.Conn) {
	c.HandleFunc(
		client.PRIVMSG,
		func(conn *client.Conn, line *client.Line) {
			go q.handle(conn, line)
		},
	)
}

func (q *Quotes) handle(conn *client.Conn, line *client.Line) {
	cmd := strings.Split(line.Args[1], " ")[0]

	// Look up if this module is allowed in the channel where it's used
	channel := line.Args[0]
	channels := q.Config.Quotes.AllowedChannels

	sort.Strings(channels)
	i := sort.SearchStrings(channels, channel)

	// If it's not, complain in channel and just return
	if i < len(channels) && channels[i] != channel {
		conn.Privmsg(channel, "Module not enabled for this channel.")
		return
	}

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
			var spacedNick string

			stmt.Scan(&qNick, &qQuote, &qDate)

			// Remove the milliseconds from date
			qDate = strings.Split(qDate, ".")[0]

			// Add zero-with-space between first and second letter to break
			// highlighting of the creator when printing a quote.
			spacedNick = qNick[0:1] + "\u200B" + qNick[1:]

			// Return quote
			conn.Privmsg(
				line.Args[0],
				"Quote added by "+spacedNick+" @ "+qDate+": "+qQuote,
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
