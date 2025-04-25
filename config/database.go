package config

import (
	"final-project/entity"
	"final-project/utils/helpers"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

type Database struct {
	*gorm.DB
}

func NewDatabase(config *Config) *Database {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		config.DBHost,
		config.DBUser,
		config.DBPassword,
		config.DBName,
		config.DBPort,
	)

	// Buka koneksi
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	helpers.Logger.Info("Connected to database")

	return &Database{db}
}

// AutoMigrate
func (db *Database) AutoMigrate() error {
	return db.DB.AutoMigrate(
		&entity.User{},
		&entity.ToyCategory{},
		&entity.Toy{},
		&entity.ToyImage{},
		&entity.Rental{},
		&entity.RentalItem{},
		&entity.Payment{},
		&entity.UserToken{},
	)
}

// CloseConnection menutup koneksi database
func (db *Database) CloseConnection() {
	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Printf("Error getting database instance: %v", err)
		return
	}
	sqlDB.Close()
}
