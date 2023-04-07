package models

import (
	"database/sql"

	"gorm.io/gorm"
)

type StoreItemReservation struct {
	gorm.Model
	StoreItemID    int
	StoreItem      StoreItem
	IsReserved     bool
	CurrentOrderId sql.NullString
}
