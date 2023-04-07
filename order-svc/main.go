package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/Roy19/distributed-transaction-2pc/order-svc/dto"
	distributedTracer "github.com/Roy19/distributed-transaction-2pc/tracer"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func done(wg *sync.WaitGroup) {
	wg.Done()
}

func createOrder() string {
	return uuid.New().String()
}

func coordinator(itemID int, wg *sync.WaitGroup, span opentracing.Span) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/store/item/"+strconv.Itoa(itemID), nil)

	span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error fetching item from store-svc: ", err)
		span.SetTag("error", true)
		done(wg)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error fetching item from store-svc: ", resp.StatusCode)
		span.SetTag("error", true)
		done(wg)
	}

	// reserve item
	req, _ = http.NewRequest(
		"POST",
		"http://localhost:8080/store/item/"+strconv.Itoa(itemID)+"/reserve",
		nil,
	)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error reserving item from store-svc: ", err)
		span.SetTag("error", true)
		done(wg)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error reserving item from store-svc: ", resp.StatusCode)
		span.SetTag("error", true)
		done(wg)
	}
	var itemReservationDto dto.ItemReservationDto
	err = json.NewDecoder(resp.Body).Decode(&itemReservationDto)
	if err != nil {
		log.Println("Error decoding item reservation response: ", err)
		span.SetTag("error", true)
		done(wg)
	}

	// reserve delivery agent
	req, _ = http.NewRequest(
		"POST",
		"http://localhost:8081/agent/reserve",
		nil,
	)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error reserving delivery agent from delivery-svc: ", err)
		span.SetTag("error", true)
		done(wg)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error reserving delivery agent from delivery-svc: ", resp.StatusCode)
		span.SetTag("error", true)
		done(wg)
	}
	var deliveryReservationDto dto.DeliveryAgentReservationDto
	err = json.NewDecoder(resp.Body).Decode(&deliveryReservationDto)
	if err != nil {
		log.Println("Error decoding delivery agent reservation response: ", err)
		span.SetTag("error", true)
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
	req, _ = http.NewRequest(
		"POST",
		"http://localhost:8080/store/item/"+strconv.Itoa(itemID)+"/book",
		toSend,
	)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error booking item from store-svc: ", err)
		span.SetTag("error", true)
		done(wg)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error booking item from store-svc: ", resp.StatusCode)
		span.SetTag("error", true)
		done(wg)
	}

	// book delivery agent
	deliveryAgentBooking := dto.DeliveryAgentBookingDto{
		OrderID:       orderID,
		ReservationID: deliveryReservationDto.ReservationID,
	}
	data, _ = json.Marshal(deliveryAgentBooking)
	toSend = bytes.NewBuffer(data)
	req, _ = http.NewRequest(
		"POST",
		"http://localhost:8081/agent/book",
		toSend,
	)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error booking delivery agent from delivery-svc: ", err)
		span.SetTag("error", true)
		done(wg)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error booking delivery agent from delivery-svc: ", resp.StatusCode)
		span.SetTag("error", true)
		done(wg)
	}

	log.Printf("Order %s created\n", orderID)
	done(wg)
}

func main() {
	tracer, err := distributedTracer.GetTracer("order-svc", os.Getenv("JAEGER_AGENT_HOST"))
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	for i := 1; i < 10; i++ {
		span := tracer.StartSpan("order-svc: Create Order", ext.RPCServerOption(nil))
		defer span.Finish()
		go coordinator(1, wg, span)
		wg.Add(1)
	}
	wg.Wait()
}
