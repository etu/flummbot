package main

import (
	"fmt"
	"github.com/fluffle/goirc/client"
)

type Invite struct {
	Config *Config
}

func (i *Invite) RegisterCallbacks(c *client.Conn) {
	c.HandleFunc(
		client.INVITE,
		func(conn *client.Conn, line *client.Line) {
			channel := line.Args[1]

			for _, nick := range i.Config.Invite.Whitelist {
				if nick == line.Nick {
					conn.Join(channel)

					conn.Privmsg(
						channel,
						fmt.Sprintf(i.Config.Invite.Message, line.Nick),
					)

					break
				}
			}
		},
	)
}
