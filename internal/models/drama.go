package models

import "gorm.io/gorm"

// Drama represents the drama model structure optimized for Go
type Drama struct {
	gorm.Model
	Title  string `json:"title" gorm:"not null"`
	Status string `json:"status" gorm:"default:'Watching'"`
	Rating int    `json:"rating"`
	Genre  string `json:"genre"`
}