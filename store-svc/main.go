package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Roy19/distributed-transaction-2pc/db"
	"github.com/Roy19/distributed-transaction-2pc/store-svc/controllers"
	"github.com/Roy19/distributed-transaction-2pc/store-svc/dto"
	"github.com/Roy19/distributed-transaction-2pc/store-svc/models"
	"github.com/Roy19/distributed-transaction-2pc/store-svc/repository"
	"github.com/Roy19/distributed-transaction-2pc/utils"
	"github.com/go-chi/chi/v5"
)

func initRoutes(mux *chi.Mux, controller *controllers.StoreController) {
	mux.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Route("/store/item/{itemID}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			itemID := chi.URLParam(r, "itemID")
			itemIDAsInt, err := strconv.ParseInt(itemID, 10, 64)
			if err != nil {
				errorMessage := map[string]any{
					"error": "itemID is required",
				}
				utils.Respond(w, http.StatusBadRequest, errorMessage)
				return
			}
			err = controller.GetItem(itemIDAsInt)
			if err != nil {
				errorMessage := map[string]any{
					"error": err.Error(),
				}
				utils.Respond(w, http.StatusNotFound, errorMessage)
				return
			}
			data := map[string]any{
				"message": "item exists in stock",
			}
			utils.Respond(w, http.StatusOK, data)
		})

		r.Post("/reserve", func(w http.ResponseWriter, r *http.Request) {
			itemID := chi.URLParam(r, "itemID")
			itemIDAsInt, err := strconv.ParseInt(itemID, 10, 64)
			if err != nil {
				errorMessage := map[string]any{
					"error": "itemID is required",
				}
				utils.Respond(w, http.StatusBadRequest, errorMessage)
				return
			}
			id, err := controller.ReserveItem(itemIDAsInt)
			if err != nil {
				errorMessage := map[string]any{
					"error": err.Error(),
				}
				utils.Respond(w, http.StatusNotFound, errorMessage)
				return
			}
			data := map[string]any{
				"message": "item reserved",
				"id":      id,
			}
			utils.Respond(w, http.StatusOK, data)
		})

		r.Post("/book", func(w http.ResponseWriter, r *http.Request) {
			itemID := chi.URLParam(r, "itemID")
			_, err := strconv.ParseInt(itemID, 10, 64)
			if err != nil {
				errorMessage := map[string]any{
					"error": "itemID is required",
				}
				utils.Respond(w, http.StatusBadRequest, errorMessage)
				return
			}
			var bookItem dto.BookItemDto
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				errorMessage := map[string]any{
					"error": "failed to read json",
				}
				utils.Respond(w, http.StatusBadRequest, errorMessage)
				return
			}
			defer r.Body.Close()
			err = json.Unmarshal(data, &bookItem)
			if err != nil {
				errorMessage := map[string]any{
					"error": "failed to unmarshal json",
				}
				utils.Respond(w, http.StatusBadRequest, errorMessage)
				return
			}
			err = controller.BookItem(bookItem.ReservationID, bookItem.OrderID)
			if err != nil {
				errorMessage := map[string]any{
					"error": err.Error(),
				}
				utils.Respond(w, http.StatusInternalServerError, errorMessage)
				return
			} else {
				data := map[string]any{
					"message": "item booked",
				}
				utils.Respond(w, http.StatusOK, data)
			}
		})
	})
}

func initDependencies() *controllers.StoreController {
	store_dsn := os.Getenv("STORE_DSN")
	db.InitDB(store_dsn, "store-svc")
	db.MigrateModels("store-svc", models.StoreItem{}, models.StoreItemReservation{})
	db.PutDummyDataStoreSvc("store-svc")
	return &controllers.StoreController{
		StoreRepository: &repository.StoreRepository{},
	}
}

func main() {
	mux := chi.NewRouter()
	controller := initDependencies()
	initRoutes(mux, controller)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("failed to start server")
	}
}
