package server

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	errRep "effective-mobile-project/internal/infrastructure/repository/errors_repo"
	"effective-mobile-project/internal/pkg/logger"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

//go:generate mockgen -package server -source=task_server.go -destination=mock_task_server.go

const StartDateFormat = "01-2006"

var _ StrictServerInterface = (*SubsServer)(nil)

var ErrBadQueryParam = errors.New("некорректные параметры запроса")

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

type GetSubsListUseCase interface {
	GetSubsList(ctx context.Context, userId entity.UserId, limit int, offset int) (subsList []entity.Subscription, err error)
}

type GetSubsListPriceUseCase interface {
	GetSubsListPrice(ctx context.Context, userId entity.UserIdNil, serviceName entity.ServiceNameNil) (subsListPrice entity.Price, err error)
}

// SubsServer реализует интерфейс сервера
type SubsServer struct {
	CreateSubsUseCase       CreateSubsUseCase
	GetSubsUseCase          GetSubsUseCase
	UpdateSubsUseCase       UpdateSubsUseCase
	CancelSubsUseCase       CancelSubsUseCase
	GetSubsListUseCase      GetSubsListUseCase
	GetSubsListPriceUseCase GetSubsListPriceUseCase
	logger                  *slog.Logger
}

// NewSubsServer конструктор сервера
func NewSubsServer(createSubs CreateSubsUseCase,
	getSubs GetSubsUseCase,
	updateSubs UpdateSubsUseCase,
	cancelSubs CancelSubsUseCase,
	getSubsList GetSubsListUseCase,
	getSubsListPrice GetSubsListPriceUseCase,
	logger *slog.Logger) *SubsServer {
	return &SubsServer{
		CreateSubsUseCase:       createSubs,
		GetSubsUseCase:          getSubs,
		UpdateSubsUseCase:       updateSubs,
		CancelSubsUseCase:       cancelSubs,
		GetSubsListUseCase:      getSubsList,
		GetSubsListPriceUseCase: getSubsListPrice,
		logger:                  logger,
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
		Message: fmt.Errorf("%s: запись не найдена %w", msg, err).Error(),
	}
}

func resp500(msg string, err error) InternalServerErrorJSONResponse {
	return InternalServerErrorJSONResponse{
		Code:    INTERNALSERVERERROR,
		Message: fmt.Errorf("%s: непредвиденная ошибка %w", msg, err).Error(),
	}
}

// CreateSubscription создание подписки
func (s *SubsServer) CreateSubscription(ctx context.Context, request CreateSubscriptionRequestObject) (CreateSubscriptionResponseObject, error) {
	startDate, err := time.Parse(StartDateFormat, request.Body.StartDate)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return CreateSubscription400JSONResponse{
			resp400("CreateSubscription: ошибка валидации параметров", err)}, nil
	}
	subs := entity.Subscription{
		ServiceName: entity.ServiceName(request.Body.ServiceName),
		Price:       entity.Price(&request.Body.Price),
		UserId:      entity.UserId(request.Body.UserId),
		StartDate:   startDate,
	}

	subsId, err := s.CreateSubsUseCase.CreateSubs(ctx, subs)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return CreateSubscription500JSONResponse{
				resp500("CreateSubscription", err)},
			nil
	}
	return CreateSubscription201JSONResponse{
		Id: int(subsId),
	}, nil
}

// GetSubscription получение подписки
func (s *SubsServer) GetSubscription(ctx context.Context, request GetSubscriptionRequestObject) (GetSubscriptionResponseObject, error) {
	subs, err := s.GetSubsUseCase.GetSubs(ctx, entity.IdSubs(request.Id))
	if err != nil {
		if errors.Is(err, errRep.ErrSubsNotFound) {
			s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
			return GetSubscription404JSONResponse{resp404("GetSubscription:", err)}, nil
		}
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return GetSubscription500JSONResponse{resp500("GetSubscription", err)}, nil
	}
	//собираем файлы для ответа
	return GetSubscription200JSONResponse(prepareSubs(subs)), nil
}

// UpdateSubscription изменение подписки
func (s *SubsServer) UpdateSubscription(ctx context.Context, request UpdateSubscriptionRequestObject) (UpdateSubscriptionResponseObject, error) {

	subs := entity.Subscription{
		Id:    entity.IdSubs(request.Id),
		Price: entity.Price(request.Body.Price),
	}
	if request.Body.EndDate != nil {
		endDate, err := time.Parse(StartDateFormat, *request.Body.EndDate)
		if err != nil {
			s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
			return UpdateSubscription400JSONResponse{
				resp400("CreateSubscription: ошибка валидации параметров", err)}, nil
		}
		subs.EndDate = &endDate
	}
	subs, err := s.UpdateSubsUseCase.UpdateSubs(ctx, subs)
	if err != nil {
		if errors.Is(err, errRep.ErrSubsNotFound) {
			s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
			return UpdateSubscription404JSONResponse{resp404("UpdateSubscription", err)}, nil
		}
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return UpdateSubscription500JSONResponse{resp500("UpdateSubscription", err)}, nil
	}

	//собираем файлы для ответа
	//return UpdateSubscription200JSONResponse{
	//	Id:          int(subs.Id),
	//	ServiceName: string(subs.ServiceName),
	//	Price:       *subs.Price,
	//	UserId:      openapiTypes.UUID(subs.UserId),
	//	StartDate:   subs.StartDate.Format(StartDateFormat),
	//	EndDate:     nilDate(subs.EndDate),
	//}, nil
	return UpdateSubscription200JSONResponse(prepareSubs(subs)), nil
}

// CancelSubscription отмена подписки
func (s *SubsServer) CancelSubscription(ctx context.Context, request CancelSubscriptionRequestObject) (CancelSubscriptionResponseObject, error) {
	err := s.CancelSubsUseCase.CancelSubs(ctx, entity.IdSubs(request.Id))
	if err != nil {
		if errors.Is(err, errRep.ErrSubsNotFound) {
			s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
			return CancelSubscription404JSONResponse{resp404("CancelSubscription", err)}, nil
		}
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return CancelSubscription500JSONResponse{resp500("CancelSubscription", err)}, nil
	}
	return CancelSubscription204Response{}, nil
}

// GetSubscriptionList получение списка подписок
func (s *SubsServer) GetSubscriptionList(ctx context.Context, request GetSubscriptionListRequestObject) (GetSubscriptionListResponseObject, error) {
	if request.Params.Limit < 0 || *request.Params.Offset < 0 {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, ErrBadQueryParam), ErrBadQueryParam.Error())
		return GetSubscriptionList400JSONResponse{resp400("GetSubscriptionList", ErrBadQueryParam)}, nil
	}
	subsList, err := s.GetSubsListUseCase.GetSubsList(ctx, entity.UserId(request.Params.UserId), request.Params.Limit, *request.Params.Offset)
	if err != nil {
		if errors.Is(err, errRep.ErrSubsNotFound) {
			s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
			return GetSubscriptionList404JSONResponse{resp404("GetSubscriptionList", err)}, nil
		}
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return GetSubscriptionList500JSONResponse{resp500("GetSubscriptionList", err)}, nil
	}
	//собираем файлы для ответа
	return GetSubscriptionList200JSONResponse(prepareSubsList(subsList)), nil
}

// GetSubscriptionPrice получение стоимости подписок
func (s *SubsServer) GetSubscriptionPrice(ctx context.Context, request GetSubscriptionPriceRequestObject) (GetSubscriptionPriceResponseObject, error) {
	if request.Params.UserId == nil && request.Params.ServiceName == nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, ErrBadQueryParam), ErrBadQueryParam.Error())
		return GetSubscriptionPrice400JSONResponse{resp400("GetSubscriptionPrice", ErrBadQueryParam)}, nil
	}
	subsListPrice, err := s.GetSubsListPriceUseCase.GetSubsListPrice(ctx, request.Params.UserId, request.Params.ServiceName)
	if err != nil {
		if errors.Is(err, errRep.ErrSubsNotFound) {
			s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
			return GetSubscriptionPrice404JSONResponse{resp404("GetSubscriptionPrice", err)}, nil
		}
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return GetSubscriptionPrice500JSONResponse{resp500("GetSubscriptionPrice", err)}, nil
	}
	//собираем файлы для ответа
	return GetSubscriptionPrice200JSONResponse{
		Price: *subsListPrice,
	}, nil
}

// nilDate вспомогательная функция для даты в структуре
func nilDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	d := t.Format(StartDateFormat)
	return &d
}

// prepareSubsList вспомогательная функция для формирование ответа по подписке списком
func prepareSubsList(subsList []entity.Subscription) []Subscription {
	subsListResp := make([]Subscription, 0, len(subsList))
	for _, subs := range subsList {
		subsListResp = append(subsListResp, prepareSubs(subs))
	}
	return subsListResp
}

// prepareSubs вспомогательная функция для формирование ответа по подписке
func prepareSubs(subs entity.Subscription) Subscription {
	return Subscription{
		IsDelete:    bool(subs.IsDelete),
		EndDate:     nilDate(subs.EndDate),
		Id:          int(subs.Id),
		Price:       *subs.Price,
		ServiceName: string(subs.ServiceName),
		StartDate:   subs.StartDate.Format(StartDateFormat),
		UserId:      uuid.UUID(subs.UserId),
	}
}
