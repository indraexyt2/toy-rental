package repository

import (
	"context"
	"final-project/entity"
	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type IPaymentRepository interface {
	IBaseRepository[entity.Payment]
	FindByOrderID(ctx context.Context, orderID string) (entity.Payment, error)
	FindByRentalID(ctx context.Context, rentalID string) ([]entity.Payment, error)
	UpdateByID(ctx context.Context, id string, payment *entity.Payment) error
	SavePaymentWithMetadata(ctx context.Context, payment *entity.Payment) error
}

type PaymentRepository struct {
	BaseRepository[entity.Payment]
}

func NewPaymentRepository(db *gorm.DB) IPaymentRepository {
	return &PaymentRepository{
		BaseRepository: BaseRepository[entity.Payment]{DB: db},
	}
}

func (r *PaymentRepository) FindByOrderID(ctx context.Context, orderID string) (entity.Payment, error) {
	var payment entity.Payment

	if err := r.DB.WithContext(ctx).Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		return entity.Payment{}, err
	}

	return payment, nil
}

func (r *PaymentRepository) FindByRentalID(ctx context.Context, rentalID string) ([]entity.Payment, error) {
	var payments []entity.Payment

	if err := r.DB.WithContext(ctx).Where("rental_id = ?", rentalID).Find(&payments).Error; err != nil {
		return nil, err
	}

	return payments, nil
}

func (r *PaymentRepository) UpdateByID(ctx context.Context, id string, payment *entity.Payment) error {
	uuid, err := uuid.FromString(id)
	if err != nil {
		return err
	}

	return r.DB.WithContext(ctx).Model(&entity.Payment{}).Where("id = ?", uuid).Updates(map[string]interface{}{
		"transaction_status": payment.TransactionStatus,
		"payment_method":     payment.PaymentMethod,
		"va_number":          payment.VANumber,
		"transaction_time":   payment.TransactionTime,
		"fraud_status":       payment.FraudStatus,
	}).Error
}

func (r *PaymentRepository) SavePaymentWithMetadata(ctx context.Context, payment *entity.Payment) error {
	return r.DB.WithContext(ctx).Create(payment).Error
}
