package modules

import (
	"github.com/etu/flummbot/src/config"
	"github.com/etu/flummbot/src/db"
	"github.com/etu/flummbot/src/irc"
	ircevent "github.com/thoj/go-ircevent"
	"strconv"
	"strings"
)

type Karma struct{}

func (k Karma) RegisterCallbacks(c *irc.IrcConnection) {
	c.IrcEventConnection.AddCallback(
		"PRIVMSG",
		func(e *ircevent.Event) {
			go k.handle(c, e)
		},
	)
}

func (k Karma) handle(c *irc.IrcConnection, e *ircevent.Event) {
	plusOperator := config.Get().Modules.Karma.PlusOperator
	minusOperator := config.Get().Modules.Karma.MinusOperator

	words := strings.Split(e.Message(), " ")
	cmd := words[0]

	if cmd == config.Get().Modules.Karma.Command && len(words) > 1 {
		if words[1] == "" {
			return
		}

		var karma db.KarmaModel
		db.Get().Gorm.Where(&db.KarmaModel{Item: words[1]}).First(&karma)

		c.IrcEventConnection.Privmsg(
			e.Arguments[0],
			karma.Item+" got the current karma "+strconv.Itoa(karma.Points)+"!",
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

			wordDiffs[word] += 1
		}

		if len(minusOperator) > 1 && strings.HasSuffix(word, minusOperator) {
			word = strings.Replace(word, minusOperator, "", 1)

			if word == "" {
				continue
			}

			wordDiffs[word] -= 1
		}
	}

	// Make a list to store my messages in
	karmaReportMessage := make([]string, 0)

	for word, points := range wordDiffs {
		var karma db.KarmaModel
		db.Get().Gorm.Where(&db.KarmaModel{Item: word}).First(&karma)

		// Calculate new total points
		karma.Points = karma.Points + points

		// Determine if we need to insert or update the row
		if karma.ID == 0 {
			// Insert item
			db.Get().Gorm.Create(&db.KarmaModel{
				Item:   word,
				Points: karma.Points,
			})
		} else {
			db.Get().Gorm.Model(&karma).UpdateColumns(karma)
		}

		// Append messages to list of messages
		karmaReportMessage = append(
			karmaReportMessage,
			word+" karma changed to "+strconv.Itoa(karma.Points),
		)
	}

	if len(karmaReportMessage) > 0 {
		c.IrcEventConnection.Privmsg(
			e.Arguments[0],
			strings.Join(karmaReportMessage, ", ")+"!",
		)
	}
}
