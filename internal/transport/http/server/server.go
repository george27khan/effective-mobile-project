package server

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	errRep "effective-mobile-project/internal/infrastructure/repository/errors_repo"
	"errors"
	"fmt"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"log/slog"
	"time"
)

//go:generate mockgen -package server -source=task_server.go -destination=mock_task_server.go

var _ StrictServerInterface = (*SubsServer)(nil)

type CreateSubsUseCase interface {
	CreateSubs(ctx context.Context, subs entity.Subscription) (idSubs entity.IdSubs, err error)
}

type GetSubsUseCase interface {
	GetSubs(ctx context.Context, id entity.IdSubs) (subs entity.Subscription, err error)
}

type UpdateSubsUseCase interface {
	UpdateSubs(ctx context.Context, subsChange entity.Subscription) (subs entity.Subscription, err error)
}

type CancelSubsUseCase interface {
	CancelSubs(ctx context.Context, subsId entity.IdSubs) (err error)
}

// SubsServer реализует интерфейс сервера
type SubsServer struct {
	CreateSubsUseCase CreateSubsUseCase
	GetSubsUseCase    GetSubsUseCase
	UpdateSubsUseCase UpdateSubsUseCase
	CancelSubsUseCase CancelSubsUseCase
}

func NewSubsServer(createSubs CreateSubsUseCase,
	getSubs GetSubsUseCase,
	updateSubs UpdateSubsUseCase,
	cancelSubs CancelSubsUseCase) *SubsServer {
	return &SubsServer{
		CreateSubsUseCase: createSubs,
		GetSubsUseCase:    getSubs,
		UpdateSubsUseCase: updateSubs,
		CancelSubsUseCase: cancelSubs,
	}
}

func resp400(msg string, err error) BadRequestJSONResponse {
	return BadRequestJSONResponse{
		Code:    BADREQUEST,
		Message: fmt.Errorf("%s: %w", msg, err).Error(),
	}
}

func resp404(msg string, err error) NotFoundJSONResponse {
	return NotFoundJSONResponse{
		Code:    NOTFOUND,
		Message: fmt.Errorf("%s: %w", msg, err).Error(),
	}
}

func resp500(msg string, err error) InternalServerErrorJSONResponse {
	return InternalServerErrorJSONResponse{
		Code:    INTERNALSERVERERROR,
		Message: fmt.Errorf("%s: %w", msg, err).Error(),
	}
}

func validate(body *CreateSubscriptionJSONRequestBody) error {
	return nil
}

// CreateSubscription создание подписки
func (s *SubsServer) CreateSubscription(ctx context.Context, request CreateSubscriptionRequestObject) (CreateSubscriptionResponseObject, error) {
	if err := validate(request.Body); err != nil {
		return CreateSubscription400JSONResponse{
			resp400("PostDownloads: ошибка валидации параметров", err)}, nil
	}
	subs := entity.Subscription{
		ServiceName: entity.ServiceName(request.Body.ServiceName),
		Price:       entity.Price(&request.Body.Price),
		UserId:      entity.UserId(request.Body.UserId),
		StartDate:   request.Body.StartDate.Time,
	}

	subsId, err := s.CreateSubsUseCase.CreateSubs(ctx, subs)
	if err != nil {
		slog.Error("ошибка ", err)
		return CreateSubscription500JSONResponse{
				resp500("PostDownloads", err)},
			nil
	}
	return CreateSubscription201JSONResponse{
		Id: int(subsId),
	}, nil
}

// GetSubscription получение таска
func (s *SubsServer) GetSubscription(ctx context.Context, request GetSubscriptionRequestObject) (GetSubscriptionResponseObject, error) {
	subs, err := s.GetSubsUseCase.GetSubs(ctx, entity.IdSubs(request.Id))
	if err != nil {
		if errors.Is(err, errRep.ErrSubsNotFound) {
			return GetSubscription404JSONResponse{resp404("GetSubscription", err)}, nil
		}
		return GetSubscription500JSONResponse{resp500("GetSubscription", err)}, nil
	}
	fmt.Println("subs.Id ", subs.Id)
	fmt.Println("subs.Price ", subs.Price)
	price := subs.Price
	//собираем файлы для ответа
	return GetSubscription200JSONResponse{
		Id:          int(subs.Id),
		ServiceName: string(subs.ServiceName),
		Price:       *price,
		UserId:      openapi_types.UUID(subs.UserId),
		StartDate:   openapi_types.Date{subs.StartDate},
		EndDate:     nilDate(subs.EndDate),
		DeleteDate:  nilDate(subs.DeleteDate),
	}, nil
}

// UpdateSubscription изменение подписки
func (s *SubsServer) UpdateSubscription(ctx context.Context, request UpdateSubscriptionRequestObject) (UpdateSubscriptionResponseObject, error) {

	subs := entity.Subscription{
		Id:    entity.IdSubs(request.Id),
		Price: entity.Price(request.Body.Price),
	}
	if request.Body.EndDate != nil {
		subs.EndDate = &request.Body.EndDate.Time
	}
	subs, err := s.UpdateSubsUseCase.UpdateSubs(ctx, subs)
	if err != nil {
		if errors.Is(err, errRep.ErrSubsNotFound) {
			return UpdateSubscription404JSONResponse{resp404("GetDownloadsId", err)}, nil
		}
		return UpdateSubscription500JSONResponse{resp500("GetDownloadsId", err)}, nil
	}

	//собираем файлы для ответа
	return UpdateSubscription200JSONResponse{
		Id:          int(subs.Id),
		ServiceName: string(subs.ServiceName),
		Price:       int(*subs.Price),
		UserId:      openapi_types.UUID(subs.UserId),
		StartDate:   openapi_types.Date{time.Time(subs.StartDate)},
		EndDate:     nilDate(subs.EndDate),
	}, nil
}

func (s *SubsServer) CancelSubscription(ctx context.Context, request CancelSubscriptionRequestObject) (CancelSubscriptionResponseObject, error) {
	err := s.CancelSubsUseCase.CancelSubs(ctx, entity.IdSubs(request.Id))
	if err != nil {
		if errors.Is(err, errRep.ErrSubsNotFound) {
			return CancelSubscription404JSONResponse{resp404("CancelSubscription", err)}, nil
		}
		return CancelSubscription500JSONResponse{resp500("CancelSubscription", err)}, nil
	}
	//собираем файлы для ответа
	return CancelSubscription204Response{}, nil
}

func (s *SubsServer) GetSubscriptionList(ctx context.Context, request GetSubscriptionListRequestObject) (GetSubscriptionListResponseObject, error) {
	return nil, nil
}

//

func nilDate(t *time.Time) *openapi_types.Date {
	if t == nil {
		return nil
	}
	d := openapi_types.Date{Time: *t}
	return &d
}
