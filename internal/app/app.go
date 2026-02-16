package app

import (
	"context"
	rep "effective-mobile-project/internal/infrastructure/repository"
	"effective-mobile-project/internal/transport/http/server"
	"effective-mobile-project/internal/transport/http/server/middleware"
	"effective-mobile-project/internal/usecase"
	"errors"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	mw "github.com/oapi-codegen/nethttp-middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	//глобальный контекст для отмены фоновых загрузок при остановке приложения
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	r := chi.NewRouter()

	spec, err := server.GetSwagger()
	if err != nil {
		log.Printf("Ошибка Swagger: %v", err)
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
		return
	}
	repository := rep.NewRepository(pool) // репозиторий
	createSubsUseCase := usecase.NewCreateSubsUseCase(repository)
	updateSubsUseCase := usecase.NewUpdateSubsUseCase(repository)
	getSubsUseCase := usecase.NewGetTaskUseCase(repository)
	cancelSubsUseCase := usecase.NewCancelTaskUseCase(repository)
	srv := server.NewSubsServer(createSubsUseCase, getSubsUseCase, updateSubsUseCase, cancelSubsUseCase)

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
		log.Printf("Start server on port 8080")
		if err := s.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("Listen error: %s", err)
		}
	}()

	//gracefull shutdown
	<-rootCtx.Done() // ожидание сигнала завершения
	log.Println("Start gracefull shutdown ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown err:", err)
	}
	log.Println("Server exiting.")
	pool.Close() // закрываем пул к базе
	log.Println("Postgres pool closed.")

}
