package repository

import (
	"context"
	"gorm.io/gorm"
)

type IBaseRepository[T any] interface {
	FindAll(ctx context.Context, limit int, offset int) ([]T, int64, error)
	FindById(ctx context.Context, id string) (T, error)
	Insert(ctx context.Context, entity *T) error
	UpdateById(ctx context.Context, id string, entity *T) error
	DeleteById(ctx context.Context, id string) error
}

type BaseRepository[T any] struct {
	DB *gorm.DB
}

func (r *BaseRepository[T]) FindAll(ctx context.Context, limit int, offset int) ([]T, int64, error) {
	var entities []T
	if err := r.DB.WithContext(ctx).Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	var totalData int64
	if err := r.DB.WithContext(ctx).Model(new(T)).Count(&totalData).Error; err != nil {
		return nil, 0, err
	}
	return entities, totalData, nil
}

func (r *BaseRepository[T]) FindById(ctx context.Context, id string) (T, error) {
	var entity T
	if err := r.DB.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		return entity, err
	}
	return entity, nil
}

func (r *BaseRepository[T]) Insert(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Create(&entity).Error
}

func (r *BaseRepository[T]) UpdateById(ctx context.Context, id string, entity *T) error {
	return r.DB.WithContext(ctx).Model(entity).Where("id = ?", id).Updates(entity).Error
}

func (r *BaseRepository[T]) DeleteById(ctx context.Context, id string) error {
	return r.DB.WithContext(ctx).Delete(new(T), "id = ?", id).Error
}
