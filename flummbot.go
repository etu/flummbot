package main

import (
	"crypto/tls"
	"flummbot"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"github.com/fluffle/goirc/logging/glog"
	"gopkg.in/gcfg.v1"
	"os"
)

func main() {
	glog.Init()

	var quit chan bool = make(chan bool)
	var tells flummbot.Tells
	var config flummbot.Config
	var quotes flummbot.Quotes
	var helpers flummbot.Helpers = flummbot.Helpers{&config}

	// Load up config
	if err := gcfg.ReadFileInto(&config, "flummbot.gcfg"); err != nil {
		fmt.Printf("Config error: %s\n", err)
		os.Exit(1)
	}

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
	cfg.SSL = config.Connection.TLS
	cfg.Server = config.Connection.Server
	cfg.SSLConfig = &tls.Config{InsecureSkipVerify: true}

	// Init irc-client
	c := irc.Client(cfg)

	// Register callbacks
	tells.RegisterCallbacks(c)
	quotes.RegisterCallbacks(c)
	helpers.RegisterCallbacks(c, quit)

	// Connect
	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())

		os.Exit(1)
	}

	// Wait for disconnect
	<-quit
}
