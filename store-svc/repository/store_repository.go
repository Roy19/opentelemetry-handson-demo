package repository

import (
	"context"
	"fmt"

	"github.com/Roy19/distributed-transaction-2pc/db"
	"github.com/Roy19/distributed-transaction-2pc/store-svc/models"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type StoreRepository struct {
}

func (s *StoreRepository) GetItem(ctx context.Context, itemID int64) (int64, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetItem: get_item_availability in db")
	defer span.Finish()

	var item models.StoreItem
	txOut := db.GetDBClient("store-svc").First(&item, itemID)
	if txOut.Error == gorm.ErrRecordNotFound {
		return 0, fmt.Errorf("item not found")
	}
	return int64(item.ID), nil
}

func (s *StoreRepository) CreateReservation(ctx context.Context, itemID int64) (uint, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "CreateReservation: create_reservation in db")
	defer span.Finish()

	txn := db.GetDBClient("store-svc").Model(&models.StoreItemReservation{}).Begin()
	var storeReservation models.StoreItemReservation
	txn = txn.Raw(`select * from store_item_reservations 
		where is_reserved = false and current_order_id is null and 
		store_item_id = ?
		for update`, int(itemID)).Scan(&storeReservation)
	if txn.Error != nil || txn.RowsAffected == 0 {
		txn.Rollback()
		return 0, fmt.Errorf("no more reservations can be done on item")
	}
	txn = txn.Exec(`update store_item_reservations
			set is_reserved = true
			where id = ?`, storeReservation.ID)
	if txn.Error != nil {
		txn.Rollback()
		span.SetTag("error", true)
		return 0, fmt.Errorf("failed to set lock on store item")
	}
	txn.Commit()
	return storeReservation.ID, nil
}

func (c *StoreRepository) BookItem(ctx context.Context, reservationID int64, orderID string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "BookItem: book_item in db")
	defer span.Finish()

	txn := db.GetDBClient("store-svc").Model(&models.StoreItemReservation{}).Begin()
	var storeReservation models.StoreItemReservation
	txn = txn.Raw(`select * from store_item_reservations 
		where is_reserved = true and id = ?
		for update`, uint(reservationID)).Scan(&storeReservation)
	if txn.Error != nil || txn.RowsAffected == 0 {
		txn.Rollback()
		return fmt.Errorf("no more reservations can be done on item")
	}
	txn = txn.Exec(`update store_item_reservations
			set is_reserved = false, current_order_id = ?
			where id = ?`, orderID, uint(reservationID))
	if txn.Error != nil {
		txn.Rollback()
		span.SetTag("error", true)
		return fmt.Errorf("failed to set lock on store item")
	}
	txn.Commit()
	return nil
}
