package repository

import (
	"context"
	"final-project/entity"
	"gorm.io/gorm"
)

type IToyRepository interface {
	IBaseRepository[entity.Toy]
}

type ToyRepository struct {
	BaseRepository[entity.Toy]
}

func NewToyRepository(db *gorm.DB) IToyRepository {
	return &ToyRepository{
		BaseRepository: BaseRepository[entity.Toy]{DB: db},
	}
}

func (r *ToyRepository) Insert(ctx context.Context, toy *entity.Toy) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		toyWithoutRelations := &entity.Toy{
			Name:              toy.Name,
			Description:       toy.Description,
			AgeRecommendation: toy.AgeRecommendation,
			Condition:         toy.Condition,
			RentalPrice:       toy.RentalPrice,
			LateFeePerDay:     toy.LateFeePerDay,
			ReplacementPrice:  toy.ReplacementPrice,
			IsAvailable:       toy.IsAvailable,
			Stock:             toy.Stock,
			PrimaryImage:      toy.PrimaryImage,
		}

		if err := tx.Create(toyWithoutRelations).Error; err != nil {
			return err
		}

		for _, category := range toy.Categories {
			if err := tx.Exec("INSERT INTO toy_toy_categories (toy_id, toy_category_id) VALUES (?, ?)",
				toyWithoutRelations.ID, category.ID).Error; err != nil {
				return err
			}
		}

		for _, image := range toy.Images {
			if err := tx.Exec("INSERT INTO toy_toy_images (toy_id, toy_image_id) VALUES (?, ?)",
				toyWithoutRelations.ID, image.ID).Error; err != nil {
				return err
			}
		}

		toy.ID = toyWithoutRelations.ID

		return nil
	})
}

func (r *ToyRepository) UpdateById(ctx context.Context, id string, toy *entity.Toy) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Toy{}).Where("id = ?", id).Updates(map[string]interface{}{
			"name":               toy.Name,
			"description":        toy.Description,
			"age_recommendation": toy.AgeRecommendation,
			"condition":          toy.Condition,
			"rental_price":       toy.RentalPrice,
			"late_fee_per_day":   toy.LateFeePerDay,
			"replacement_price":  toy.ReplacementPrice,
			"is_available":       toy.IsAvailable,
			"stock":              toy.Stock,
			"primary_image":      toy.PrimaryImage,
		}).Error; err != nil {
			return err
		}

		if err := tx.Exec("DELETE FROM toy_toy_categories WHERE toy_id = ?", id).Error; err != nil {
			return err
		}

		for _, category := range toy.Categories {
			if err := tx.Exec("INSERT INTO toy_toy_categories (toy_id, toy_category_id) VALUES (?, ?)",
				id, category.ID).Error; err != nil {
				return err
			}
		}

		if err := tx.Exec("DELETE FROM toy_toy_images WHERE toy_id = ?", id).Error; err != nil {
			return err
		}

		for _, image := range toy.Images {
			if err := tx.Exec("INSERT INTO toy_toy_images (toy_id, toy_image_id) VALUES (?, ?)",
				id, image.ID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *ToyRepository) FindAll(ctx context.Context, limit int, offset int) ([]entity.Toy, int64, error) {
	var entities []entity.Toy
	if err := r.DB.WithContext(ctx).
		Preload("Categories").
		Preload("Images").
		Limit(limit).Offset(offset).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	var totalData int64
	if err := r.DB.WithContext(ctx).Model(new(entity.Toy)).Count(&totalData).Error; err != nil {
		return nil, 0, err
	}
	return entities, totalData, nil
}

func (r *ToyRepository) FindById(ctx context.Context, id string) (entity.Toy, error) {
	var toy entity.Toy
	if err := r.DB.WithContext(ctx).
		Preload("Categories").
		Preload("Images").
		Where("id = ?", id).
		First(&toy).Error; err != nil {
		return toy, err
	}
	return toy, nil
}
