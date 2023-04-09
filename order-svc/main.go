package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Roy19/distributed-transaction-2pc/order-svc/dto"
	distributedTracer "github.com/Roy19/distributed-transaction-2pc/tracer"
	"github.com/Roy19/distributed-transaction-2pc/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var (
	tracer opentracing.Tracer
)

func createOrder() string {
	return uuid.New().String()
}

func coordinator(ctx context.Context, itemID int) {
	span, _ := opentracing.StartSpanFromContext(ctx, "coordinator")
	defer span.Finish()

	req, _ := http.NewRequestWithContext(ctx, "GET",
		"http://localhost:8080/store/item/"+strconv.Itoa(itemID), nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error fetching item from store-svc: ", err)
		span.SetTag("error", true)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error fetching item from store-svc: ", resp.StatusCode)
		span.SetTag("error", true)
	}

	// reserve item
	req, _ = http.NewRequestWithContext(ctx,
		"POST",
		"http://localhost:8080/store/item/"+strconv.Itoa(itemID)+"/reserve",
		nil,
	)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error reserving item from store-svc: ", err)
		span.SetTag("error", true)

	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error reserving item from store-svc: ", resp.StatusCode)
		span.SetTag("error", true)
	}
	var itemReservationDto dto.ItemReservationDto
	err = json.NewDecoder(resp.Body).Decode(&itemReservationDto)
	if err != nil {
		log.Println("Error decoding item reservation response: ", err)
		span.SetTag("error", true)
	}

	// reserve delivery agent
	req, _ = http.NewRequestWithContext(ctx,
		"POST",
		"http://localhost:8081/agent/reserve",
		nil,
	)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error reserving delivery agent from delivery-svc: ", err)
		span.SetTag("error", true)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error reserving delivery agent from delivery-svc: ", resp.StatusCode)
		span.SetTag("error", true)
	}
	var deliveryReservationDto dto.DeliveryAgentReservationDto
	err = json.NewDecoder(resp.Body).Decode(&deliveryReservationDto)
	if err != nil {
		log.Println("Error decoding delivery agent reservation response: ", err)
		span.SetTag("error", true)
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
	req, _ = http.NewRequestWithContext(ctx,
		"POST",
		"http://localhost:8080/store/item/"+strconv.Itoa(itemID)+"/book",
		toSend,
	)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error booking item from store-svc: ", err)
		span.SetTag("error", true)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error booking item from store-svc: ", resp.StatusCode)
		span.SetTag("error", true)
	}

	// book delivery agent
	deliveryAgentBooking := dto.DeliveryAgentBookingDto{
		OrderID:       orderID,
		ReservationID: deliveryReservationDto.ReservationID,
	}
	data, _ = json.Marshal(deliveryAgentBooking)
	toSend = bytes.NewBuffer(data)
	req, _ = http.NewRequestWithContext(ctx,
		"POST",
		"http://localhost:8081/agent/book",
		toSend,
	)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error booking delivery agent from delivery-svc: ", err)
		span.SetTag("error", true)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error booking delivery agent from delivery-svc: ", resp.StatusCode)
		span.SetTag("error", true)
	}

	log.Printf("Order %s created\n", orderID)
}

func registerRoutes(router *chi.Mux) {
	router.Post("/order", func(w http.ResponseWriter, r *http.Request) {
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))
		span := tracer.StartSpan("order-svc: Create Order", ext.RPCServerOption(spanCtx))
		defer span.Finish()

		ctx := opentracing.ContextWithSpan(r.Context(), span)

		var createOrderRequest dto.CreateOrderRequest
		err := json.NewDecoder(r.Body).Decode(&createOrderRequest)
		defer r.Body.Close()
		if err != nil {
			message := map[string]string{
				"message": "Error decoding request body",
			}
			span.SetTag("error", true)
			utils.Respond(w, http.StatusBadRequest, message)
			return
		}
		coordinator(ctx, createOrderRequest.ItemID)
		message := map[string]string{
			"message": "Order created",
		}
		utils.Respond(w, http.StatusOK, message)
	})
}

func main() {
	t, err := distributedTracer.GetTracer("order-svc", os.Getenv("JAEGER_AGENT_HOST"))
	if err != nil {
		log.Fatal(err)
	}
	tracer = t
	opentracing.SetGlobalTracer(tracer)

	router := chi.NewRouter()
	registerRoutes(router)

	if err := http.ListenAndServe(":8082", router); err != nil {
		log.Fatal(err)
	}
}
