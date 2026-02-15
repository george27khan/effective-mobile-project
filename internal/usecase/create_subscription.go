package usecase

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	"fmt"
)

//go:generate mockgen -package task -source=create_subscription.go -destination=mock_create_subscription.go
type CreateSubsRepository interface {
	Create(ctx context.Context, subs entity.Subscription) (id entity.IdSubs, err error)
}

type CreateSubsUseCase struct {
	Repository CreateSubsRepository
}

func NewCreateSubsUseCase(repository CreateSubsRepository) *CreateSubsUseCase {
	return &CreateSubsUseCase{
		Repository: repository,
	}
}

// CreateSubs функция создания подписки
func (cs *CreateSubsUseCase) CreateSubs(ctx context.Context, subs entity.Subscription) (idSubs entity.IdSubs, err error) {
	idSubs, err = cs.Repository.Create(ctx, subs) //создаем запись в бд
	if err != nil {
		return 0, fmt.Errorf("TaskService.CreateTask: %w", err)
	}

	subs.Id = idSubs
	return subs.Id, nil
}
