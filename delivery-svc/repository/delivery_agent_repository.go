package repository

import (
	"fmt"

	"github.com/Roy19/distributed-transaction-2pc/db"
	"github.com/Roy19/distributed-transaction-2pc/delivery-svc/models"
)

type DeliveryAgentRepository struct {
}

func (s *DeliveryAgentRepository) CreateReservation() (uint, error) {
	txn := db.GetDBClient("delivery-svc").Model(&models.DeliveryAgentReservation{}).Begin()
	var deliveryAgentReservation models.DeliveryAgentReservation
	txn = txn.Raw(`select * from delivery_agent_reservations 
		where is_reserved = false and current_order_id is null
		limit 1
		for update`).Scan(&deliveryAgentReservation)
	if txn.Error != nil || txn.RowsAffected == 0 {
		txn.Rollback()
		return 0, fmt.Errorf("no more reservations can be done on item")
	}
	txn = txn.Exec(`update delivery_agent_reservations
			set is_reserved = true
			where id = ?`, deliveryAgentReservation.ID)
	if txn.Error != nil {
		txn.Rollback()
		return 0, fmt.Errorf("failed to set lock on delivery agent reservation")
	}
	txn.Commit()
	return deliveryAgentReservation.ID, nil
}

func (c *DeliveryAgentRepository) BookItem(reservationID int64, orderID string) error {
	txn := db.GetDBClient("delivery-svc").Model(&models.DeliveryAgentReservation{}).Begin()
	var deliveryAgentReservation models.DeliveryAgentReservation
	txn = txn.Raw(`select * from delivery_agent_reservations 
		where is_reserved = true and id = ?
		for update`, uint(reservationID)).Scan(&deliveryAgentReservation)
	if txn.Error != nil || txn.RowsAffected == 0 {
		txn.Rollback()
		return fmt.Errorf("no more reservations can be done on item")
	}
	txn = txn.Exec(`update delivery_agent_reservations
			set is_reserved = false, current_order_id = ?
			where id = ?`, orderID, uint(reservationID))
	if txn.Error != nil {
		txn.Rollback()
		return fmt.Errorf("failed to set lock on delivery agent reservation")
	}
	txn.Commit()
	return nil
}
