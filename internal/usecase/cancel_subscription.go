package usecase

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	"fmt"
	"time"
)

//go:generate mockgen -package task -source=get_task.go -destination=mock_get_task.go
type CancelSubsRepository interface {
	Get(ctx context.Context, id entity.IdSubs) (subs entity.Subscription, err error)
	Update(ctx context.Context, subs entity.Subscription) (err error)
}

type CancelSubsUseCase struct {
	Repository CancelSubsRepository
}

func NewCancelTaskUseCase(repo CancelSubsRepository) *CancelSubsUseCase {
	return &CancelSubsUseCase{
		Repository: repo,
	}
}

// CancelSubs удалить подписку
func (s *CancelSubsUseCase) CancelSubs(ctx context.Context, id entity.IdSubs) error {
	subs, err := s.Repository.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("TaskService.GetTask: %w", err)
	}
	// для идемпотентности
	if subs.DeleteDate != nil {
		return nil
	}
	now := time.Now()
	subs.DeleteDate = &now
	if err = s.Repository.Update(ctx, subs); err != nil {
		return err
	}
	return nil
}
