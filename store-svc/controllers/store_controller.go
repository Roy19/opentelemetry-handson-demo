package controllers

import (
	"context"
	"log"

	"github.com/Roy19/distributed-transaction-2pc/store-svc/repository"
	"github.com/opentracing/opentracing-go"
)

type StoreController struct {
	StoreRepository *repository.StoreRepository
}

func (c *StoreController) GetItem(ctx context.Context, itemID int64) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "StoreController.GetItem: get_item_availability")
	defer span.Finish()

	_, err := c.StoreRepository.GetItem(ctx, itemID)
	if err != nil {
		log.Printf("[ERROR] Failed to get item from db\n")
	}
	return err
}

func (c *StoreController) ReserveItem(ctx context.Context, itemID int64) (uint, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "StoreController.ReserveItem: reserve_item")
	defer span.Finish()

	id, err := c.StoreRepository.CreateReservation(ctx, itemID)
	if err != nil {
		log.Printf("[ERROR] Failed to create a reservation on that item\n")
	}
	return id, err
}

func (c *StoreController) BookItem(ctx context.Context, reservationID int64, orderID string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "StoreController.BookItem: book_item")
	defer span.Finish()

	err := c.StoreRepository.BookItem(ctx, reservationID, orderID)
	if err != nil {
		log.Printf("[ERROR] Failed to book the item")
	}
	return err
}
