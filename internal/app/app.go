package app

import (
	"context"
	rep "effective-mobile-project/internal/infrastructure/repository"
	"effective-mobile-project/internal/pkg/logger"
	"effective-mobile-project/internal/transport/http/server"
	"effective-mobile-project/internal/transport/http/server/middleware"
	"effective-mobile-project/internal/usecase"
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	mw "github.com/oapi-codegen/nethttp-middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	//глобальный контекст для отмены фоновых загрузок при остановке приложения
	slog := logger.InitLogging()
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	r := chi.NewRouter()

	spec, err := server.GetSwagger()
	if err != nil {
		slog.Info(fmt.Sprintf("Ошибка Swagger: %v", err))
		return
	}

	// Middleware для доступа swagger с браузера
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	}))
	// Middleware проверки запросов
	r.Use(mw.OapiRequestValidatorWithOptions(
		spec, // Добавление валидатора сваггера
		&mw.Options{
			Options:      openapi3filter.Options{},
			ErrorHandler: server.SwaggerErrorHandlerFunc, // добавление обработчика ошибок на уровне проверки сваггером
		},
	))

	ctx := context.Background()
	pool, err := rep.NewPostgresPool(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Postgres pool error: %s", err))
		return
	}
	repository := rep.NewRepository(pool, slog) // репозиторий
	createSubsUseCase := usecase.NewCreateSubsUseCase(repository, slog)
	updateSubsUseCase := usecase.NewUpdateSubsUseCase(repository, slog)
	getSubsUseCase := usecase.NewGetSubsUseCase(repository, slog)
	cancelSubsUseCase := usecase.NewCancelTaskUseCase(repository, slog)
	getSubsListUseCase := usecase.NewGetSubsListUseCase(repository, slog)
	getSubsListPriceUseCase := usecase.NewGetSubsListPriceUseCase(repository, slog)
	srv := server.NewSubsServer(createSubsUseCase, getSubsUseCase, updateSubsUseCase, cancelSubsUseCase, getSubsListUseCase, getSubsListPriceUseCase, slog)

	// Регистрируем все эндпоинты из OpenAPI
	srvStrict := server.NewStrictHandlerWithOptions(
		srv,
		[]server.StrictMiddlewareFunc{middleware.AddRequestId, middleware.PanicRecover},
		server.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  server.RequestErrorHandlerFunc,  // ловят ошибки на промежуточном уровне
			ResponseErrorHandlerFunc: server.ResponseErrorHandlerFunc, // ловят ошибки на промежуточном уровне
		},
	)

	r.Route("/api/v1", func(r chi.Router) {
		server.HandlerFromMux(srvStrict, r)
	})
	s := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	go func() {
		slog.Info("Start server on port 8080")
		if err := s.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			slog.Error(fmt.Sprintf("Listen error: %s", err))
		}
	}()

	//gracefull shutdown
	<-rootCtx.Done() // ожидание сигнала завершения
	slog.Info("Start gracefull shutdown ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		slog.Error(fmt.Sprintf("Server Shutdown err: %s", err))
	}
	slog.Info("Server exiting.")
	pool.Close() // закрываем пул к базе
	slog.Info("Postgres pool closed.")

}
