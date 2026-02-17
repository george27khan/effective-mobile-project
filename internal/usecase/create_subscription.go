package usecase

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	"effective-mobile-project/internal/transport/http/server"
	"log/slog"
)

//go:generate mockgen -package task -source=create_subscription.go -destination=mock_create_subscription.go

var _ server.CreateSubsUseCase = (*CreateSubsUseCase)(nil)

type CreateSubsRepository interface {
	Create(ctx context.Context, subs entity.Subscription) (id entity.IdSubs, err error)
}

type CreateSubsUseCase struct {
	Repository CreateSubsRepository
	logger     *slog.Logger
}

func NewCreateSubsUseCase(repository CreateSubsRepository, logger *slog.Logger) *CreateSubsUseCase {
	return &CreateSubsUseCase{
		Repository: repository,
		logger:     logger,
	}
}

// CreateSubs функция создания подписки
func (cs *CreateSubsUseCase) CreateSubs(ctx context.Context, subs entity.Subscription) (idSubs entity.IdSubs, err error) {
	idSubs, err = cs.Repository.Create(ctx, subs) //создаем запись в бд
	if err != nil {
		return 0, err
	}
	subs.Id = idSubs
	return subs.Id, nil
}
