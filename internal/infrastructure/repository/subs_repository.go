package repository

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	"effective-mobile-project/internal/infrastructure/repository/errors_repo"
	"effective-mobile-project/internal/pkg/logger"
	"effective-mobile-project/internal/usecase"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

var (
	_ usecase.CreateSubsRepository       = (*Repository)(nil)
	_ usecase.GetSubsRepository          = (*Repository)(nil)
	_ usecase.GetSubsListRepository      = (*Repository)(nil)
	_ usecase.GetSubsListPriceRepository = (*Repository)(nil)
	_ usecase.UpdateSubsRepository       = (*Repository)(nil)
	_ usecase.CancelSubsRepository       = (*Repository)(nil)
)

type Repository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewPostgresPool отдельный конструктор для пула, т.к. он должен быть общим для всех репозиториев
func NewPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	//user := os.Getenv("POSTGRES_USER")
	//pass := os.Getenv("POSTGRES_PASSWORD")
	//db := os.Getenv("POSTGRES_DB")
	//host := os.Getenv("POSTGRES_HOST")
	//port := os.Getenv("POSTGRES_PORT")
	user := "postgres"
	pass := "postgres"
	db := "postgres"
	host := "localhost"
	port := "5432"
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, db,
	)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}
	return pool, nil
}

// NewRepository конструктор для репозитория
func NewRepository(pool *pgxpool.Pool, logger *slog.Logger) *Repository {
	return &Repository{pool: pool, logger: logger}
}

// Close закрытие пула соединений
func (r *Repository) Close() {
	r.pool.Close()
}

// Create вставка подписки в БД
func (r *Repository) Create(ctx context.Context, subs entity.Subscription) (id entity.IdSubs, err error) {
	query := `INSERT INTO subscription(service_name, price, user_id, start_date) 
			  VALUES (@service_name, @price, @user_id, @start_date)
			  RETURNING id`

	args := pgx.NamedArgs{
		"service_name": subs.ServiceName,
		"price":        subs.Price,
		"user_id":      subs.UserId,
		"start_date":   subs.StartDate, // с кастомным типом даты ошибка при записи
	}
	if err = r.pool.QueryRow(ctx, query, args).Scan(&id); err != nil {
		ctx = logger.WithSubs(ctx, subs)
		ctx = logger.WithStack(ctx, logger.ShortStack(2))
		return 0, logger.WrapError(ctx, err)
	}
	return
}

// Get получение подписки
func (r *Repository) Get(ctx context.Context, id entity.IdSubs) (subs entity.Subscription, err error) {
	query := `select id, service_name, price, user_id, start_date, end_date, is_delete
				from subscription
			   where id = $1`
	err = r.pool.
		QueryRow(ctx, query, id).
		Scan(
			&subs.Id,
			&subs.ServiceName,
			&subs.Price,
			&subs.UserId,
			&subs.StartDate,
			&subs.EndDate,
			&subs.IsDelete,
		)
	if errors.Is(err, pgx.ErrNoRows) {
		ctx = logger.WithID(ctx, id)
		ctx = logger.WithStack(ctx, logger.ShortStack(2))
		return entity.Subscription{}, logger.WrapError(ctx, errors_repo.ErrSubsNotFound)
	}
	if err != nil {
		ctx = logger.WithID(ctx, id)
		ctx = logger.WithStack(ctx, logger.ShortStack(2))
		return entity.Subscription{}, logger.WrapError(ctx, err)
	}
	return
}

// Update изменение подписки
func (r *Repository) Update(ctx context.Context, subs entity.Subscription) (err error) {
	query := `UPDATE subscription 
				 SET price = $1
				    ,end_date = $2
				    ,is_delete = $3
			   WHERE id = $4`

	cmd, err := r.pool.Exec(ctx, query, subs.Price, subs.EndDate, subs.IsDelete, subs.Id)

	if err != nil {
		ctx = logger.WithSubs(ctx, subs)
		ctx = logger.WithStack(ctx, logger.ShortStack(2))
		return logger.WrapError(ctx, err)
	}

	if cmd.RowsAffected() == 0 {
		ctx = logger.WithSubs(ctx, subs)
		ctx = logger.WithStack(ctx, logger.ShortStack(2))
		return logger.WrapError(ctx, errors_repo.ErrSubsNotFound)
	}
	return
}

// GetList получение списка подписок
func (r *Repository) GetList(ctx context.Context, userId entity.UserId, limit int, offset int) (subsList []entity.Subscription, err error) {
	query := `select id, service_name, price, user_id, start_date, end_date, is_delete
				from subscription
			   where user_id = $1
			     and is_delete = false
			     and start_date <= NOW() 
			     and (NOW() <= end_date or end_date is null)
			order by start_date
			   limit $2
			  offset $3`
	row, err := r.pool.Query(ctx, query, userId, limit, offset)
	if err != nil {
		ctx = logger.WithUserID(ctx, userId)
		ctx = logger.WithStack(ctx, logger.ShortStack(2))
		return nil, logger.WrapError(ctx, err)
	}
	subs := entity.Subscription{}
	for row.Next() {
		if err = row.Scan(
			&subs.Id,
			&subs.ServiceName,
			&subs.Price,
			&subs.UserId,
			&subs.StartDate,
			&subs.EndDate,
			&subs.IsDelete,
		); err != nil {
			ctx = logger.WithUserID(ctx, userId)
			ctx = logger.WithStack(ctx, logger.ShortStack(2))
			return nil, logger.WrapError(ctx, err)
		}
		subsList = append(subsList, subs)
	}
	return subsList, nil
}

// GetListPrice получение стоимости списка подписок
func (r *Repository) GetListPrice(ctx context.Context, userId entity.UserIdNil, serviceName entity.ServiceNameNil) (subsListPrice entity.Price, err error) {
	query := `select sum(price)
				from subscription
			   where (user_id = $1 or $1 is null)
			     and (service_name = $2 or $2 is null)
			     and is_delete = false
			     and start_date <= NOW() 
			     and (NOW() <= end_date or end_date is null)`
	err = r.pool.QueryRow(ctx, query, userId, serviceName).Scan(&subsListPrice)

	if err != nil {
		if userId != nil {
			ctx = logger.WithUserID(ctx, entity.UserId(*userId))
		}
		if serviceName != nil {
			ctx = logger.WithServiceName(ctx, entity.ServiceName(*serviceName))
		}
		ctx = logger.WithStack(ctx, logger.ShortStack(2))
		return nil, logger.WrapError(ctx, err)
	}
	return
}
