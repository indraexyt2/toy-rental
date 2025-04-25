package entity

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	_ "github.com/gofrs/uuid/v5"
	"regexp"
)

const (
	RoleAdmin    = "admin"
	RoleCustomer = "customer"
)

type User struct {
	BaseEntity
	Email       string `gorm:"size:255;not null" json:"email"`
	Username    string `gorm:"size:100;not null" json:"username"`
	Password    string `gorm:"type:text;not null" json:"password,omitempty"`
	FullName    string `gorm:"size:255;not null" json:"full_name"`
	PhoneNumber string `gorm:"size:20" json:"phone_number"`
	Address     string `gorm:"type:text" json:"address"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	Role        string `gorm:"size:20;not null;default:customer;check:role IN ('admin', 'customer')" json:"role"`

	Rentals    []Rental    `gorm:"foreignKey:UserID" json:"-"`
	UserTokens []UserToken `gorm:"foreignKey:UserID" json:"-"`
}

func (*User) TableName() string {
	return "users"
}

func (u *User) Validate(passValidate bool) []string {
	var (
		hasUppercase = regexp.MustCompile(`[A-Z]`)
		hasSymbol    = regexp.MustCompile(`[!@#~$%^&*()+|_{}:<>?,./;'[\]\\=\-]`)
		hasNumber    = regexp.MustCompile(`[0-9]`)
	)

	err := validation.ValidateStruct(u,
		validation.Field(&u.Email,
			validation.Required.Error("Email wajib diisi"),
			validation.RuneLength(5, 255).Error("Email harus antara 5-255 karakter"),
			is.Email.Error("Format email tidak valid"),
		),
		validation.Field(&u.Username,
			validation.Required.Error("Username wajib diisi"),
			validation.RuneLength(3, 100).Error("Username harus antara 3-100 karakter"),
		),
		validation.Field(&u.Password,
			validation.When(passValidate,
				validation.Required.Error("Password wajib diisi"),
				validation.RuneLength(8, 255).Error("Password harus antara 8-255 karakter"),
				validation.Match(hasUppercase).Error("Password harus mengandung huruf kapital"),
				validation.Match(hasSymbol).Error("Password harus mengandung simbol (misal @, #, !, dll)"),
				validation.Match(hasNumber).Error("Password harus mengandung angka"),
			),
		),
		validation.Field(&u.FullName,
			validation.Required.Error("Nama lengkap wajib diisi"),
			validation.RuneLength(2, 255).Error("Nama lengkap harus antara 2-255 karakter"),
		),
		validation.Field(&u.PhoneNumber,
			validation.RuneLength(7, 20).Error("Nomor telepon harus antara 7-20 karakter"),
			validation.Match(regexp.MustCompile(`^[0-9+\-\s]*$`)).Error("Nomor telepon hanya boleh berisi angka, +, - dan spasi"),
		),
		validation.Field(&u.Address,
			validation.When(u.Address != "", validation.RuneLength(5, 1000).Error("Alamat terlalu panjang, maksimal 1000 karakter")),
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

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
