package irc

import (
	"../db"
	"crypto/tls"
	"fmt"
	ircevent "github.com/thoj/go-ircevent"
	"log"
	"time"
)

type Config struct {
	Name             string   // Name of network
	Server           string   // IRC server
	Port             uint16   // IRC port
	Channels         []string // Default channels to join
	User             string   // IRC username
	Nick             string   // IRC nickname
	Password         string   // Server password
	UseTLS           bool     // Connect with TLS or not
	ClientVersion    string   // Version string of client
	NickservIdentify string   // Message to send to Nickserv on connect
	Debug            bool     // Log debug info
}

type IrcConnection struct {
	IrcEventConnection *ircevent.Connection
	Config             *Config
	Db                 db.Db
}

func New(config *Config, db db.Db) IrcConnection {
	ircEvent := ircevent.IRC(config.Nick, config.User)
	ircEvent.Version = config.ClientVersion
	ircEvent.Password = config.Password
	ircEvent.UseTLS = config.UseTLS
	ircEvent.TLSConfig = &tls.Config{
		ServerName: config.Server,
	}
	ircEvent.VerboseCallbackHandler = config.Debug

	ic := IrcConnection{ircEvent, config, db}

	ircEvent.AddCallback("001", ic.onWelcome)

	return ic
}

func (i *IrcConnection) Run(chQuitted chan string) {
	serverString := fmt.Sprintf("%s:%d", i.Config.Server, i.Config.Port)

	err := i.IrcEventConnection.Connect(serverString)
	if err != nil {
		log.Fatal(err)
	}

	i.IrcEventConnection.Loop()

	chQuitted <- i.Config.Name
}

func (i *IrcConnection) onWelcome(e *ircevent.Event) {
	// Identify to nickserv before joining channels
	if len(i.Config.NickservIdentify) > 0 {
		log.Println(fmt.Sprintf("Sending: '%s' to nickserv", i.Config.NickservIdentify))
		i.IrcEventConnection.Privmsg("nickserv", i.Config.NickservIdentify)
		time.Sleep(time.Second)
	}

	for _, channel := range i.Config.Channels {
		i.IrcEventConnection.Join(channel)
	}
}
