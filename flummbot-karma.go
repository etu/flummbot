package main

import (
	"database/sql"
	"github.com/fluffle/goirc/client"
	"sort"
	"strconv"
	"strings"
)

type Karma struct {
	Config *Config
	Db     *sql.DB
}

func (k *Karma) DbSetup() {
	k.Db.Exec(`
		CREATE TABLE IF NOT EXISTS karma (
			"id" integer NOT NULL PRIMARY KEY,
			"item" text NOT NULL UNIQUE,
			"points" integer NOT NULL
		);
	`)
}

func (k *Karma) RegisterCallbacks(c *client.Conn) {
	c.HandleFunc(
		client.PRIVMSG,
		func(conn *client.Conn, line *client.Line) {
			go k.handle(conn, line)
		},
	)
}

func (k *Karma) handle(conn *client.Conn, line *client.Line) {
	plusOperator := k.Config.Karma.PlusOperator
	minusOperator := k.Config.Karma.MinusOperator

	words := strings.Split(line.Args[1], " ")
	cmd := words[0]

	// Look up if this module is allowed in the channel where it's used
	channel := line.Args[0]
	channels := k.Config.Karma.AllowedChannels

	sort.Strings(channels)
	i := sort.SearchStrings(channels, channel)

	// If it's not break the function
	if i < len(channels) && channels[i] != channel {
		// If used the command, complain in channel and just return
		if cmd == k.Config.Karma.Command {
			conn.Privmsg(channel, "Module not enabled for this channel.")
		}

		return
	}

	if cmd == k.Config.Karma.Command && len(words) > 1 {
		if words[1] == "" {
			return
		}

		stmt, _ := k.Db.Query(
			"SELECT points FROM karma WHERE `item` = ?",
			words[1],
		)
		defer stmt.Close()

		stmt.Next()

		var qPoints int

		// Read out items from database
		stmt.Scan(&qPoints)

		conn.Privmsg(channel, words[1]+" got the current karma "+strconv.Itoa(qPoints))
		return
	}

	// Set up map to store diffs in
	wordDiffs := make(map[string]int, len(words))

	for _, word := range words {
		if len(plusOperator) > 1 && strings.HasSuffix(word, plusOperator) {
			word = strings.Replace(word, plusOperator, "", 1)

			if word == "" {
				continue
			}

			wordDiffs[word] += 1
		}

		if len(minusOperator) > 1 && strings.HasSuffix(word, minusOperator) {
			word = strings.Replace(word, minusOperator, "", 1)

			if word == "" {
				continue
			}

			wordDiffs[word] -= 1
		}
	}

	// Make a list to store my messages in
	karmaReportMessage := make([]string, 0)

	for word, points := range wordDiffs {
		stmt, _ := k.Db.Query(
			"SELECT item, points FROM karma WHERE `item` = ?",
			word,
		)
		defer stmt.Close()

		stmt.Next()

		var qItem string
		var qPoints int

		// Read out items from database
		stmt.Scan(&qItem, &qPoints)
		stmt.Close()

		// Calculate new total points
		totalPoints := qPoints + points

		if qItem == word {
			// Prepare update statement
			dstmt, _ := k.Db.Prepare(`
				DELETE FROM karma WHERE item = ?
			`)

			defer dstmt.Close()

			// Exec query to save points
			dstmt.Exec(word)
			dstmt.Close()
		}
		// Prepare insert statement
		istmt, _ := k.Db.Prepare(`
			INSERT INTO karma("item", "points") VALUES(?, ?)
		`)

		defer istmt.Close()

		// Exec query to save points
		istmt.Exec(word, totalPoints)

		// Append messages to list of messages
		karmaReportMessage = append(
			karmaReportMessage,
			word+" karma changed to "+strconv.Itoa(totalPoints),
		)
	}

	if len(karmaReportMessage) > 0 {
		conn.Privmsg(
			line.Args[0],
			strings.Join(karmaReportMessage, ", ")+"!",
		)
	}
}
