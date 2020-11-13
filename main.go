package main // import "github.com/etu/flummbot"

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/etu/flummbot/src/config"
	"github.com/etu/flummbot/src/db"
	"github.com/etu/flummbot/src/irc"
	"github.com/etu/flummbot/src/modules"
)

func main() {
	var configFile string
	var debug bool

	connections := make(map[string]irc.IrcConnection)
	chQuitted := make(chan string)

	//
	// Parse command line flags
	//
	flag.BoolVar(&debug, "debug", false, "Enable or disable debug output")
	flag.StringVar(&configFile, "config", "flummbot.toml", "Specify path to the config file")
	flag.Parse()

	//
	// Parse config file
	//
	parsedConfig := config.New(configFile)

	//
	// Set up database
	//
	conn := db.New(&parsedConfig)
	defer conn.Gorm.Close()

	//
	// Listen to SIGUSR1 to reload the config
	//
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGUSR1)
	go func() {
		for {
			<-sigc
			log.Println("Recieved SIGUSR1, reloading config file")
			config.New(configFile)
		}
	}()

	//
	// Set up modules
	//
	modules := [...]modules.Module{
		modules.Corrections{},
		modules.Extras{},
		modules.Karma{},
		modules.Quotes{},
		modules.Tells{},
	}

	//
	// Set up connections per network connection defined
	//
	for _, network := range parsedConfig.Connections {
		config := irc.Config{
			Name:             network.Name,
			Server:           network.Server,
			Port:             network.Port,
			Channels:         network.Channels,
			User:             network.User,
			Nick:             network.Nick,
			Password:         network.Password,
			UseTLS:           network.UseTLS,
			ClientVersion:    "flummbot %version%",
			NickservIdentify: network.NickservIdentify,
			Debug:            debug,
		}

		conn := irc.New(&config)

		// Register callbacks for modules
		for _, module := range modules {
			module.RegisterCallbacks(&conn)
		}

		// Run client
		go conn.Run(chQuitted)

		// Store connection in map to keep track of it
		connections[network.Name] = conn
	}

	//
	// While we have active connections
	//
	for len(connections) > 0 {
		select {
		case quitted := <-chQuitted:
			// Delete quitted connections from the list of connections
			delete(connections, quitted)
		}
	}

	// End program when we don't have any connections left
}
