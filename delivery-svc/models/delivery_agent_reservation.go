package models

import (
	"database/sql"

	"gorm.io/gorm"
)

type DeliveryAgentReservation struct {
	gorm.Model
	IsReserved     bool
	CurrentOrderID sql.NullString
}
