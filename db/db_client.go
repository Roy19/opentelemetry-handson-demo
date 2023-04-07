package db

import (
	"log"
	"sync"

	deliveryAgentSvcModels "github.com/Roy19/distributed-transaction-2pc/delivery-svc/models"
	storeSvcModels "github.com/Roy19/distributed-transaction-2pc/store-svc/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dbClients map[string]*gorm.DB
	dbonce    sync.Once
)

func InitDB(dsn string, svcName string) {
	dbonce.Do(func() {
		dbClients = make(map[string]*gorm.DB)
		dbClient, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Error connecting to database: %v\n", err)
		}
		dbClients[svcName] = dbClient
	})
}

func GetDBClient(svcName string) *gorm.DB {
	return dbClients[svcName]
}

func MigrateModels(svcName string, models ...interface{}) {
	if dbClients[svcName] != nil {
		for _, model := range models {
			dbClients[svcName].AutoMigrate(&model)
		}
	}
}

func PutDummyDataStoreSvc(svcName string) {
	if dbClients[svcName] != nil {
		storeItem := storeSvcModels.StoreItem{
			Name: "iPhone 12",
		}
		dbClients[svcName].Create(&storeItem)
		storeItemReservations := []storeSvcModels.StoreItemReservation{
			{
				StoreItem:  storeItem,
				IsReserved: false,
			},
			{
				StoreItem:  storeItem,
				IsReserved: false,
			},
			{
				StoreItem:  storeItem,
				IsReserved: false,
			},
			{
				StoreItem:  storeItem,
				IsReserved: false,
			},
			{
				StoreItem:  storeItem,
				IsReserved: false,
			},
			{
				StoreItem:  storeItem,
				IsReserved: false,
			},
			{
				StoreItem:  storeItem,
				IsReserved: false,
			},
			{
				StoreItem:  storeItem,
				IsReserved: false,
			},
			{
				StoreItem:  storeItem,
				IsReserved: false,
			},
			{
				StoreItem:  storeItem,
				IsReserved: false,
			},
		}
		dbClients[svcName].Create(&storeItemReservations)
	}
}

func PutDummyDataDeliveryAgent(svcName string) {
	if dbClients[svcName] != nil {
		deliveryAgentReservations := []deliveryAgentSvcModels.DeliveryAgentReservation{
			{
				IsReserved: false,
			},
			{
				IsReserved: false,
			},
			{
				IsReserved: false,
			},
			{
				IsReserved: false,
			},
			{
				IsReserved: false,
			},
			{
				IsReserved: false,
			},
			{
				IsReserved: false,
			},
			{
				IsReserved: false,
			},
			{
				IsReserved: false,
			},
			{
				IsReserved: false,
			},
		}
		dbClients[svcName].Create(&deliveryAgentReservations)
	}
}
