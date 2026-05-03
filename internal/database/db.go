package database

import (
	"github.com/glebarez/sqlite"
	"github.com/smgarciagr/codeact-agent-drama/internal/models"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	// Creates the local kdrama.db file
	DB, err = gorm.Open(sqlite.Open("kdrama.db"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to the database")
	}

	// Migration: Creates the table if it doesn't exist
	DB.AutoMigrate(&models.Drama{})

	// Seed: Insert initial dramas
	var count int64
	DB.Model(&models.Drama{}).Count(&count)
	if count == 0 {
		dramas := []models.Drama{
			{Title: "Crash Landing on You", Status: "Finished", Rating: 10, Genre: "Romance"},
			{Title: "Goblin", Status: "Finished", Rating: 9, Genre: "Fantasy"},
		}
		DB.Create(&dramas)
	}
}
