package entity

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofrs/uuid/v5"
)

const (
	RentalItemStatusRented   = "rented"
	RentalItemStatusReturned = "returned"
	RentalItemStatusDamaged  = "damaged"
	RentalItemStatusLost     = "lost"
)

type RentalItem struct {
	BaseEntity
	RentalID          uuid.UUID `gorm:"type:uuid;not null" json:"rental_id"`
	ToyID             uuid.UUID `gorm:"type:uuid;not null" json:"toy_id"`
	Quantity          int       `gorm:"not null;default:1" json:"quantity"`
	PricePerUnit      float64   `gorm:"type:decimal(10,2);not null" json:"price_per_unit"`
	ConditionBefore   string    `gorm:"size:50;not null;check:condition_before IN ('new', 'excellent', 'good', 'fair', 'poor')" json:"condition_before"`
	ConditionAfter    string    `gorm:"size:50;check:condition_after IN ('new', 'excellent', 'good', 'fair', 'poor', 'damaged', 'lost')" json:"condition_after"`
	DamageDescription string    `gorm:"type:text" json:"damage_description"`
	DamageFee         float64   `gorm:"type:decimal(10,2)" json:"damage_fee"`
	Status            string    `gorm:"size:50;not null;default:rented;check:status IN ('rented', 'returned', 'damaged', 'lost')" json:"status"`

	Rental Rental `gorm:"foreignKey:RentalID" json:"-"`
	Toy    Toy    `gorm:"foreignKey:ToyID" json:"toy"`
}

func (*RentalItem) TableName() string {
	return "rental_items"
}

func (ri *RentalItem) Validate() []string {
	isDamageDescriptionRequired := ri.Status == RentalItemStatusDamaged || ri.ConditionAfter == "damaged"
	isDamageFeeRequired := ri.Status == RentalItemStatusDamaged || ri.ConditionAfter == "damaged"

	err := validation.ValidateStruct(ri,
		validation.Field(&ri.RentalID,
			validation.Required.Error("ID rental wajib diisi"),
			validation.NotNil.Error("ID rental tidak boleh kosong"),
		),
		validation.Field(&ri.ToyID,
			validation.Required.Error("ID mainan wajib diisi"),
			validation.NotNil.Error("ID mainan tidak boleh kosong"),
		),
		validation.Field(&ri.Quantity,
			validation.Required.Error("Jumlah wajib diisi"),
			validation.Min(1).Error("Jumlah minimal 1"),
		),
		validation.Field(&ri.PricePerUnit,
			validation.Required.Error("Harga per unit wajib diisi"),
			validation.Min(0.0).Error("Harga per unit tidak boleh negatif"),
		),
		validation.Field(&ri.ConditionBefore,
			validation.Required.Error("Kondisi awal mainan wajib diisi"),
			validation.In(ConditionNew, ConditionExcellent, ConditionGood, ConditionFair, ConditionPoor).
				Error("Kondisi awal harus salah satu dari: new, excellent, good, fair, atau poor"),
		),
		validation.Field(&ri.ConditionAfter,
			validation.When(ri.Status != RentalItemStatusRented, validation.Required.Error("Kondisi akhir mainan wajib diisi")),
			validation.When(ri.ConditionAfter != "", validation.In(ConditionNew, ConditionExcellent, ConditionGood, ConditionFair, ConditionPoor, "damaged", "lost").
				Error("Kondisi akhir harus salah satu dari: new, excellent, good, fair, atau poor")),
		),
		validation.Field(&ri.DamageDescription,
			validation.When(isDamageDescriptionRequired, validation.Required.Error("Deskripsi kerusakan wajib diisi ketika mainan rusak")),
			validation.When(ri.DamageDescription != "", validation.RuneLength(10, 1000).Error("Deskripsi kerusakan harus antara 10-1000 karakter")),
		),
		validation.Field(&ri.DamageFee,
			validation.When(isDamageFeeRequired, validation.Required.Error("Biaya kerusakan wajib diisi ketika mainan rusak")),
			validation.When(ri.DamageFee > 0, validation.Min(0.0).Error("Biaya kerusakan tidak boleh negatif")),
		),
		validation.Field(&ri.Status,
			validation.Required.Error("Status wajib diisi"),
			validation.In(RentalItemStatusRented, RentalItemStatusReturned, RentalItemStatusDamaged, RentalItemStatusLost).
				Error("Status harus salah satu dari: rented, returned, damaged, atau lost"),
		),
	)

	if err == nil {
		return nil
	}

	var errorMessages []string
	if validationErrors, ok := err.(validation.Errors); ok {
		for _, fieldErr := range validationErrors {
			errorMessages = append(errorMessages, fieldErr.Error())
		}
	} else {
		errorMessages = append(errorMessages, err.Error())
	}

	return errorMessages
}
