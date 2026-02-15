package usecase

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	"fmt"
)

//go:generate mockgen -package task -source=get_task.go -destination=mock_get_task.go
type GetSubsRepository interface {
	Get(ctx context.Context, id entity.IdSubs) (subs entity.Subscription, err error)
}

type GetSubsUseCase struct {
	Repository GetSubsRepository
}

func NewGetTaskUseCase(repo GetSubsRepository) *GetSubsUseCase {
	return &GetSubsUseCase{
		Repository: repo,
	}
}

// GetSubs получить подписку
func (s *GetSubsUseCase) GetSubs(ctx context.Context, id entity.IdSubs) (entity.Subscription, error) {
	subs, err := s.Repository.Get(ctx, id)
	if err != nil {
		return entity.Subscription{}, fmt.Errorf("TaskService.GetTask: %w", err)
	}
	return subs, nil
}
