package usecase

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	"effective-mobile-project/internal/transport/http/server"
	"log/slog"
)

//go:generate mockgen -package task -source=get_task.go -destination=mock_get_task.go

var _ server.CancelSubsUseCase = (*CancelSubsUseCase)(nil)

type CancelSubsRepository interface {
	Get(ctx context.Context, id entity.IdSubs) (subs entity.Subscription, err error)
	Update(ctx context.Context, subs entity.Subscription) (err error)
}

type CancelSubsUseCase struct {
	Repository CancelSubsRepository
	logger     *slog.Logger
}

func NewCancelTaskUseCase(repo CancelSubsRepository, logger *slog.Logger) *CancelSubsUseCase {
	return &CancelSubsUseCase{
		Repository: repo,
		logger:     logger,
	}
}

// CancelSubs удалить подписку
func (s *CancelSubsUseCase) CancelSubs(ctx context.Context, id entity.IdSubs) error {
	subs, err := s.Repository.Get(ctx, id)
	if err != nil {
		return err
	}
	// для идемпотентности
	if subs.IsDelete {
		return nil
	}
	subs.IsDelete = true
	if err = s.Repository.Update(ctx, subs); err != nil {
		return err
	}
	return nil
}
