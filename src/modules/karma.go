package modules

import (
	"github.com/etu/flummbot/src/config"
	"github.com/etu/flummbot/src/db"
	"github.com/etu/flummbot/src/irc"
	"github.com/jinzhu/gorm"
	ircevent "github.com/thoj/go-ircevent"
	"strconv"
	"strings"
)

type KarmaModel struct {
	gorm.Model
	Item   string `gorm:"unique;not null"`
	Points int
}

type Karma struct {
	Config *config.ClientConfig
	Db     *db.Db
}

func (k *Karma) DbSetup() {
	k.Db.Gorm.AutoMigrate(&KarmaModel{})
}

func (k *Karma) RegisterCallbacks(c *irc.IrcConnection) {
	c.IrcEventConnection.AddCallback(
		"PRIVMSG",
		func(e *ircevent.Event) {
			go k.handle(c, e)
		},
	)
}

func (k *Karma) handle(c *irc.IrcConnection, e *ircevent.Event) {
	plusOperator := k.Config.Modules.Karma.PlusOperator
	minusOperator := k.Config.Modules.Karma.MinusOperator

	words := strings.Split(e.Message(), " ")
	cmd := words[0]

	if cmd == k.Config.Modules.Karma.Command && len(words) > 1 {
		if words[1] == "" {
			return
		}

		var karma KarmaModel
		k.Db.Gorm.Where(&KarmaModel{Item: words[1]}).First(&karma)

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
		var karma KarmaModel
		k.Db.Gorm.Where(&KarmaModel{Item: word}).First(&karma)

		// Calculate new total points
		karma.Points = karma.Points + points

		// Determine if we need to insert or update the row
		if karma.ID == 0 {
			// Insert item
			k.Db.Gorm.Create(&KarmaModel{
				Item:   word,
				Points: karma.Points,
			})
		} else {
			k.Db.Gorm.Model(&karma).UpdateColumns(karma)
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
