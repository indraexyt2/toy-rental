package service

import (
	"context"
	"final-project/entity"
	"final-project/repository"
	"final-project/utils/helpers"
	"os"
)

type IToyImageService interface {
	IBaseService[entity.ToyImage]
}

type ToyImageService struct {
	BaseService[entity.ToyImage]
}

func NewToyImageService(repo repository.IToyImageRepository) IToyImageService {
	return &ToyImageService{
		BaseService: BaseService[entity.ToyImage]{repository: repo},
	}
}

func (s *ToyImageService) DeleteById(ctx context.Context, id string) error {
	var logger = helpers.Logger

	image, err := s.repository.FindById(ctx, id)
	if err != nil {
		return err
	}

	if image.ImageURL != "" {
		if err := os.Remove(image.ImageURL); err != nil {
			logger.Warn("Gagal menghapus file fisik: ", err)
		} else {
			logger.Info("File berhasil dihapus: ", image.ImageURL)
		}
	}

	return s.repository.DeleteById(ctx, id)
}
