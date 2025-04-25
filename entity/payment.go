package entity

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"
)

const (
	PaymentTypeRental    = "rental"
	PaymentTypeLateFee   = "late_fee"
	PaymentTypeDamageFee = "damage_fee"
	PaymentTypeCombined  = "combined"
)

const (
	TransactionStatusPending       = "pending"
	TransactionStatusCapture       = "capture"
	TransactionStatusSettlement    = "settlement"
	TransactionStatusDeny          = "deny"
	TransactionStatusCancel        = "cancel"
	TransactionStatusExpire        = "expire"
	TransactionStatusFailure       = "failure"
	TransactionStatusRefund        = "refund"
	TransactionStatusPartialRefund = "partial_refund"
)

type Payment struct {
	BaseEntity
	RentalID          uuid.UUID  `gorm:"type:uuid;not null" json:"rental_id"`
	OrderID           string     `gorm:"null" json:"order_id"`
	PaymentType       string     `gorm:"size:50;not null;check:payment_type IN ('rental', 'late_fee', 'damage_fee', 'combined', 'extension')" json:"payment_type"`
	GrossAmount       float64    `gorm:"type:decimal(10,2);not null" json:"gross_amount"`
	SnapToken         string     `gorm:"type:text" json:"snap_token"`
	SnapURL           string     `gorm:"type:text" json:"snap_url"`
	ExpiryTime        *time.Time `json:"expiry_time"`
	TransactionTime   *time.Time `json:"transaction_time"`
	TransactionStatus string     `gorm:"size:50" json:"transaction_status"`
	PaymentMethod     string     `gorm:"size:50" json:"payment_method"`
	VANumber          string     `gorm:"size:100" json:"va_number"`
	FraudStatus       string     `gorm:"size:50" json:"fraud_status"`
	Metadata          []byte     `gorm:"type:jsonb" json:"-"`

	Rental Rental `gorm:"foreignKey:RentalID" json:"-"`
}

func (*Payment) TableName() string {
	return "payments"
}

type CreatePaymentRequest struct {
	RentalID string `json:"rental_id" binding:"required"`
}

func (p *Payment) GetExtensionMetadata() (*ExtensionMetadata, error) {
	if p.PaymentType != PaymentTypeExtension || len(p.Metadata) == 0 {
		return nil, nil
	}
	var metadata ExtensionMetadata
	err := json.Unmarshal(p.Metadata, &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (p *Payment) SetExtensionMetadata(metadata *ExtensionMetadata) error {
	if metadata == nil {
		p.Metadata = nil
		return nil
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	p.Metadata = data
	return nil
}
