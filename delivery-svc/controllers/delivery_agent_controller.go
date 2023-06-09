package controllers

import (
	"context"
	"log"

	"github.com/Roy19/distributed-transaction-2pc/delivery-svc/repository"
	"github.com/opentracing/opentracing-go"
)

type DeliveryAgentController struct {
	DeliveryAgentRepository *repository.DeliveryAgentRepository
}

func (c *DeliveryAgentController) ReserveDeliveryAgent(ctx context.Context) (uint, error) {
	span, _ := opentracing.StartSpanFromContext(ctx,
		"DeliveryAgentController.ReserveDeliveryAgent: reserve_delivery_agent")
	defer span.Finish()

	id, err := c.DeliveryAgentRepository.CreateReservation(
		opentracing.ContextWithSpan(ctx, span),
	)
	if err != nil {
		log.Printf("[ERROR] Failed to create a reservation on that item\n")
	}
	return id, err
}

func (c *DeliveryAgentController) BookDeliveryAgent(ctx context.Context,
	reservationID int64, orderID string) error {
	span, _ := opentracing.StartSpanFromContext(ctx,
		"DeliveryAgentController.BookDeliveryAgent: book_delivery_agent")
	defer span.Finish()

	err := c.DeliveryAgentRepository.BookItem(opentracing.ContextWithSpan(ctx, span),
		reservationID, orderID)
	if err != nil {
		log.Printf("[ERROR] Failed to book the item")
	}
	return err
}
