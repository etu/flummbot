package main

import (
	"crypto/tls"
	"encoding/json"
	"flummbot"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"github.com/fluffle/goirc/logging/glog"
	"io/ioutil"
	"os"
)

func main() {
	glog.Init()

	var quit chan bool = make(chan bool)
	var tells flummbot.Tells
	var config flummbot.Config
	var quotes flummbot.Quotes
	var invite flummbot.Invite
	var helpers flummbot.Helpers = flummbot.Helpers{&config}

	// Read the configfile
	file, err := ioutil.ReadFile("./flummbot.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	// Parse config
	if err := json.Unmarshal(file, &config); err != nil {
		fmt.Printf("Config error: %v\n", err)
		os.Exit(1)
	}

	// Load up database
	db := helpers.SetupDatabase()
	defer db.Close()

	// Load up modules
	quotes = flummbot.Quotes{&config, db}
	tells = flummbot.Tells{&config, db}
	invite = flummbot.Invite{&config}

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
	invite.RegisterCallbacks(c)
	helpers.RegisterCallbacks(c, quit)

	// Connect
	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())

		os.Exit(1)
	}

	// Wait for disconnect
	<-quit
}
