package service

import (
	"context"
	"final-project/repository"
)

type IBaseService[T any] interface {
	FindAll(ctx context.Context, limit int, offset int) ([]T, int64, error)
	FindById(ctx context.Context, id string) (T, error)
	Insert(ctx context.Context, entity *T) error
	UpdateById(ctx context.Context, id string, entity *T) error
	DeleteById(ctx context.Context, id string) error
}

type BaseService[T any] struct {
	repository repository.IBaseRepository[T]
}

func (s *BaseService[T]) FindAll(ctx context.Context, limit int, offset int) ([]T, int64, error) {
	return s.repository.FindAll(ctx, limit, offset)
}

func (s *BaseService[T]) FindById(ctx context.Context, id string) (T, error) {
	return s.repository.FindById(ctx, id)
}

func (s *BaseService[T]) Insert(ctx context.Context, entity *T) error {
	return s.repository.Insert(ctx, entity)
}

func (s *BaseService[T]) UpdateById(ctx context.Context, id string, entity *T) error {
	return s.repository.UpdateById(ctx, id, entity)
}

func (s *BaseService[T]) DeleteById(ctx context.Context, id string) error {
	return s.repository.DeleteById(ctx, id)
}
