package repository

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	"effective-mobile-project/internal/infrastructure/repository/errors_repo"
	"effective-mobile-project/internal/usecase"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"time"
)

var (
	_ usecase.CreateSubsRepository = (*Repository)(nil)
	_ usecase.GetSubsRepository    = (*Repository)(nil)
)

type Repository struct {
	pool *pgxpool.Pool
}

// NewPostgresPool отдельный конструктор для пула, т.к. он должен быть общим для всех репозиториев
func NewPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	db := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, db,
	)
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("DB ping error:  %w", err)
	}

	log.Println("Postgres pool created.")
	return pool, nil
}

// NewRepository конструктор для репозитория
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
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
		"start_date":   time.Time(subs.StartDate), // с кастомным типом даты ошибка при записи
	}
	if err = r.pool.QueryRow(ctx, query, args).Scan(&id); err != nil {
		return 0, fmt.Errorf("Create error: %w", err)
	}
	return
}

// Get получение подписки
func (r *Repository) Get(ctx context.Context, id entity.IdSubs) (subs entity.Subscription, err error) {
	query := `select id, service_name, price, user_id, start_date, end_date, delete_date
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
			&subs.DeleteDate,
		)
	if err != nil {
		return entity.Subscription{}, fmt.Errorf("repository.GetArticleToSend r.pool.Query error: %w", err)
	}
	return
}

// UpdateSubs изменение подписки
func (r *Repository) Update(ctx context.Context, subs entity.Subscription) (err error) {
	query := `UPDATE subscription 
				 SET price = $1
				    ,end_date = $2
				    ,delete_date = $3
			   WHERE id = $4`

	cmd, err := r.pool.Exec(ctx, query, subs.Price, subs.EndDate, subs.DeleteDate, subs.Id)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return errors_repo.ErrSubsNotFound
	}
	return
}

//// InsertBatch вставка слайса статей в news
//func (r *Repository) InsertBatch(ctx context.Context, articles []domain.Article) error {
//	var errs []error
//	query := `INSERT INTO news(source, category, article_id, title, url, data_json, published_at)
//			  VALUES (@source, @category, @article_id, @title, @url, @data_json, @published_at)`
//	batch := &pgx.Batch{}
//	for _, article := range articles {
//		articleDTO, err := r.articleDTO(article)
//		if err != nil {
//			return fmt.Errorf("NewsRepository.InsertBatch error: %w", err)
//		}
//		args := pgx.NamedArgs{
//			"source":       articleDTO.Source,
//			"category":     articleDTO.Category,
//			"article_id":   articleDTO.ArticleId,
//			"title":        articleDTO.Title,
//			"url":          articleDTO.URL,
//			"data_json":    articleDTO.DataJson,
//			"published_at": articleDTO.PublishedAt,
//		}
//		batch.Queue(query, args)
//	}
//
//	br := r.pool.SendBatch(ctx, batch)
//
//	defer br.Close()
//	for range articles {
//		if _, err := br.Exec(); err != nil {
//			errs = append(errs, err)
//		}
//	}
//	if len(errs) > 0 {
//		return fmt.Errorf("NewsRepository.InsertBatch error: %w", errors.Join(errs...))
//	}
//	return nil
//}
//
//// GetLastArticleId получение ID последней статьи за день
//func (r *Repository) GetLastArticleId(ctx context.Context, path string) (int, error) {
//	var articleIdLast int
//	query := `select article_id
//			    from news
//			   where published_at > CURRENT_DATE
//			     and url like $1
//			   order by published_at desc limit 1`
//	row := r.pool.QueryRow(ctx, query, path+"%")
//	if err := row.Scan(&articleIdLast); err != nil {
//		return 0, fmt.Errorf("NewsRepository.GetLastArticleUrl error: %w", err)
//	}
//	slog.Debug("GetLastArticleDt", "articleIdLast", articleIdLast)
//	return articleIdLast, nil
//}
//
//// GetArticleToSend получение статей на отправку за день
//func (r *Repository) GetArticleToSend(ctx context.Context) ([]string, error) {
//	query := `select data_json
//			    from news
//			   where published_at > CURRENT_DATE
//			     and is_send = false`
//	row, err := r.pool.Query(ctx, query)
//	if err != nil {
//		return nil, fmt.Errorf("repository.GetArticleToSend r.pool.Query error: %w", err)
//	}
//	articlesJson := make([]string, 0)
//	data := ""
//	for row.Next() {
//		if err = row.Scan(&data); err != nil {
//			return nil, fmt.Errorf("repository.GetArticleToSend row.Scan(&data) error: %w", err)
//		}
//		articlesJson = append(articlesJson, data)
//	}
//	return articlesJson, nil
//}
