package controllers

import (
	"log"

	"github.com/Roy19/distributed-transaction-2pc/store-svc/repository"
)

type StoreController struct {
	StoreRepository *repository.StoreRepository
}

func (c *StoreController) GetItem(itemID int64) error {
	_, err := c.StoreRepository.GetItem(itemID)
	if err != nil {
		log.Printf("[ERROR] Failed to get item from db\n")
	}
	return err
}

func (c *StoreController) ReserveItem(itemID int64) (uint, error) {
	id, err := c.StoreRepository.CreateReservation(itemID)
	if err != nil {
		log.Printf("[ERROR] Failed to create a reservation on that item\n")
	}
	return id, err
}

func (c *StoreController) BookItem(reservationID int64, orderID string) error {
	err := c.StoreRepository.BookItem(reservationID, orderID)
	if err != nil {
		log.Printf("[ERROR] Failed to book the item")
	}
	return err
}
