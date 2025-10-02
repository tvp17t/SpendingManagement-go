package db

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Open opens a SQLite database file (./data.db). Pure Go driver (no CGO).
func Open() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1) // SQLite: single writer
	}
	log.Println("DB ready: sqlite ./data.db")
	return db, nil
}
