package main

import (
	"flummbot"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"gopkg.in/gcfg.v1"
	"os"
)

func main() {
	var quit chan bool = make(chan bool)
	var tells flummbot.Tells
	var config flummbot.Config
	var quotes flummbot.Quotes
	var helpers flummbot.Helpers

	// Load up config
	if err := gcfg.ReadFileInto(&config, "flummbot.gcfg"); err != nil {
		fmt.Printf("Config error: %s\n", err)
		os.Exit(1)
	}

	// Load up helpers
	helpers = flummbot.Helpers{&config}

	// Load up database
	db := helpers.SetupDatabase()
	defer db.Close()

	// Load up modules
	quotes = flummbot.Quotes{&config, db}
	tells = flummbot.Tells{&config, db}

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
	c = helpers.RegisterCallbacks(c, quit)

	// Connect
	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())

		os.Exit(1)
	}

	// Wait for disconnect
	<-quit
}
