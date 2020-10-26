package modules

import (
	"github.com/etu/flummbot/src/config"
	"github.com/etu/flummbot/src/db"
	"github.com/etu/flummbot/src/irc"
	"github.com/jinzhu/gorm"
	ircevent "github.com/thoj/go-ircevent"
	"strings"
)

type TellsModel struct {
	gorm.Model
	From    string `gorm:"size:32"`
	To      string `gorm:"size:32"`
	Network string `gorm:"size:64"`
	Channel string `gorm:"size:64"`
	Body    string `gorm:"size:512"`
}

type Tells struct {
	Config *config.ClientConfig
	Db     *db.Db
}

func (t Tells) DbSetup() {
	t.Db.Gorm.AutoMigrate(&TellsModel{})
}

func (t Tells) RegisterCallbacks(c *irc.IrcConnection) {
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

func (t Tells) register(c *irc.IrcConnection, e *ircevent.Event) {
	parts := strings.Split(e.Message(), " ")

	if parts[0] == t.Config.Modules.Tells.Command && len(parts) > 2 {
		t.Db.Gorm.Create(&TellsModel{
			From:    e.Nick,
			To:      strings.ToLower(parts[1]),
			Network: c.Config.Name,
			Channel: e.Arguments[0],
			Body:    strings.Join(parts[2:], " "),
		})

		c.IrcEventConnection.Privmsg(
			e.Arguments[0],
			"Alright, I'm going to tell "+parts[1]+": "+strings.Join(parts[2:], " "),
		)
	}
}

func (t Tells) deliver(c *irc.IrcConnection, e *ircevent.Event) {
	rows, _ := t.Db.Gorm.Model(&TellsModel{}).Where(&TellsModel{
		Network: c.Config.Name,
		Channel: e.Arguments[0],
		To:      strings.ToLower(e.Nick),
	}).Rows()
	defer rows.Close()

	// Make map with rows to delete from database
	toDelete := make(map[uint]bool)

	for rows.Next() {
		var tell TellsModel

		t.Db.Gorm.ScanRows(rows, &tell)

		// Format the timestamp
		date := tell.CreatedAt.Format("2006-01-02 15:04:05")

		c.IrcEventConnection.Privmsg(
			tell.Channel,
			tell.To+": \""+tell.Body+"\" -- "+tell.From+" @ "+date,
		)

		toDelete[tell.ID] = true
	}

	rows.Close()

	// Loop trough the map with ids to remove
	for id := range toDelete {
		t.Db.Gorm.Unscoped().Where("id = ?", id).Delete(&TellsModel{})
	}
}
