package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	irc "github.com/fluffle/goirc/client"
	"github.com/fluffle/goirc/logging/glog"
	"io/ioutil"
	"os"
)

func main() {
	configFilePtr := flag.String("config", "flummbot.toml", "Path to config file")
	flag.Parse()

	var config Config

	// Read the configfile
	file, err := ioutil.ReadFile(*configFilePtr)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	// Parse config
	if _, err := toml.Decode(string(file), &config); err != nil {
		fmt.Printf("Config error: %v\n", err)
		os.Exit(1)
	}

	runBot(config)
}

func runBot(config Config) {
	glog.Init()

	var quit chan bool = make(chan bool)
	var tells Tells
	var quotes Quotes
	var invite Invite
	var helpers Helpers = Helpers{&config}

	// Load up database
	db := helpers.SetupDatabase()
	defer db.Close()

	// Load up modules
	quotes = Quotes{&config, db}
	tells = Tells{&config, db}
	invite = Invite{&config}

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
