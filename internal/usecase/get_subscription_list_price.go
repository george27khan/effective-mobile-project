package usecase

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	"effective-mobile-project/internal/transport/http/server"
	"log/slog"
)

var _ server.GetSubsListPriceUseCase = (*GetSubsListPriceUseCase)(nil)

type GetSubsListPriceRepository interface {
	GetListPrice(ctx context.Context, userId entity.UserIdNil, serviceName entity.ServiceNameNil) (subsListPrice entity.Price, err error)
}

type GetSubsListPriceUseCase struct {
	Repository GetSubsListPriceRepository
	logger     *slog.Logger
}

func NewGetSubsListPriceUseCase(repo GetSubsListPriceRepository, logger *slog.Logger) *GetSubsListPriceUseCase {
	return &GetSubsListPriceUseCase{
		Repository: repo,
		logger:     logger,
	}
}

// GetSubsListPrice получить список подписок
func (s *GetSubsListPriceUseCase) GetSubsListPrice(ctx context.Context, userId entity.UserIdNil, serviceName entity.ServiceNameNil) (subsListPrice entity.Price, err error) {
	subsListPrice, err = s.Repository.GetListPrice(ctx, userId, serviceName)
	if err != nil {
		return nil, err
	}
	return subsListPrice, nil
}
