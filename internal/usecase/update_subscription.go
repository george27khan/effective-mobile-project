package usecase

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	"effective-mobile-project/internal/transport/http/server"
	"log/slog"
)

var _ server.UpdateSubsUseCase = (*UpdateSubsUseCase)(nil)

type UpdateSubsRepository interface {
	Get(ctx context.Context, id entity.IdSubs) (subs entity.Subscription, err error)
	Update(ctx context.Context, subs entity.Subscription) (err error)
}

type UpdateSubsUseCase struct {
	Repository UpdateSubsRepository
	logger     *slog.Logger
}

func NewUpdateSubsUseCase(repo UpdateSubsRepository, logger *slog.Logger) *UpdateSubsUseCase {
	return &UpdateSubsUseCase{
		Repository: repo,
		logger:     logger,
	}
}

// UpdateSubs изменить подписку
func (s *UpdateSubsUseCase) UpdateSubs(ctx context.Context, subsChange entity.Subscription) (subs entity.Subscription, err error) {
	subs, err = s.Repository.Get(ctx, subsChange.Id)
	if err != nil {
		return entity.Subscription{}, err
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
