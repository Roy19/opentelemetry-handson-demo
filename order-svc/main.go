package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/Roy19/distributed-transaction-2pc/order-svc/dto"
	"github.com/google/uuid"
)

func done(wg *sync.WaitGroup) {
	wg.Done()
}

func createOrder() string {
	return uuid.New().String()
}

func coordinator(itemID int, wg *sync.WaitGroup) {
	client := &http.Client{}
	resp, err := client.Get("http://localhost:8080/store/item/" + strconv.Itoa(itemID))
	if err != nil {
		log.Println("Error fetching item from store-svc: ", err)
		done(wg)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error fetching item from store-svc: ", resp.StatusCode)
		done(wg)
	}

	// reserve item
	resp, err = client.Post("http://localhost:8080/store/item/"+strconv.Itoa(itemID)+"/reserve", "application/json", nil)
	if err != nil {
		log.Println("Error reserving item from store-svc: ", err)
		done(wg)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error reserving item from store-svc: ", resp.StatusCode)
		done(wg)
	}
	var itemReservationDto dto.ItemReservationDto
	err = json.NewDecoder(resp.Body).Decode(&itemReservationDto)
	if err != nil {
		log.Println("Error decoding item reservation response: ", err)
		done(wg)
	}

	// reserve delivery agent
	resp, err = client.Post("http://localhost:8081/agent/reserve", "application/json", nil)
	if err != nil {
		log.Println("Error reserving delivery agent from delivery-svc: ", err)
		done(wg)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error reserving delivery agent from delivery-svc: ", resp.StatusCode)
		done(wg)
	}
	var deliveryReservationDto dto.DeliveryAgentReservationDto
	err = json.NewDecoder(resp.Body).Decode(&deliveryReservationDto)
	if err != nil {
		log.Println("Error decoding delivery agent reservation response: ", err)
		done(wg)
	}

	// create order
	orderID := createOrder()

	// book store item
	itemBooking := dto.ItemBookingDto{
		OrderID:       orderID,
		ReservationID: itemReservationDto.ReservationID,
	}
	data, _ := json.Marshal(itemBooking)
	toSend := bytes.NewBuffer(data)
	resp, err = client.Post("http://localhost:8080/store/item/"+strconv.Itoa(itemID)+"/book", "application/json", toSend)
	if err != nil {
		log.Println("Error booking item from store-svc: ", err)
		done(wg)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error booking item from store-svc: ", resp.StatusCode)
		done(wg)
	}

	// book delivery agent
	deliveryAgentBooking := dto.DeliveryAgentBookingDto{
		OrderID:       orderID,
		ReservationID: deliveryReservationDto.ReservationID,
	}
	data, _ = json.Marshal(deliveryAgentBooking)
	toSend = bytes.NewBuffer(data)

	resp, err = client.Post("http://localhost:8081/agent/book", "application/json", toSend)
	if err != nil {
		log.Println("Error booking delivery agent from delivery-svc: ", err)
		done(wg)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error booking delivery agent from delivery-svc: ", resp.StatusCode)
		done(wg)
	}

	log.Printf("Order %s created\n", orderID)
	done(wg)
}

func main() {
	wg := &sync.WaitGroup{}
	for i := 1; i < 10; i++ {
		go coordinator(1, wg)
		wg.Add(1)
	}
	wg.Wait()
}
