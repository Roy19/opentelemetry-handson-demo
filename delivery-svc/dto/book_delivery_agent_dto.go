package dto

type BookDeliveryAgentDto struct {
	ReservationID int64  `json:"reservationId"`
	OrderID       string `json:"orderId"`
}
