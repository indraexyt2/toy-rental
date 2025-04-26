package service

import (
	"context"
	"errors"
	"final-project/entity"
	"final-project/repository"
	"github.com/gofrs/uuid/v5"
)

type IToyService interface {
	IBaseService[entity.Toy]
	CreateToy(ctx context.Context, toyRequest entity.ToyRequest) (*entity.Toy, error)
	UpdateToy(ctx context.Context, id string, toyRequest entity.ToyUpdateRequest) (*entity.Toy, error)
}

type ToyService struct {
	BaseService[entity.Toy]
	toyRepo      repository.IToyRepository
	toyImageRepo repository.IToyImageRepository
	categoryRepo repository.IToyCategoryRepository
}

func NewToyService(
	repo repository.IToyRepository,
	imageRepo repository.IToyImageRepository,
	categoryRepo repository.IToyCategoryRepository,
) IToyService {
	return &ToyService{
		BaseService:  BaseService[entity.Toy]{repository: repo},
		toyRepo:      repo,
		toyImageRepo: imageRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *ToyService) CreateToy(ctx context.Context, toyRequest entity.ToyRequest) (*entity.Toy, error) {
	if len(toyRequest.CategoryIDs) == 0 {
		return nil, errors.New("kategori wajib dipilih")
	}

	if len(toyRequest.ImageIDs) == 0 {
		return nil, errors.New("setidaknya satu gambar diperlukan")
	}

	toyCategories, err := s.prepareCategoriesFromIDs(ctx, toyRequest.CategoryIDs)
	if err != nil {
		return nil, err
	}

	toyImages, err := s.prepareImagesFromIDs(ctx, toyRequest.ImageIDs)
	if err != nil {
		return nil, err
	}

	primaryImageURL, err := s.getPrimaryImageURL(ctx, toyRequest.PrimaryImageID, toyRequest.ImageIDs, toyImages)
	if err != nil {
		return nil, err
	}

	toy := &entity.Toy{
		Name:              toyRequest.Name,
		Description:       toyRequest.Description,
		AgeRecommendation: toyRequest.AgeRecommendation,
		Condition:         toyRequest.Condition,
		RentalPrice:       toyRequest.RentalPrice,
		LateFeePerDay:     toyRequest.LateFeePerDay,
		ReplacementPrice:  toyRequest.ReplacementPrice,
		IsAvailable:       toyRequest.IsAvailable,
		Stock:             toyRequest.Stock,
		PrimaryImage:      primaryImageURL,
		Categories:        toyCategories,
		Images:            toyImages,
	}

	if errs := toy.Validate(); len(errs) > 0 {
		return nil, errors.New("validasi gagal: " + errs[0])
	}

	if err := s.repository.Insert(ctx, toy); err != nil {
		return nil, err
	}

	newToy, err := s.toyRepo.FindById(ctx, toy.ID.String())
	if err != nil {
		return nil, err
	}

	return &newToy, nil
}

func (s *ToyService) UpdateToy(ctx context.Context, id string, toyRequest entity.ToyUpdateRequest) (*entity.Toy, error) {
	existingToy, err := s.FindById(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(toyRequest.CategoryIDs) == 0 {
		return nil, errors.New("kategori wajib dipilih")
	}

	if len(toyRequest.ImageIDs) == 0 {
		return nil, errors.New("setidaknya satu gambar diperlukan")
	}

	existingToy.Name = toyRequest.Name
	existingToy.Description = toyRequest.Description
	existingToy.AgeRecommendation = toyRequest.AgeRecommendation
	existingToy.Condition = toyRequest.Condition
	existingToy.RentalPrice = toyRequest.RentalPrice
	existingToy.LateFeePerDay = toyRequest.LateFeePerDay
	existingToy.ReplacementPrice = toyRequest.ReplacementPrice
	existingToy.IsAvailable = toyRequest.IsAvailable
	existingToy.Stock = toyRequest.Stock

	toyCategories, err := s.prepareCategoriesFromIDs(ctx, toyRequest.CategoryIDs)
	if err != nil {
		return nil, err
	}
	existingToy.Categories = toyCategories

	toyImages, err := s.prepareImagesFromIDs(ctx, toyRequest.ImageIDs)
	if err != nil {
		return nil, err
	}
	existingToy.Images = toyImages

	if toyRequest.PrimaryImageID != "" || (len(toyRequest.ImageIDs) > 0 && existingToy.PrimaryImage == "") {
		primaryImageURL, err := s.getPrimaryImageURL(ctx, toyRequest.PrimaryImageID, toyRequest.ImageIDs, toyImages)
		if err != nil {
			return nil, err
		}
		existingToy.PrimaryImage = primaryImageURL
	}

	if errs := existingToy.Validate(); len(errs) > 0 {
		return nil, errors.New("validasi gagal: " + errs[0])
	}

	if err := s.repository.UpdateById(ctx, id, &existingToy); err != nil {
		return nil, err
	}

	return &existingToy, nil
}

func (s *ToyService) prepareCategoriesFromIDs(ctx context.Context, categoryIDs []string) ([]entity.ToyCategory, error) {
	toyCategories := make([]entity.ToyCategory, 0, len(categoryIDs))

	for _, categoryIDStr := range categoryIDs {
		categoryID, err := uuid.FromString(categoryIDStr)
		if err != nil {
			return nil, errors.New("format ID kategori tidak valid")
		}

		_, err = s.categoryRepo.FindById(ctx, categoryID.String())
		if err != nil {
			return nil, errors.New("kategori dengan ID " + categoryIDStr + " tidak ditemukan")
		}

		toyCategories = append(toyCategories, entity.ToyCategory{
			BaseEntity: entity.BaseEntity{
				ID: categoryID,
			},
		})
	}

	return toyCategories, nil
}

func (s *ToyService) prepareImagesFromIDs(ctx context.Context, imageIDs []string) ([]entity.ToyImage, error) {
	toyImages := make([]entity.ToyImage, 0, len(imageIDs))

	for _, imageIDStr := range imageIDs {
		imageID, err := uuid.FromString(imageIDStr)
		if err != nil {
			return nil, errors.New("format ID gambar tidak valid")
		}

		_, err = s.toyImageRepo.FindById(ctx, imageID.String())
		if err != nil {
			return nil, errors.New("gambar dengan ID " + imageIDStr + " tidak ditemukan")
		}

		toyImages = append(toyImages, entity.ToyImage{
			BaseEntity: entity.BaseEntity{
				ID: imageID,
			},
		})
	}

	return toyImages, nil
}

func (s *ToyService) getPrimaryImageURL(ctx context.Context, primaryImageID string, imageIDs []string, toyImages []entity.ToyImage) (string, error) {
	if primaryImageID != "" {
		primaryID, err := uuid.FromString(primaryImageID)
		if err != nil {
			return "", errors.New("format ID gambar utama tidak valid")
		}

		validPrimaryImage := false
		for _, img := range toyImages {
			if img.ID == primaryID {
				validPrimaryImage = true
				break
			}
		}

		if !validPrimaryImage {
			return "", errors.New("ID gambar utama tidak ditemukan dalam daftar gambar yang dipilih")
		}

		primaryImage, err := s.toyImageRepo.FindById(ctx, primaryID.String())
		if err != nil {
			return "", errors.New("gagal mendapatkan gambar utama")
		}

		return primaryImage.ImageURL, nil
	} else if len(imageIDs) > 0 {
		firstImageID, err := uuid.FromString(imageIDs[0])
		if err != nil {
			return "", errors.New("format ID gambar tidak valid")
		}

		firstImage, err := s.toyImageRepo.FindById(ctx, firstImageID.String())
		if err != nil {
			return "", errors.New("gagal mendapatkan gambar pertama")
		}

		return firstImage.ImageURL, nil
	}

	return "", nil
}
