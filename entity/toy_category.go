package entity

import validation "github.com/go-ozzo/ozzo-validation/v4"

type ToyCategory struct {
	BaseEntity
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`

	Toys []Toy `gorm:"many2many:toy_toy_categories" json:"-"`
}

func (*ToyCategory) TableName() string {
	return "toy_categories"
}

func (c *ToyCategory) Validate() []string {
	err := validation.ValidateStruct(c,
		validation.Field(&c.Name,
			validation.Required.Error("Nama kategori wajib diisi"),
			validation.RuneLength(3, 25).Error("Nama kategori harus antara 3-25 karakter"),
		),
		validation.Field(&c.Description,
			validation.When(c.Description != "", validation.RuneLength(10, 5000).Error("Deskripsi harus antara 10-5000 karakter")),
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
