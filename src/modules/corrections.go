package modules

import (
	"github.com/etu/flummbot/src/config"
	"github.com/etu/flummbot/src/db"
	"github.com/etu/flummbot/src/irc"
	"github.com/jinzhu/gorm"
	ircevent "github.com/thoj/go-ircevent"
	"strings"
)

type CorrectionsModel struct {
	gorm.Model
	Nick    string `gorm:"size:32"`
	Body    string `gorm:"size:512"`
	Network string `gorm:"size:64"`
	Channel string `gorm:"size:64"`
}

type Corrections struct {
	Config *config.ClientConfig
	Db     *db.Db
}

func (c Corrections) DbSetup() {
	c.Db.Gorm.AutoMigrate(&CorrectionsModel{})
}

func (c Corrections) RegisterCallbacks(conn *irc.IrcConnection) {
	conn.IrcEventConnection.AddCallback(
		"PRIVMSG",
		func(e *ircevent.Event) {
			go c.handle(conn, e)
		},
	)
}

func (c Corrections) handle(conn *irc.IrcConnection, e *ircevent.Event) {
	var correction CorrectionsModel

	msg := strings.Trim(e.Message(), " ")
	separator := c.Config.Modules.Corrections.Separator
	prefix := "s" + separator

	// Check so we don't go out of bounds and look for the prefix
	if len(msg) > 1 && msg[0:2] == prefix {
		c.Db.Gorm.Where(&CorrectionsModel{
			Nick:    e.Nick,
			Network: conn.Config.Name,
			Channel: e.Arguments[0],
		}).Last(&correction)

		// Split replacement message
		subs := strings.Split(e.Message(), separator)

		// Check so we don't go out of bounds
		if len(subs) > 2 {
			// Correct string
			corrected := strings.ReplaceAll(correction.Body, subs[1], subs[2])

			// Respond on IRC
			conn.IrcEventConnection.Privmsg(
				e.Arguments[0],
				"What "+e.Nick+" meant to say was: "+corrected,
			)
		}

	} else { // Record and clean messages for this user
		correction = CorrectionsModel{
			Nick:    e.Nick,
			Body:    msg,
			Network: conn.Config.Name,
			Channel: e.Arguments[0],
		}

		c.Db.Gorm.Create(&correction)

		for {
			var correctionToDelete CorrectionsModel

			// Select old items
			c.Db.Gorm.Where(&CorrectionsModel{
				Nick:    e.Nick,
				Network: conn.Config.Name,
				Channel: e.Arguments[0],
			}).First(&correctionToDelete)

			// Make sure to not clash with the newly added item
			if correction.ID == correctionToDelete.ID {
				break
			}

			// Do a real delete of this item
			c.Db.Gorm.Unscoped().Delete(correctionToDelete)
		}
	}
}
