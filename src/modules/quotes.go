package modules

import (
	"strings"

	"github.com/etu/flummbot/src/config"
	"github.com/etu/flummbot/src/db"
	"github.com/etu/flummbot/src/irc"
	ircevent "github.com/thoj/go-ircevent"
)

type Quotes struct{}

func (q Quotes) RegisterCallbacks(c *irc.IrcConnection) {
	if config.Get().Modules.Quotes.Enable {
		c.IrcEventConnection.AddCallback(
			"PRIVMSG",
			func(e *ircevent.Event) {
				go q.handle(c, e)
			},
		)
	}
}

func (q Quotes) handle(c *irc.IrcConnection, e *ircevent.Event) {
	cmd := strings.Split(e.Message(), " ")[0]

	if cmd == config.Get().Modules.Quotes.Command {
		msg := strings.Replace(e.Message(), cmd, "", 1)
		msg = strings.Trim(msg, " ")

		if len(msg) == 0 { // No message given: Fetch random quote
			var quote db.QuotesModel

			// Select random quote from the database
			db.Get().Gorm.Where(&db.QuotesModel{
				Network: c.Config.Name,
				Channel: e.Arguments[0],
			}).Order("RANDOM()").First(&quote)

			// If we have a real quote returned, print it, otherwise skip it.
			// Not having one is probably because of an empty database.
			if quote.ID != 0 {
				// Format the timestamp
				date := quote.CreatedAt.Format("2006-01-02 15:04:05")

				// Add zero-with-space between first and second letter to break
				// highlighting of the creator when printing a quote.
				spacedNick := quote.Nick[0:1] + "\u200B" + quote.Nick[1:]

				c.IrcEventConnection.Privmsg(
					e.Arguments[0],
					"Quote added by "+spacedNick+" @ "+date+": "+quote.Body,
				)
			}

		} else { // Add quote to database
			db.Get().Gorm.Create(&db.QuotesModel{
				Nick:    e.Nick,
				Body:    msg,
				Network: c.Config.Name,
				Channel: e.Arguments[0],
			})

			c.IrcEventConnection.Privmsg(
				e.Arguments[0],
				"Quote added, use "+cmd+" without params to get a random quote",
			)
		}
	}
}
