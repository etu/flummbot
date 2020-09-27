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
		rows, err := c.Db.Gorm.Model(&CorrectionsModel{}).Where(&CorrectionsModel{
			Nick:    e.Nick,
			Network: conn.Config.Name,
			Channel: e.Arguments[0],
		}).Order("ID DESC").Rows()

		if err != nil {
			panic(err)
		}

		// Split replacement message
		subs := strings.Split(e.Message(), separator)

		if len(subs) > 2 {
			for rows.Next() {
				c.Db.Gorm.ScanRows(rows, &correction)

				if strings.Contains(correction.Body, subs[1]) {
					// Correct string
					corrected := strings.ReplaceAll(correction.Body, subs[1], subs[2])

					rows.Close()

					// Store in model
					correction = CorrectionsModel{
						Nick:    e.Nick,
						Body:    corrected,
						Network: conn.Config.Name,
						Channel: e.Arguments[0],
					}

					// Save the corrected one to the database as a new entry
					c.Db.Gorm.Create(&correction)

					// Respond on IRC
					conn.IrcEventConnection.Privmsg(
						e.Arguments[0],
						"What "+e.Nick+" meant to say was: "+corrected,
					)

					break
				}
			}
		}

	} else { // Record and clean messages for this user
		correction = CorrectionsModel{
			Nick:    e.Nick,
			Body:    msg,
			Network: conn.Config.Name,
			Channel: e.Arguments[0],
		}

		c.Db.Gorm.Create(&correction)

		var userCorrections []CorrectionsModel

		// Select old items
		rows, _ := c.Db.Gorm.Model(&CorrectionsModel{}).Where(&CorrectionsModel{
			Nick:    e.Nick,
			Network: conn.Config.Name,
			Channel: e.Arguments[0],
		}).Rows()

		// Aggregate all items in a slice
		for rows.Next() {
			c.Db.Gorm.ScanRows(rows, &correction)

			userCorrections = append(userCorrections, correction)
		}

		// Read user log size from config
		userLogSize := c.Config.Modules.Corrections.UserLogSize

		if len(userCorrections) > userLogSize {
			// Remove all but the three last items in the correction log
			for _, correction := range userCorrections[:len(userCorrections)-userLogSize] {
				c.Db.Gorm.Unscoped().Delete(correction)
			}
		}
	}
}
