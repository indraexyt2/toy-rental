package repository

import (
	"final-project/entity"
	"gorm.io/gorm"
)

type IToyCategoryRepository interface {
	IBaseRepository[entity.ToyCategory]
}

type ToyCategoryRepository struct {
	*BaseRepository[entity.ToyCategory]
}

func NewToyCategoryRepository(db *gorm.DB) IToyCategoryRepository {
	return &ToyCategoryRepository{
		BaseRepository: &BaseRepository[entity.ToyCategory]{DB: db},
	}
}
