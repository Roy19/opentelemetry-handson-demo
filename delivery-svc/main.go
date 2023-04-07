package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Roy19/distributed-transaction-2pc/db"
	"github.com/Roy19/distributed-transaction-2pc/delivery-svc/controllers"
	"github.com/Roy19/distributed-transaction-2pc/delivery-svc/dto"
	"github.com/Roy19/distributed-transaction-2pc/delivery-svc/models"
	"github.com/Roy19/distributed-transaction-2pc/delivery-svc/repository"
	"github.com/Roy19/distributed-transaction-2pc/utils"
	"github.com/go-chi/chi/v5"
)

func initRoutes(mux *chi.Mux, controller *controllers.DeliveryAgentController) {
	mux.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Route("/agent", func(r chi.Router) {

		r.Post("/reserve", func(w http.ResponseWriter, r *http.Request) {
			id, err := controller.ReserveDeliveryAgent()
			if err != nil {
				errorMessage := map[string]any{
					"error": err.Error(),
				}
				utils.Respond(w, http.StatusNotFound, errorMessage)
				return
			}
			data := map[string]any{
				"message": "delivery agent reserved",
				"id":      id,
			}
			utils.Respond(w, http.StatusOK, data)
		})

		r.Post("/book", func(w http.ResponseWriter, r *http.Request) {
			var bookDeliveryAgent dto.BookDeliveryAgentDto
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				errorMessage := map[string]any{
					"error": "failed to read json",
				}
				utils.Respond(w, http.StatusBadRequest, errorMessage)
				return
			}
			defer r.Body.Close()
			err = json.Unmarshal(data, &bookDeliveryAgent)
			if err != nil {
				errorMessage := map[string]any{
					"error": "failed to unmarshal json",
				}
				utils.Respond(w, http.StatusBadRequest, errorMessage)
				return
			}
			err = controller.BookDeliveryAgent(bookDeliveryAgent.ReservationID, bookDeliveryAgent.OrderID)
			if err != nil {
				errorMessage := map[string]any{
					"error": err.Error(),
				}
				utils.Respond(w, http.StatusInternalServerError, errorMessage)
				return
			} else {
				data := map[string]any{
					"message": "delivery agent booked",
				}
				utils.Respond(w, http.StatusOK, data)
			}
		})
	})
}

func initDependencies() *controllers.DeliveryAgentController {
	store_dsn := os.Getenv("DELIVERY_DSN")
	db.InitDB(store_dsn, "delivery-svc")
	db.MigrateModels("delivery-svc", models.DeliveryAgentReservation{})
	db.PutDummyDataDeliveryAgent("delivery-svc")
	return &controllers.DeliveryAgentController{
		DeliveryAgentRepository: &repository.DeliveryAgentRepository{},
	}
}

func main() {
	mux := chi.NewRouter()
	controller := initDependencies()
	initRoutes(mux, controller)
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatal("failed to start server")
	}
}
