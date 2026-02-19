package usecase

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	"effective-mobile-project/internal/transport/http/server"
	"log/slog"
)

var _ server.GetSubsListUseCase = (*GetSubsListUseCase)(nil)

type GetSubsListRepository interface {
	GetList(ctx context.Context, userId entity.UserId, limit int, offset int) (subsList []entity.Subscription, err error)
}

type GetSubsListUseCase struct {
	Repository GetSubsListRepository
	logger     *slog.Logger
}

func NewGetSubsListUseCase(repo GetSubsListRepository, logger *slog.Logger) *GetSubsListUseCase {
	return &GetSubsListUseCase{
		Repository: repo,
		logger:     logger,
	}
}

// GetSubsList получить список подписок
func (s *GetSubsListUseCase) GetSubsList(ctx context.Context, userId entity.UserId, limit int, offset int) (subsList []entity.Subscription, err error) {
	subsList, err = s.Repository.GetList(ctx, userId, limit, offset)
	if err != nil {
		return nil, err
	}
	return subsList, nil
}
