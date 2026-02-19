package usecase

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	"effective-mobile-project/internal/transport/http/server"
	"log/slog"
)

var _ server.GetSubsUseCase = (*GetSubsUseCase)(nil)

type GetSubsRepository interface {
	Get(ctx context.Context, id entity.IdSubs) (subs entity.Subscription, err error)
}

type GetSubsUseCase struct {
	Repository GetSubsRepository
	logger     *slog.Logger
}

func NewGetSubsUseCase(repo GetSubsRepository, logger *slog.Logger) *GetSubsUseCase {
	return &GetSubsUseCase{
		Repository: repo,
		logger:     logger,
	}
}

// GetSubs получить подписку
func (s *GetSubsUseCase) GetSubs(ctx context.Context, id entity.IdSubs) (entity.Subscription, error) {
	subs, err := s.Repository.Get(ctx, id)
	if err != nil {
		return entity.Subscription{}, err
	}
	return subs, nil
}
