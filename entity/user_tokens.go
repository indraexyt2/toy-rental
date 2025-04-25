package entity

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type UserToken struct {
	BaseEntity
	UserID                uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	AccessToken           string    `gorm:"type:text;not null" json:"access_token"`
	RefreshToken          string    `gorm:"type:text;not null" json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `gorm:"not null" json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `gorm:"not null" json:"refresh_token_expires_at"`
	IsBlocked             bool      `gorm:"default:false" json:"is_blocked"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (*UserToken) TableName() string {
	return "user_tokens"
}

func (t *UserToken) IsAccessTokenExpired() bool {
	return time.Now().After(t.AccessTokenExpiresAt)
}

func (t *UserToken) IsRefreshTokenExpired() bool {
	return time.Now().After(t.RefreshTokenExpiresAt)
}
