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
	format := irc.GetFormat()
	parts := strings.SplitN(e.Message(), " ", 2)

	if parts[0] == config.Get().Modules.Quotes.Command {
		if len(parts) < 2 { // No message given: Fetch random quote
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

				c.IrcEventConnection.Privmsgf(
					e.Arguments[0],
					config.Get().Modules.Quotes.PrintMessage,
					format.Bold+spacedNick+format.Reset,
					format.Bold+date+format.Reset,
					format.Italics+quote.Body,
				)
			}

		} else { // Add quote to database
			msg := strings.Trim(parts[1], " ")

			db.Get().Gorm.Create(&db.QuotesModel{
				Nick:    e.Nick,
				Body:    msg,
				Network: c.Config.Name,
				Channel: e.Arguments[0],
			})

			c.IrcEventConnection.Privmsgf(
				e.Arguments[0],
				config.Get().Modules.Quotes.AddMessage,
				format.Bold+parts[0]+format.Reset,
			)
		}
	}
}
