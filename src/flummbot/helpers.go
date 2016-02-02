package flummbot

import (
	"database/sql"
	"fmt"
	"github.com/fluffle/goirc/client"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

type Helpers struct {
	Config *Config
}

func (h *Helpers) SetupDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", h.Config.Database.File)

	if err != nil {
		fmt.Println("Failed to open database:", err)

		os.Exit(1)
	}

	return db
}

func (h *Helpers) RegisterCallbacks(c *client.Conn, q chan bool) *client.Conn {
	c.HandleFunc(
		client.CONNECTED,
		func(conn *client.Conn, line *client.Line) {
			// Identify to services
			if len(h.Config.Connection.NickservIdentify) > 0 {
				conn.Privmsg(
					"NickServ",
					"IDENTIFY "+h.Config.Connection.NickservIdentify,
				)
			}

			// Sleep while auth happens
			time.Sleep(time.Second)

			// Then join channel
			conn.Join(h.Config.Connection.Channel)

			// Greet everyone
			conn.Privmsg(h.Config.Connection.Channel, "Hejj")
		},
	)

	c.HandleFunc(
		client.DISCONNECTED,
		func(conn *client.Conn, line *client.Line) {
			q <- true
		},
	)

	return c
}
