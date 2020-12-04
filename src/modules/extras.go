package modules

import (
	"fmt"
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

				if len(words) > 1 && words[0] == config.Get().Modules.Extras.CountdownCommand {
					go func(conn *irc.IrcConnection, channel string, text string) {
						format := irc.GetFormat()

						for i := 10; i > 0; i-- {
							conn.IrcEventConnection.Privmsgf(
								channel,
								format.Bold+config.Get().Modules.Extras.CountdownMessageN,
								fmt.Sprintf("%s%d", format.Reset, i),
							)

							time.Sleep(time.Second)
						}

						conn.IrcEventConnection.Privmsgf(
							channel,
							format.Bold+config.Get().Modules.Extras.CountdownMessage0,
							fmt.Sprintf("%s%d", format.Reset, 0),
							format.Bold+text+format.Reset,
						)
					}(conn, e.Arguments[0], words[1])
				}
			},
		)
	}
}
