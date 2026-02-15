package usecase

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	"effective-mobile-project/internal/transport/http/server"
	"fmt"
)

//go:generate mockgen -package task -source=get_task.go -destination=mock_get_task.go

var _ server.UpdateSubsUseCase = (*UpdateSubsUseCase)(nil)

type UpdateSubsRepository interface {
	Get(ctx context.Context, id entity.IdSubs) (subs entity.Subscription, err error)
	Update(ctx context.Context, subs entity.Subscription) (err error)
}

type UpdateSubsUseCase struct {
	Repository UpdateSubsRepository
}

func NewUpdateSubsUseCase(repo UpdateSubsRepository) *UpdateSubsUseCase {
	return &UpdateSubsUseCase{
		Repository: repo,
	}
}

// UpdateSubs изменить подписку
func (s *UpdateSubsUseCase) UpdateSubs(ctx context.Context, subsChange entity.Subscription) (subs entity.Subscription, err error) {
	subs, err = s.Repository.Get(ctx, subsChange.Id)
	if err != nil {
		return entity.Subscription{}, fmt.Errorf("TaskService.GetTask: %w", err)
	}

	if subsChange.Price != nil {
		subs.Price = subsChange.Price
	}

	if subsChange.EndDate != nil {
		subs.EndDate = subsChange.EndDate
	}
	if err = s.Repository.Update(ctx, subs); err != nil {
		return entity.Subscription{}, err
	}
	return
}
