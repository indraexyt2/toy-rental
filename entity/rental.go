package entity

import (
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

const (
	RentalStatusPending   = "pending"
	RentalStatusActive    = "active"
	RentalStatusCompleted = "completed"
	RentalStatusOverdue   = "overdue"
	RentalStatusCancelled = "cancelled"
	PaymentTypeExtension  = "extension"
)

const (
	PaymentStatusUnpaid        = "unpaid"
	PaymentStatusPending       = "pending"
	PaymentStatusPaid          = "paid"
	PaymentStatusExpired       = "expired"
	PaymentStatusFailed        = "failed"
	PaymentStatusRefunded      = "refunded"
	PaymentStatusPartiallyPaid = "partially_paid"
)

var (
	ErrInvalidReturnDate       = errors.New("tanggal pengembalian harus setelah tanggal rental")
	ErrInvalidActualReturnDate = errors.New("tanggal pengembalian aktual tidak boleh sebelum tanggal rental")
)

type Rental struct {
	BaseEntity
	UserID             uuid.UUID  `gorm:"type:uuid;not null" json:"user_id,omitempty"`
	Status             string     `gorm:"size:50;not null;check:status IN ('pending', 'active', 'completed', 'overdue', 'cancelled', 'extension')" json:"status,omitempty"`
	RentalDate         time.Time  `gorm:"not null" json:"rental_date,omitempty"`
	ExpectedReturnDate time.Time  `gorm:"not null" json:"expected_return_date,omitempty"`
	ActualReturnDate   *time.Time `json:"actual_return_date,omitempty"`
	TotalRentalPrice   float64    `gorm:"type:decimal(10,2);not null" json:"total_rental_price,omitempty"`
	LateFee            float64    `gorm:"type:decimal(10,2)" json:"late_fee,omitempty"`
	DamageFee          float64    `gorm:"type:decimal(10,2)" json:"damage_fee,omitempty"`
	TotalAmount        float64    `gorm:"-" json:"total_amount,omitempty"`
	PaymentStatus      string     `gorm:"size:50;not null;default:unpaid;check:payment_status IN ('unpaid', 'pending', 'paid', 'expired', 'failed', 'refunded', 'partially_paid', 'extension')" json:"payment_status,omitempty"`
	Notes              string     `gorm:"type:text" json:"notes,omitempty"`

	User        User         `gorm:"foreignKey:UserID" json:"-" swaggerignore:"true"`
	RentalItems []RentalItem `gorm:"foreignKey:RentalID" json:"rental_items,omitempty"`
	Payments    []Payment    `gorm:"foreignKey:RentalID" json:"payments,omitempty" swaggerignore:"true"`
}

func (*Rental) TableName() string {
	return "rentals"
}

func (r *Rental) BeforeCreate(tx *gorm.DB) error {
	if err := r.BaseEntity.BeforeCreate(tx); err != nil {
		return err
	}

	if !r.ExpectedReturnDate.After(r.RentalDate) {
		return ErrInvalidReturnDate
	}

	if r.ActualReturnDate != nil && r.ActualReturnDate.Before(r.RentalDate) {
		return ErrInvalidActualReturnDate
	}

	return nil
}

//func (r *Rental) Validate() []string {
//	validateExpectedReturnDate := func(value interface{}) error {
//		date, _ := value.(time.Time)
//		if date.IsZero() {
//			return nil
//		}
//
//		if !date.After(r.RentalDate) {
//			return errors.New("Tanggal pengembalian harus setelah tanggal rental")
//		}
//		return nil
//	}
//
//	isActualDateRequired := r.Status == RentalStatusActive
//	validateActualReturnDate := func(value interface{}) error {
//		date, ok := value.(*time.Time)
//		if !ok || date == nil {
//			return nil
//		}
//
//		if date.Before(r.RentalDate) {
//			return errors.New("Tanggal pengembalian aktual tidak boleh sebelum tanggal rental")
//		}
//		return nil
//	}
//
//	isLateFeeRequired := r.Status == RentalStatusOverdue
//
//	isDamageFeeRequired := false
//	for _, item := range r.RentalItems {
//		if item.Status == RentalItemStatusDamaged || item.ConditionAfter == "damaged" {
//			isDamageFeeRequired = true
//			break
//		}
//	}
//
//	err := validation.ValidateStruct(r,
//		validation.Field(&r.Status,
//			validation.Required.Error("Status rental wajib diisi"),
//			validation.In(RentalStatusPending, RentalStatusActive, RentalStatusCompleted, RentalStatusOverdue, RentalStatusCancelled).
//				Error("Status harus salah satu dari: pending, active, completed, overdue, atau cancelled"),
//		),
//		validation.Field(&r.RentalDate,
//			validation.Required.Error("Tanggal rental wajib diisi"),
//			validation.By(func(value interface{}) error {
//				date, _ := value.(time.Time)
//				if date.IsZero() {
//					return errors.New("Tanggal rental wajib diisi")
//				}
//				return nil
//			}),
//		),
//		validation.Field(&r.ExpectedReturnDate,
//			validation.Required.Error("Tanggal pengembalian yang diharapkan wajib diisi"),
//			validation.By(validateExpectedReturnDate),
//		),
//		validation.Field(&r.ActualReturnDate,
//			validation.When(isActualDateRequired, validation.By(validateActualReturnDate),
//				validation.When(r.Status == RentalStatusCompleted, validation.Required.Error("Tanggal pengembalian aktual wajib diisi untuk rental yang sudah selesai"))),
//		),
//		//validation.Field(&r.TotalRentalPrice,
//		//	validation.Required.Error("Total harga rental wajib diisi"),
//		//	validation.Min(0.0).Error("Total harga rental tidak boleh negatif"),
//		//),
//		validation.Field(&r.LateFee,
//			validation.When(isLateFeeRequired, validation.Required.Error("Biaya keterlambatan wajib diisi untuk rental yang terlambat")),
//			validation.Min(0.0).Error("Biaya keterlambatan tidak boleh negatif"),
//		),
//		validation.Field(&r.DamageFee,
//			validation.When(isDamageFeeRequired, validation.Required.Error("Biaya kerusakan wajib diisi karena ada item yang rusak")),
//			validation.Min(0.0).Error("Biaya kerusakan tidak boleh negatif"),
//		),
//	)
//
//	if err == nil {
//		return nil
//	}
//
//	var errorMessages []string
//	if validationErrors, ok := err.(validation.Errors); ok {
//		for _, fieldErr := range validationErrors {
//			errorMessages = append(errorMessages, fieldErr.Error())
//		}
//	} else {
//		errorMessages = append(errorMessages, err.Error())
//	}
//
//	return errorMessages
//}

type CreateRentalRequest struct {
	UserID             uuid.UUID                 `json:"user_id"`
	RentalDate         time.Time                 `json:"rental_date"`
	ExpectedReturnDate time.Time                 `json:"expected_return_date"`
	Items              []CreateRentalItemRequest `json:"items"`
	Notes              string                    `json:"notes"`
}

type CreateRentalItemRequest struct {
	ToyID           uuid.UUID `json:"toy_id"`
	Quantity        int       `json:"quantity"`
	ConditionBefore string    `json:"condition_before"`
}

type ReturnRentalRequest struct {
	ActualReturnDate time.Time                 `json:"actual_return_date" binding:"required"`
	Items            []ReturnRentalItemRequest `json:"items" binding:"required"`
	Notes            string                    `json:"notes"`
}

type ReturnRentalItemRequest struct {
	RentalItemID      uuid.UUID `json:"rental_item_id" binding:"required"`
	ConditionAfter    string    `json:"condition_after" binding:"required"`
	DamageDescription string    `json:"damage_description"`
}

type ExtendRentalRequest struct {
	NewExpectedReturnDate time.Time `json:"new_expected_return_date" binding:"required"`
	Notes                 string    `json:"notes"`
}

type ExtensionMetadata struct {
	OldExpectedReturnDate time.Time `json:"old_expected_return_date"`
	NewExpectedReturnDate time.Time `json:"new_expected_return_date"`
	AdditionalDays        int       `json:"additional_days"`
	OriginalRentalPrice   float64   `json:"original_rental_price"`
	AdditionalCost        float64   `json:"additional_cost"`
}
