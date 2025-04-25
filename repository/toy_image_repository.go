package repository

import (
	"final-project/entity"
	"gorm.io/gorm"
)

type IToyImageRepository interface {
	IBaseRepository[entity.ToyImage]
}

type ToyImageRepository struct {
	BaseRepository[entity.ToyImage]
}

func NewToyImageRepository(db *gorm.DB) IToyImageRepository {
	return &ToyImageRepository{
		BaseRepository: BaseRepository[entity.ToyImage]{DB: db},
	}
}
