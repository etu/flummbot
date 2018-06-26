package irc

import (
	"../db"
	"crypto/tls"
	"fmt"
	ircevent "github.com/thoj/go-ircevent"
	"log"
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
	ircEvent.AddCallback("332", ic.callbackHandler)
	ircEvent.AddCallback("353", ic.callbackHandler)
	ircEvent.AddCallback("CTCP_ACTION", ic.callbackHandler)
	ircEvent.AddCallback("JOIN", ic.callbackHandler)
	ircEvent.AddCallback("MODE", ic.callbackHandler)
	ircEvent.AddCallback("NICK", ic.callbackHandler)
	ircEvent.AddCallback("NOTICE", ic.callbackHandler)
	ircEvent.AddCallback("PART", ic.callbackHandler)
	ircEvent.AddCallback("PRIVMSG", ic.callbackHandler)
	ircEvent.AddCallback("TOPIC", ic.callbackHandler)

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

func (i *IrcConnection) callbackHandler(e *ircevent.Event) {
	switch e.Code {
	case "PRIVMSG": // Event for regular messages
		fmt.Println("got privmsg")

	case "332": // Event for topics you get on join
	case "TOPIC": // Event for topics on change
		fmt.Println("got topic")

	case "353": // Event for nick-list on join
		fmt.Println("got nick list on join")

	case "JOIN": // Event for channel joins
		fmt.Println("someone joined the channel")

	case "PART": // Event for channel leaves
		fmt.Println("someone left the channel")

	case "MODE": // Event for mode changes
		fmt.Println("got mode change")

	case "NICK": // Event for nick-changes
		fmt.Println("got nick change")

	case "NOTICE": // Event for notices
		fmt.Println("got notice")

	case "CTCP_ACTION": // Event for /me actions
		fmt.Println("got ctcp_action")
	}

	fmt.Println(e)
}

func (i *IrcConnection) onWelcome(e *ircevent.Event) {
	// Identify to nickserv before joining channels
	if len(i.Config.NickservIdentify) > 0 {
		i.IrcEventConnection.Privmsg("Nickserv", i.Config.NickservIdentify)
	}

	for _, channel := range i.Config.Channels {
		i.IrcEventConnection.Join(channel)
	}
}
