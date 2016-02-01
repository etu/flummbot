package main

import (
	"database/sql"
	"flummbot"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/gcfg.v1"
	"os"
	"strings"
	"time"
)

func main() {
	var quit chan bool = make(chan bool)
	var config flummbot.Config
	var tells flummbot.Tells
	var quotes flummbot.Quotes

	// Load up config
	if err := gcfg.ReadFileInto(&config, "flummbot.gcfg"); err != nil {
		fmt.Printf("Config error: %s\n", err)
		os.Exit(1)
	}

	// Load up helpers/modules
	helpers := flummbot.Helpers{config}
	quotes = flummbot.Quotes{config}
	tells = flummbot.Tells{config}

	// Load up database
	db := helpers.SetupDatabase()
	defer db.Close()

	// Init databases
	tells.DbSetup(db)
	quotes.DbSetup(db)

	// Init irc-config
	cfg := irc.NewConfig(config.Connection.Nick)
	cfg.SSL = false
	cfg.Server = config.Connection.Server

	// Init irc-client
	c := irc.Client(cfg)

	// Add handlers to do things
	c.HandleFunc(irc.CONNECTED,
		connectCallback(
			config.Connection.Channel,
			config.Connection.NickservIdentify,
		),
	)
	c.HandleFunc(irc.DISCONNECTED, disconnectCallback(quit))
	c.HandleFunc(irc.PRIVMSG, privmsgCallback(db))
	c.HandleFunc(irc.JOIN, joinCallback(db))

	// Connect
	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())

		os.Exit(1)
	}

	// Wait for disconnect
	<-quit
}

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
// Callback to handle Privmsgs //
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
func privmsgCallback(db *sql.DB) func(*irc.Conn, *irc.Line) {
	return func(conn *irc.Conn, line *irc.Line) {
		cmd := strings.Split(line.Args[1], " ")[0]

		// Launch wrapper to deliver tells
		go deliverTells(db, conn, line)

		switch {
		case cmd == "!tell":
			target := strings.Split(line.Args[1], " ")[1]
			msg := strings.Replace(line.Args[1], "!tell "+target+" ", "", 1)

			// Prepare query
			stmt, _ := db.Prepare(`
				INSERT INTO tells("from", "to", "body", "date", "channel") VALUES(?, ?, ?, ?, ?)
			`)
			defer stmt.Close()

			// Exec query: nick, target, msg, time,      channel
			stmt.Exec(line.Nick, target, msg, line.Time, line.Args[0])

			// Respond in channel
			conn.Privmsg(line.Args[0], "Alright, I'm going to tell "+target+": "+msg)

		case cmd == "!quote":
			msg := strings.Replace(line.Args[1], "!quote", "", 1)
			msg = strings.Trim(msg, " ")

			if len(msg) == 0 { // No message given: Fetch random quote
				// Prepare query
				stmt, _ := db.Query("SELECT nick, quote, date FROM quotes ORDER BY RANDOM() LIMIT 1")
				defer stmt.Close()

				stmt.Next()

				var qNick string
				var qQuote string
				var qDate string

				stmt.Scan(&qNick, &qQuote, &qDate)

				// Remove the milliseconds from date
				qDate = strings.Split(qDate, ".")[0]

				// Return quote
				conn.Privmsg(line.Args[0], "Quote added by "+qNick+" @ "+qDate+": "+qQuote)

			} else { // Add quote to database
				// Prepare query
				stmt, _ := db.Prepare(`
					INSERT INTO quotes("nick", "quote", "date") VALUES(?, ?, ?)
				`)
				defer stmt.Close()

				// Exec query: nick, quote, date
				stmt.Exec(line.Nick, msg, line.Time)

				// Respond in channel
				conn.Privmsg(line.Args[0], "Quote added, use !quote without params to get a random quote")
			}
		}

		// fmt.Println("from", line.Nick)
		// fmt.Println("chan", line.Args[0])
		// fmt.Println("msg",  line.Args[1])
		// fmt.Println("time", line.Time)
	}
}

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
// Callback that launches the function to deliver tell messages to users. //
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
func joinCallback(db *sql.DB) func(*irc.Conn, *irc.Line) {
	return func(conn *irc.Conn, line *irc.Line) {
		// Launch wrapper to deliver tells
		go deliverTells(db, conn, line)
	}
}

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
// Callback with function to execute after connect. //
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
func connectCallback(channel string, nickserv string) func(*irc.Conn, *irc.Line) {
	return func(conn *irc.Conn, line *irc.Line) {
		// Identify to services
		if len(nickserv) > 0 {
			conn.Privmsg("NickServ", "IDENTIFY "+nickserv)
		}

		// Sleep while auth happens
		time.Sleep(time.Second)

		// Then join channel
		conn.Join(channel)

		// Greet everyone
		conn.Privmsg(channel, "Hejj")
	}
}

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
// Callback with function on disconnect event. This will end the program. //
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
func disconnectCallback(quit chan bool) func(*irc.Conn, *irc.Line) {
	return func(conn *irc.Conn, line *irc.Line) {
		quit <- true
	}
}

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
// Wrapper to deliver !tell messages to the recipient. This will be used in //
// several locations.                                                       //
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
func deliverTells(db *sql.DB, conn *irc.Conn, line *irc.Line) {
	// Check for messages to this person who joined
	rows, _ := db.Query(
		"SELECT `id`, `from`, `body`, `date` FROM tells WHERE `to` = ? AND `channel` = ?",
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
		go conn.Privmsg(line.Args[0], line.Nick+": \""+body+"\" -- "+from+" @ "+date)

		// Append to map to delete
		toDelete[id] = true
	}

	rows.Close()

	// Loop trough the map with ids to remove
	for id, _ := range toDelete {
		stmt, _ := db.Prepare("DELETE FROM tells WHERE id = ?")
		stmt.Exec(id)

		stmt.Close()
	}
}
