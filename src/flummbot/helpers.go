package flummbot

import (
	"database/sql"
	"fmt"
	"github.com/fluffle/goirc/client"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

type Helpers struct {
	Config Config
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
