package modules

import (
	"fmt"
	"log"
	"strings"

	"github.com/etu/flummbot/src/config"
	"github.com/etu/flummbot/src/db"
	"github.com/etu/flummbot/src/irc"
	ircevent "github.com/thoj/go-ircevent"
)

type Karma struct{}

func (k Karma) RegisterCallbacks(c *irc.IrcConnection) {
	if config.Get().Modules.Karma.Enable {
		c.IrcEventConnection.AddCallback(
			"PRIVMSG",
			func(e *ircevent.Event) {
				go k.handle(c, e)
			},
		)
		c.IrcEventConnection.AddCallback(
			"CTCP_ACTION",
			func(e *ircevent.Event) {
				go k.handle(c, e)
			},
		)
	}
}

func (k Karma) handle(c *irc.IrcConnection, e *ircevent.Event) {
	format := irc.GetFormat()
	plusOperator := config.Get().Modules.Karma.PlusOperator
	minusOperator := config.Get().Modules.Karma.MinusOperator

	words := strings.Split(e.Message(), " ")
	cmd := words[0]

	if cmd == config.Get().Modules.Karma.Command && len(words) > 1 {
		if words[1] == "" {
			return
		}

		var karma db.KarmaModel

		// If this word has a static value, just ignore the database and set a value
		if val, ok := config.Get().Modules.Karma.StaticValues[words[1]]; ok {
			karma.Points = val
		} else {
			// Otherwise, look it up in the database.
			db.Get().Gorm.Where(&db.KarmaModel{Item: strings.ToLower(words[1])}).First(&karma)
		}

		c.IrcEventConnection.Privmsgf(
			e.Arguments[0],
			config.Get().Modules.Karma.ReportMessage,
			format.Bold+strings.ToLower(words[1])+format.Reset,
			fmt.Sprintf(
				"%s%d%s",
				format.Bold,
				karma.Points,
				format.Reset,
			),
		)

		return
	}

	// Set up map to store diffs in
	wordDiffs := make(map[string]int, len(words))

	for _, word := range words {
		if len(plusOperator) > 1 && strings.HasSuffix(word, plusOperator) {
			word = strings.Replace(word, plusOperator, "", 1)

			if word == "" {
				continue
			}

			wordDiffs[strings.ToLower(word)] += 1
		}

		if len(minusOperator) > 1 && strings.HasSuffix(word, minusOperator) {
			word = strings.Replace(word, minusOperator, "", 1)

			if word == "" {
				continue
			}

			wordDiffs[strings.ToLower(word)] -= 1
		}
	}

	// Make a list to store my messages in
	karmaReportMessage := make([]string, 0)

	for word, points := range wordDiffs {
		if _, ok := config.Get().Modules.Karma.StaticValues[word]; ok {
			log.Printf("Karma change for static word %s was ignored", word)
			continue
		}

		var karma db.KarmaModel
		db.Get().Gorm.Where(&db.KarmaModel{Item: strings.ToLower(word)}).First(&karma)

		// Calculate new total points
		karma.Points = karma.Points + points

		// Determine if we need to insert or update the row
		if karma.ID == 0 {
			// Insert item
			db.Get().Gorm.Create(&db.KarmaModel{
				Item:   strings.ToLower(word),
				Points: karma.Points,
			})
		} else {
			db.Get().Gorm.Model(&db.KarmaModel{}).Where("ID = ?", karma.ID).Update("points", karma.Points)
		}

		// Append messages to list of messages
		karmaReportMessage = append(
			karmaReportMessage,
			fmt.Sprintf(
				config.Get().Modules.Karma.ChangeMessage,
				format.Bold+word+format.Reset,
				fmt.Sprintf(
					"%s%d%s",
					format.Bold,
					karma.Points,
					format.Reset,
				),
			),
		)
	}

	if len(karmaReportMessage) > 0 {
		c.IrcEventConnection.Privmsg(
			e.Arguments[0],
			strings.Join(karmaReportMessage, ", ")+"!",
		)
	}
}
