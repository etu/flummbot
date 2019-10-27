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

var (
	db     Db
	dbConn *gorm.DB
	err    error
)

func New(config *config.ClientConfig) Db {
	if db.Gorm == nil {
		dbConn, err = gorm.Open(config.Database.Dialect, config.Database.Args)

		if err != nil {
			log.Fatal(err)
		}

		db.Gorm = dbConn
	}

	return db
}
