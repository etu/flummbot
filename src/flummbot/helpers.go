package flummbot

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

type Helpers struct {
	Config Config
}

func (h *Helpers) SetupDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "./flummbot.db")

	if err != nil {
		fmt.Println("Failed to open database:", err)

		os.Exit(1)
	}

	return db
}
