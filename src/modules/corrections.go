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
	var rows []db.CorrectionsModel
	var correction db.CorrectionsModel

	msg := strings.Trim(e.Message(), " ")

	// Fetch database entries to consider
	db.Get().Gorm.Model(&db.CorrectionsModel{}).Where(&db.CorrectionsModel{
		Nick:    strings.ToLower(e.Nick),
		Network: conn.Config.Name,
		Channel: e.Arguments[0],
	}).Order("ID DESC").Find(&rows)

	// Record messages for this user if it wasn't a correction
	if c.processCorrections(msg, rows, conn, e) == false {
		correction = db.CorrectionsModel{
			Nick:    strings.ToLower(e.Nick),
			Type:    e.Code,
			Body:    msg,
			Network: conn.Config.Name,
			Channel: e.Arguments[0],
		}

		db.Get().Gorm.Create(&correction)
	}

	// Select all items
	db.Get().Gorm.Model(&db.CorrectionsModel{}).Where(&db.CorrectionsModel{
		Nick:    strings.ToLower(e.Nick),
		Network: conn.Config.Name,
		Channel: e.Arguments[0],
	}).Find(&rows)

	// Read user log size from config
	userLogSize := config.Get().Modules.Corrections.UserLogSize

	if len(rows) > userLogSize {
		// Remove all but the configured last items in the correction log
		for _, row := range rows[:len(rows)-userLogSize] {
			db.Get().Gorm.Unscoped().Delete(row)
		}
	}
}

func (c Corrections) processCorrections(
	msg string,
	rows []db.CorrectionsModel,
	conn *irc.IrcConnection,
	e *ircevent.Event,
) bool {
	var correction db.CorrectionsModel
	prefixes := make(map[string]bool)
	format := irc.GetFormat()

	// Build a map of separator with the key as value for lookup of the separator.
	for _, value := range config.Get().Modules.Corrections.Separators {
		prefixes["s"+value] = true
	}

	for _, row := range rows {
		for prefix := range prefixes {
			if len(msg) > 1 && strings.Index(msg, prefix) == 0 {
				// Split replaement message
				subs := strings.SplitN(msg, prefix[1:], 3)

				// If we have a replacement match
				if len(subs) == 3 && strings.Contains(row.Body, subs[1]) {
					trailSeparatorIndex := strings.LastIndex(subs[2], prefix[1:])
					// Find out if we have a trailing separator
					if trailSeparatorIndex > 0 && len(subs[2])-len(prefix[1:]) == trailSeparatorIndex {
						// And cut it off the end
						subs[2] = subs[2][0:trailSeparatorIndex]
					}

					// Correct string
					corrected := strings.ReplaceAll(row.Body, subs[1], subs[2])

					// Store in model
					correction = db.CorrectionsModel{
						Nick:    strings.ToLower(e.Nick),
						Type:    correction.Type,
						Body:    corrected,
						Network: conn.Config.Name,
						Channel: e.Arguments[0],
					}

					// Save the corrected one to the database as a new entry
					db.Get().Gorm.Create(&correction)

					// Have a different prefix before message if it's an ACTION message
					prefix := ""
					if row.Type == "CTCP_ACTION" {
						prefix = "* " + format.Bold + e.Nick + format.Reset + ": "
					}

					// Respond on IRC
					conn.IrcEventConnection.Privmsgf(
						e.Arguments[0],
						config.Get().Modules.Corrections.Message,
						format.Bold+e.Nick+format.Reset,
						prefix+format.Italics+corrected,
					)

					return true
				}
			}
		}
	}

	return false
}
