package db

import (
	"github.com/etu/flummbot/src/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
)

type Db struct {
	Gorm *gorm.DB
}

var db Db

func New(config *config.ClientConfig) Db {
	if db.Gorm == nil {
		conn, err := gorm.Open(config.Database.Dialect, config.Database.Args)

		if err != nil {
			log.Fatal(err)
		}

		conn.AutoMigrate(&CorrectionsModel{})
		conn.AutoMigrate(&KarmaModel{})
		conn.AutoMigrate(&QuotesModel{})
		conn.AutoMigrate(&TellsModel{})

		db.Gorm = conn
	}

	return db
}

func Get() Db {
	return db
}
