package repository

import (
	"context"
	"final-project/entity"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type IRentalRepository interface {
	IBaseRepository[entity.Rental]
	UpdateToyStock(ctx context.Context, toyID string, quantity int) error
	ReturnRental(ctx context.Context, rental *entity.Rental) error
	UpdateRentalItem(ctx context.Context, rentalItem *entity.RentalItem) error
	UpdatePaymentStatus(ctx context.Context, rentalID string, status string) error
	UpdateStatus(ctx context.Context, rentalID string, status string) error
	ExtendRental(ctx context.Context, rental *entity.Rental, newExpectedReturnDate time.Time, additionalCost float64, notes string) error
	RollbackExtension(ctx context.Context, rentalID string, oldExpectedReturnDate time.Time, oldPrice float64) error
}

type RentalRepository struct {
	BaseRepository[entity.Rental]
}

func NewRentalRepository(db *gorm.DB) IRentalRepository {
	return &RentalRepository{
		BaseRepository: BaseRepository[entity.Rental]{DB: db},
	}
}

func (r *RentalRepository) FindById(ctx context.Context, id string) (entity.Rental, error) {
	var model entity.Rental
	if err := r.DB.WithContext(ctx).Where("id = ?", id).
		Preload("RentalItems").
		Preload("RentalItems.Toy").
		Preload("User").
		First(&model).Error; err != nil {
		return model, err
	}
	return model, nil
}

func (r *RentalRepository) Insert(ctx context.Context, model *entity.Rental) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("RentalItems").Create(model).Error; err != nil {
			return err
		}

		for i := range model.RentalItems {
			model.RentalItems[i].RentalID = model.ID
			if err := tx.Create(&model.RentalItems[i]).Error; err != nil {
				return err
			}

			if err := tx.Model(&entity.Toy{}).Where("id = ?", model.RentalItems[i].ToyID).
				UpdateColumn("stock", gorm.Expr("stock - ?", model.RentalItems[i].Quantity)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *RentalRepository) UpdateToyStock(ctx context.Context, toyID string, quantity int) error {
	return r.DB.WithContext(ctx).Model(&entity.Toy{}).Where("id = ?", toyID).
		UpdateColumn("stock", gorm.Expr("stock - ?", quantity)).Error
}

func (r *RentalRepository) ReturnRental(ctx context.Context, rental *entity.Rental) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(rental).
			Select("status", "actual_return_date", "late_fee", "damage_fee", "total_amount").
			Updates(rental).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *RentalRepository) UpdateRentalItem(ctx context.Context, rentalItem *entity.RentalItem) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(rentalItem).
			Select("condition_after", "damage_description", "damage_fee", "status").
			Updates(rentalItem).Error; err != nil {
			return err
		}

		if rentalItem.Status == "returned" {
			if err := tx.Model(&entity.Toy{}).
				Where("id = ?", rentalItem.ToyID).
				UpdateColumn("stock", gorm.Expr("stock + ?", rentalItem.Quantity)).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *RentalRepository) UpdatePaymentStatus(ctx context.Context, rentalID string, status string) error {
	return r.DB.WithContext(ctx).Model(&entity.Rental{}).Where("id = ?", rentalID).
		Update("payment_status", status).Error
}

func (r *RentalRepository) UpdateStatus(ctx context.Context, rentalID string, status string) error {
	return r.DB.WithContext(ctx).Model(&entity.Rental{}).Where("id = ?", rentalID).
		Update("status", status).Error
}

func (r *RentalRepository) ExtendRental(ctx context.Context, rental *entity.Rental, newExpectedReturnDate time.Time, additionalCost float64, notes string) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Perbarui tanggal pengembalian yang diharapkan, total harga rental, dan catatan
		fmt.Println("Additional Cost:", additionalCost)
		updateMap := map[string]interface{}{
			"expected_return_date": newExpectedReturnDate,
			"total_rental_price":   gorm.Expr("total_rental_price + ?", additionalCost),
		}

		if notes != "" {
			updateMap["notes"] = gorm.Expr("notes || '\n' || ?", notes)
		}

		if err := tx.Model(rental).
			Where("id = ?", rental.ID).
			Updates(updateMap).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *RentalRepository) RollbackExtension(ctx context.Context, rentalID string, oldExpectedReturnDate time.Time, oldPrice float64) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Rental{}).
			Where("id = ?", rentalID).
			Updates(map[string]interface{}{
				"expected_return_date": oldExpectedReturnDate,
				"total_rental_price":   oldPrice,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}
