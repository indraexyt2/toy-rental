package service

import (
	"final-project/entity"
	"final-project/repository"
)

type IToyCategoryService interface {
	IBaseService[entity.ToyCategory]
}

type ToyCategoryService struct {
	BaseService[entity.ToyCategory]
}

func NewToyCategoryService(repo repository.IToyCategoryRepository) IToyCategoryService {
	return &ToyCategoryService{
		BaseService: BaseService[entity.ToyCategory]{repository: repo},
	}
}
