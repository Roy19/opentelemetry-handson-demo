package dto

type DeliveryAgentBookingDto struct {
	ReservationID int64  `json:"reservationId"`
	OrderID       string `json:"orderId"`
}
