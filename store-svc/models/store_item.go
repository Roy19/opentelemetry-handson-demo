package models

import "gorm.io/gorm"

type StoreItem struct {
	gorm.Model
	Name string `gorm:"not null"`
}
