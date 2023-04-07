package dto

type ItemBookingDto struct {
	ReservationID int64  `json:"reservationId"`
	OrderID       string `json:"orderId"`
}
