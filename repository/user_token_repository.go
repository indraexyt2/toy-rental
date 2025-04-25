package repository

import (
	"context"
	"final-project/entity"
	"gorm.io/gorm"
)

type IUserTokenRepository interface {
	IBaseRepository[entity.UserToken]
	FindByAccessToken(ctx context.Context, accessToken string) (entity.UserToken, error)
	DeleteByAccessToken(ctx context.Context, accessToken string) error
	UpdateByRefreshToken(ctx context.Context, refreshToken string, entity *entity.UserToken) error
}

type UserTokenRepository struct {
	BaseRepository[entity.UserToken]
}

func NewUserTokenRepository(db *gorm.DB) IUserTokenRepository {
	return &UserTokenRepository{
		BaseRepository: BaseRepository[entity.UserToken]{DB: db},
	}
}

func (r *UserTokenRepository) FindByAccessToken(ctx context.Context, accessToken string) (entity.UserToken, error) {
	var userToken entity.UserToken
	if err := r.DB.WithContext(ctx).Where("access_token = ?", accessToken).First(&userToken).Error; err != nil {
		return userToken, err
	}
	return userToken, nil
}

func (r *UserTokenRepository) DeleteByAccessToken(ctx context.Context, accessToken string) error {
	return r.DB.WithContext(ctx).Where("access_token = ?", accessToken).Delete(&entity.UserToken{}).Error
}

func (r *UserTokenRepository) UpdateByRefreshToken(ctx context.Context, refreshToken string, entity *entity.UserToken) error {
	return r.DB.WithContext(ctx).Model(entity).Where("refresh_token = ?", refreshToken).Updates(entity).Error
}
