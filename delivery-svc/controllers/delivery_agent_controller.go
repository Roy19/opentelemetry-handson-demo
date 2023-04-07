package controllers

import (
	"log"

	"github.com/Roy19/distributed-transaction-2pc/delivery-svc/repository"
)

type DeliveryAgentController struct {
	DeliveryAgentRepository *repository.DeliveryAgentRepository
}

func (c *DeliveryAgentController) ReserveDeliveryAgent() (uint, error) {
	id, err := c.DeliveryAgentRepository.CreateReservation()
	if err != nil {
		log.Printf("[ERROR] Failed to create a reservation on that item\n")
	}
	return id, err
}

func (c *DeliveryAgentController) BookDeliveryAgent(reservationID int64, orderID string) error {
	err := c.DeliveryAgentRepository.BookItem(reservationID, orderID)
	if err != nil {
		log.Printf("[ERROR] Failed to book the item")
	}
	return err
}
