package modules

import (
	"strings"
	"time"

	"github.com/etu/flummbot/src/config"
	"github.com/etu/flummbot/src/irc"
	ircevent "github.com/thoj/go-ircevent"
)

type Extras struct{}

func (_ Extras) RegisterCallbacks(conn *irc.IrcConnection) {
	if config.Get().Modules.Extras.Enable {
		conn.IrcEventConnection.AddCallback(
			"PRIVMSG",
			func(e *ircevent.Event) {
				words := strings.SplitN(e.Message(), " ", 2)

				if len(words) > 1 && words[0] == "!countdown" {
					go func(conn *irc.IrcConnection, channel string, text string) {
						format := irc.GetFormat()

						for i := 3; i > 0; i-- {
							conn.IrcEventConnection.Privmsgf(
								channel,
								"%sCountdown:%s %d",
								format.Bold,
								format.Reset,
								i,
							)

							time.Sleep(time.Second)
						}

						conn.IrcEventConnection.Privmsgf(
							channel,
							"%sCountdown:%s 0!!11!one! %s is happening!",
							format.Bold,
							format.Reset,
							format.Bold+format.Color+format.Colors.Magenta+text+format.Reset,
						)
					}(conn, e.Arguments[0], words[1])
				}
			},
		)
	}
}
