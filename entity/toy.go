package entity

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"regexp"
)

const (
	ConditionNew       = "new"
	ConditionExcellent = "excellent"
	ConditionGood      = "good"
	ConditionFair      = "fair"
	ConditionPoor      = "poor"
)

type Toy struct {
	BaseEntity
	Name              string  `gorm:"size:255;not null" json:"name"`
	Description       string  `gorm:"type:text" json:"description"`
	AgeRecommendation string  `gorm:"size:50" json:"age_recommendation"`
	Condition         string  `gorm:"size:50;not null;check:condition IN ('new', 'excellent', 'good', 'fair', 'poor')" json:"condition"`
	RentalPrice       float64 `gorm:"type:decimal(10,2);not null" json:"rental_price"`
	LateFeePerDay     float64 `gorm:"type:decimal(10,2);not null" json:"late_fee_per_day"`
	ReplacementPrice  float64 `gorm:"type:decimal(10,2);not null" json:"replacement_price"`
	IsAvailable       bool    `gorm:"default:true" json:"is_available"`
	Stock             int     `gorm:"not null" json:"stock"`
	PrimaryImage      string  `gorm:"type:text" json:"primary_image"`

	Categories  []ToyCategory `gorm:"many2many:toy_toy_categories" json:"categories"`
	Images      []ToyImage    `gorm:"many2many:toy_toy_images" json:"images"`
	RentalItems []RentalItem  `gorm:"foreignKey:ToyID" json:"-"`
}

func (*Toy) TableName() string {
	return "toys"
}

func (t *Toy) Validate() []string {
	err := validation.ValidateStruct(t,
		validation.Field(&t.Categories,
			validation.Required.Error("Kategori wajib diisi"),
			validation.NotNil.Error("Kategori tidak boleh kosong"),
		),
		validation.Field(&t.Name,
			validation.Required.Error("Nama mainan wajib diisi"),
			validation.RuneLength(3, 255).Error("Nama mainan harus antara 3-255 karakter"),
		),
		validation.Field(&t.Description,
			validation.When(t.Description != "", validation.RuneLength(10, 5000).Error("Deskripsi harus antara 10-5000 karakter")),
		),
		validation.Field(&t.AgeRecommendation,
			validation.When(t.AgeRecommendation != "",
				validation.Match(regexp.MustCompile(`^[0-9\-+]+$`)).Error("Format rekomendasi usia tidak valid (contoh: 3-5, 5+)")),
		),
		validation.Field(&t.Condition,
			validation.Required.Error("Kondisi mainan wajib diisi"),
			validation.In(ConditionNew, ConditionExcellent, ConditionGood, ConditionFair, ConditionPoor).
				Error("Kondisi harus salah satu dari: new, excellent, good, fair, atau poor"),
		),
		validation.Field(&t.RentalPrice,
			validation.Required.Error("Harga rental wajib diisi"),
			validation.Min(0.0).Error("Harga rental tidak boleh negatif"),
		),
		validation.Field(&t.LateFeePerDay,
			validation.Required.Error("Biaya keterlambatan per hari wajib diisi"),
			validation.Min(0.0).Error("Biaya keterlambatan tidak boleh negatif"),
		),
		validation.Field(&t.ReplacementPrice,
			validation.Required.Error("Harga penggantian wajib diisi"),
			validation.Min(0.0).Error("Harga penggantian tidak boleh negatif"),
		),
		validation.Field(&t.Stock,
			validation.Required.Error("Stok wajib diisi"),
			validation.Min(0).Error("Stok tidak boleh negatif"),
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

type ToyRequest struct {
	Name              string   `json:"name" binding:"required"`
	Description       string   `json:"description"`
	AgeRecommendation string   `json:"age_recommendation"`
	Condition         string   `json:"condition" binding:"required"`
	RentalPrice       float64  `json:"rental_price" binding:"required"`
	LateFeePerDay     float64  `json:"late_fee_per_day" binding:"required"`
	ReplacementPrice  float64  `json:"replacement_price" binding:"required"`
	IsAvailable       bool     `json:"is_available"`
	Stock             int      `json:"stock" binding:"required"`
	CategoryIDs       []string `json:"category_ids" binding:"required"`
	ImageIDs          []string `json:"image_ids" binding:"required"`
	PrimaryImageID    string   `json:"primary_image_id"`
}

type ToyUpdateRequest struct {
	Name              string   `json:"name" binding:"required"`
	Description       string   `json:"description"`
	AgeRecommendation string   `json:"age_recommendation"`
	Condition         string   `json:"condition" binding:"required"`
	RentalPrice       float64  `json:"rental_price" binding:"required"`
	LateFeePerDay     float64  `json:"late_fee_per_day" binding:"required"`
	ReplacementPrice  float64  `json:"replacement_price" binding:"required"`
	IsAvailable       bool     `json:"is_available"`
	Stock             int      `json:"stock" binding:"required"`
	CategoryIDs       []string `json:"category_ids" binding:"required"`
	ImageIDs          []string `json:"image_ids" binding:"required"`
	PrimaryImageID    string   `json:"primary_image_id"`
}
