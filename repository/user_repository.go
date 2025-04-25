package repository

import (
	"context"
	"final-project/entity"
	"gorm.io/gorm"
)

type IUserRepository interface {
	IBaseRepository[entity.User]
	FindByEmailOrUsername(ctx context.Context, email string) (*entity.User, error)
}

type UserRepository struct {
	*BaseRepository[entity.User]
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{
		BaseRepository: &BaseRepository[entity.User]{DB: db},
	}
}

func (r *UserRepository) FindAll(ctx context.Context, limit int, offset int) ([]entity.User, int64, error) {
	var entities []entity.User
	if err := r.DB.WithContext(ctx).Omit("password").Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	var totalData int64
	if err := r.DB.WithContext(ctx).Model(entity.User{}).Count(&totalData).Error; err != nil {
		return nil, 0, err
	}
	return entities, totalData, nil
}

func (r *UserRepository) FindByEmailOrUsername(ctx context.Context, emailOrUsername string) (*entity.User, error) {
	var user entity.User
	if err := r.DB.WithContext(ctx).Where("email = ? OR username = ?", emailOrUsername, emailOrUsername).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
