package main // import "github.com/etu/flummbot"

import (
	"github.com/etu/flummbot/src/args"
	"github.com/etu/flummbot/src/config"
	"github.com/etu/flummbot/src/db"
	"github.com/etu/flummbot/src/irc"
	"github.com/etu/flummbot/src/modules"
)

func main() {
	connections := make(map[string]irc.IrcConnection)
	chQuitted := make(chan string)

	cmdArguments := args.Parse()

	// Parse config file
	config := config.New(cmdArguments.ConfigFile)

	// Set up database
	database := db.New(&config)
	defer database.Gorm.Close()

	// Set up modules
	moduleCorrections := modules.Corrections{&config, &database}
	moduleKarma := modules.Karma{&config, &database}
	moduleQuotes := modules.Quotes{&config, &database}
	moduleTells := modules.Tells{&config, &database}

	moduleCorrections.DbSetup()
	moduleKarma.DbSetup()
	moduleQuotes.DbSetup()
	moduleTells.DbSetup()

	// Set up connections per network connection defined
	for _, network := range config.Connections {
		config := irc.Config{
			Name:             network.Name,
			Server:           network.Server,
			Port:             network.Port,
			Channels:         network.Channels,
			User:             network.User,
			Nick:             network.Nick,
			Password:         network.Password,
			UseTLS:           network.UseTLS,
			ClientVersion:    "flummbot 2.0.0-alpha1",
			NickservIdentify: network.NickservIdentify,
			Debug:            cmdArguments.Debug,
		}

		conn := irc.New(&config, database)

		// Register callbacks for modules
		moduleCorrections.RegisterCallbacks(&conn)
		moduleKarma.RegisterCallbacks(&conn)
		moduleQuotes.RegisterCallbacks(&conn)
		moduleTells.RegisterCallbacks(&conn)

		// Run client
		go conn.Run(chQuitted)

		// Store connection in map to keep track of it
		connections[network.Name] = conn
	}

	// While we have active connections
	for len(connections) > 0 {
		select {
		case quitted := <-chQuitted:
			// Delete quitted connections from the list of connections
			delete(connections, quitted)
		}
	}

	// End program when we don't have any connections left
}
