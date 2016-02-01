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
	c = quotes.RegisterCallbacks(c)

	// Add handlers to do things
	c.HandleFunc(irc.CONNECTED,
		connectCallback(
			config.Connection.Channel,
			config.Connection.NickservIdentify,
		),
	)
	c.HandleFunc(irc.DISCONNECTED, disconnectCallback(quit))

	// Connect
	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())

		os.Exit(1)
	}

	// Wait for disconnect
	<-quit
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
