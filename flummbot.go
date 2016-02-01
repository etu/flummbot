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
	var tells flummbot.Tells
	var config flummbot.Config
	var quotes flummbot.Quotes

	// Load up config
	if err := gcfg.ReadFileInto(&config, "flummbot.gcfg"); err != nil {
		fmt.Printf("Config error: %s\n", err)
		os.Exit(1)
	}

	// Load up helpers
	helpers := flummbot.Helpers{config}

	// Load up database
	db := helpers.SetupDatabase()
	defer db.Close()

	// Load up modules
	quotes = flummbot.Quotes{config, db}
	tells = flummbot.Tells{config, db}

	// Init databases
	tells.DbSetup()
	quotes.DbSetup()

	// Init irc-config
	cfg := irc.NewConfig(config.Connection.Nick)
	cfg.SSL = false
	cfg.Server = config.Connection.Server

	// Init irc-client
	c := irc.Client(cfg)

	// Register callbacks
	c = tells.RegisterCallbacks(c)

	// Add handlers to do things
	c.HandleFunc(irc.CONNECTED,
		connectCallback(
			config.Connection.Channel,
			config.Connection.NickservIdentify,
		),
	)
	c.HandleFunc(irc.DISCONNECTED, disconnectCallback(quit))
	c.HandleFunc(irc.PRIVMSG, privmsgCallback(db))

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
