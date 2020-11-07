package db

import "github.com/jinzhu/gorm"

type CorrectionsModel struct {
	gorm.Model
	Nick    string `gorm:"size:32"`
	Type    string `gorm:"size:11"`
	Body    string `gorm:"size:512"`
	Network string `gorm:"size:64"`
	Channel string `gorm:"size:64"`
}

type KarmaModel struct {
	gorm.Model
	Item   string `gorm:"unique;not null"`
	Points int
}

type QuotesModel struct {
	gorm.Model
	Nick    string `gorm:"size:32"`
	Body    string `gorm:"size:512"`
	Network string `gorm:"size:64"`
	Channel string `gorm:"size:64"`
}

type TellsModel struct {
	gorm.Model
	From    string `gorm:"size:32"`
	To      string `gorm:"size:32"`
	Network string `gorm:"size:64"`
	Channel string `gorm:"size:64"`
	Body    string `gorm:"size:512"`
}
