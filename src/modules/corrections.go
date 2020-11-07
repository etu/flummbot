package modules

import (
	"strings"

	"github.com/etu/flummbot/src/config"
	"github.com/etu/flummbot/src/db"
	"github.com/etu/flummbot/src/irc"
	ircevent "github.com/thoj/go-ircevent"
)

type Corrections struct{}

func (c Corrections) RegisterCallbacks(conn *irc.IrcConnection) {
	if config.Get().Modules.Corrections.Enable {
		conn.IrcEventConnection.AddCallback(
			"PRIVMSG",
			func(e *ircevent.Event) {
				go c.handle(conn, e)
			},
		)
		conn.IrcEventConnection.AddCallback(
			"CTCP_ACTION",
			func(e *ircevent.Event) {
				go c.handle(conn, e)
			},
		)
	}
}

func (c Corrections) handle(conn *irc.IrcConnection, e *ircevent.Event) {
	var correction db.CorrectionsModel
	prefixes := make(map[string]bool)
	format := irc.GetFormat()

	// Build a map of separator with the key as value for lookup of the separator.
	for _, value := range config.Get().Modules.Corrections.Separators {
		prefixes["s"+value] = true
	}

	msg := strings.Trim(e.Message(), " ")

	// Check so we don't go out of bounds and look for the prefix
	if len(msg) > 1 && prefixes[msg[0:2]] == true {
		rows, err := db.Get().Gorm.Model(&db.CorrectionsModel{}).Where(&db.CorrectionsModel{
			Nick:    e.Nick,
			Network: conn.Config.Name,
			Channel: e.Arguments[0],
		}).Order("ID DESC").Rows()

		if err != nil {
			panic(err)
		}

		// Split replacement message
		subs := strings.Split(e.Message(), msg[1:2])

		if len(subs) > 2 {
			for rows.Next() {
				db.Get().Gorm.ScanRows(rows, &correction)

				if strings.Contains(correction.Body, subs[1]) {
					// Correct string
					corrected := strings.ReplaceAll(correction.Body, subs[1], subs[2])

					rows.Close()

					// Store in model
					correction = db.CorrectionsModel{
						Nick:    e.Nick,
						Type:    correction.Type,
						Body:    corrected,
						Network: conn.Config.Name,
						Channel: e.Arguments[0],
					}

					// Save the corrected one to the database as a new entry
					db.Get().Gorm.Create(&correction)

					// Have a different prefix before message if it's an ACTION message
					prefix := ""
					if correction.Type == "CTCP_ACTION" {
						prefix = "* " + format.Bold + e.Nick + format.Reset + ": "
					}

					// Respond on IRC
					conn.IrcEventConnection.Privmsgf(
						e.Arguments[0],
						"What %s meant to say was: %s%s",
						format.Bold+e.Nick+format.Reset,
						prefix,
						format.Italics+corrected,
					)

					break
				}
			}
		}

	} else { // Record messages for this user
		correction = db.CorrectionsModel{
			Nick:    e.Nick,
			Type:    e.Code,
			Body:    msg,
			Network: conn.Config.Name,
			Channel: e.Arguments[0],
		}

		db.Get().Gorm.Create(&correction)
	}

	// When everything is done, let's go through and clean up so we don't store too
	// many messages for the user
	var userCorrections []db.CorrectionsModel

	// Select all items
	rows, _ := db.Get().Gorm.Model(&db.CorrectionsModel{}).Where(&db.CorrectionsModel{
		Nick:    e.Nick,
		Network: conn.Config.Name,
		Channel: e.Arguments[0],
	}).Rows()

	// Aggregate all items in a slice
	for rows.Next() {
		db.Get().Gorm.ScanRows(rows, &correction)

		userCorrections = append(userCorrections, correction)
	}

	// Read user log size from config
	userLogSize := config.Get().Modules.Corrections.UserLogSize

	if len(userCorrections) > userLogSize {
		// Remove all but the configured last items in the correction log
		for _, correction := range userCorrections[:len(userCorrections)-userLogSize] {
			db.Get().Gorm.Unscoped().Delete(correction)
		}
	}
}
