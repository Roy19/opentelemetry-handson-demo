package dto

type BookItemDto struct {
	ReservationID int64  `json:"reservationId"`
	OrderID       string `json:"orderId"`
}
