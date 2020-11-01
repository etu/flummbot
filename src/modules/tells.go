package modules

import (
	"strings"

	"github.com/etu/flummbot/src/config"
	"github.com/etu/flummbot/src/db"
	"github.com/etu/flummbot/src/irc"
	ircevent "github.com/thoj/go-ircevent"
)

type Tells struct{}

func (t Tells) RegisterCallbacks(c *irc.IrcConnection) {
	if config.Get().Modules.Tells.Enable {
		c.IrcEventConnection.AddCallback(
			"JOIN",
			func(e *ircevent.Event) {
				go t.deliver(c, e)
			},
		)

		c.IrcEventConnection.AddCallback(
			"PRIVMSG",
			func(e *ircevent.Event) {
				go t.deliver(c, e)
				go t.register(c, e)
			},
		)

		c.IrcEventConnection.AddCallback(
			"CTCP_ACTION",
			func(e *ircevent.Event) {
				go t.deliver(c, e)
			},
		)
	}
}

func (t Tells) register(c *irc.IrcConnection, e *ircevent.Event) {
	format := irc.GetFormat()
	parts := strings.Split(e.Message(), " ")

	if parts[0] == config.Get().Modules.Tells.Command && len(parts) > 2 {
		db.Get().Gorm.Create(&db.TellsModel{
			From:    e.Nick,
			To:      strings.ToLower(parts[1]),
			Network: c.Config.Name,
			Channel: e.Arguments[0],
			Body:    strings.Join(parts[2:], " "),
		})

		c.IrcEventConnection.Privmsgf(
			e.Arguments[0],
			"%sAlright, I'm going to tell %s%s:%s %s%s",
			format.Color+format.Colors.LightBlue,
			format.Color+format.Colors.LightCyan+format.Bold,
			parts[1],
			format.Reset,
			format.Italics,
			strings.Join(parts[2:], " "),
		)
	}
}

func (t Tells) deliver(c *irc.IrcConnection, e *ircevent.Event) {
	format := irc.GetFormat()

	rows, _ := db.Get().Gorm.Model(&db.TellsModel{}).Where(&db.TellsModel{
		Network: c.Config.Name,
		Channel: e.Arguments[0],
		To:      strings.ToLower(e.Nick),
	}).Rows()
	defer rows.Close()

	// Make map with rows to delete from database
	toDelete := make(map[uint]bool)

	for rows.Next() {
		var tell db.TellsModel

		db.Get().Gorm.ScanRows(rows, &tell)

		// Format the timestamp
		date := tell.CreatedAt.Format("2006-01-02 15:04:05")

		c.IrcEventConnection.Privmsgf(
			tell.Channel,
			"%s: '%s' %s--%s %s%s @ %s",
			format.Bold+format.Color+format.Colors.LightCyan+tell.To+format.Reset,
			format.Italics+tell.Body+format.Reset,
			format.Color+format.Colors.Gray,
			format.Reset,
			format.Color+format.Colors.Blue,
			tell.From,
			date,
		)

		toDelete[tell.ID] = true
	}

	rows.Close()

	// Loop trough the map with ids to remove
	for id := range toDelete {
		db.Get().Gorm.Unscoped().Where("id = ?", id).Delete(&db.TellsModel{})
	}
}
