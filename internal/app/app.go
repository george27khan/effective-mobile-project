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

	// Middleware проверки запросов
	r.Use(mw.OapiRequestValidatorWithOptions(
		spec, // Добавление валидатора сваггера
		&mw.Options{
			Options:      openapi3filter.Options{},
			ErrorHandler: server.SwaggerErrorHandlerFunc, // добавление обработчика ошибок на уровне проверки сваггером
		},
	))
	ctx := context.Background()
	pool, err := rep.NewPostgresPool(ctx, "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
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
	//r.Mount("/api/v1", r)
	//server.HandlerFromMux(srvStrict, r)
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
			log.Printf("listen error: %s", err)
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
	log.Println("Server exiting")

	log.Println("Background stoped")
}

//
//func GetApp() *fx.App {
//	return fx.New(
//		fx.Provide(
//			fx.Annotate(
//				fileRep.NewTaskRepository,
//				fx.As(new(task.TaskFileRepository)),
//				fx.As(new(task.GetTaskRepository)),
//				fx.As(new(task.CreateTaskRepository)),
//			),
//			fx.Annotate(task.NewAsyncRunner,
//				fx.As(new(task.BackgroundRunner))),
//			fx.Annotate(http_loader.NewHttpLoader,
//				fx.As(new(task.HttpLoader))),
//			fx.Annotate(
//				task.NewCreateTaskUseCase,
//				fx.As(new(server.TaskCreateUseCase)),
//			),
//			fx.Annotate(
//				task.NewGetTaskUseCase,
//				fx.As(new(server.TaskGetUseCase)),
//			),
//			fx.Annotate(
//				task.NewTaskFileUseCase,
//				fx.As(new(server.TaskFileUseCase)),
//			),
//			server.NewTaskServer,
//			NewHttpServer),
//		fx.Invoke(func(*http.Server) {}),
//	)
//}
//
//func NewHttpServer(lc fx.Lifecycle, ssi *server.TaskServer, runner task.BackgroundRunner) *http.Server {
//	router := chi.NewRouter()
//	spec, err := server.GetSwagger()
//	if err != nil {
//		log.Printf("Ошибка Swagger: %v", err)
//		return nil
//	}
//
//	// Middleware проверки запросов
//	router.Use(mw.OapiRequestValidatorWithOptions(
//		spec, // Добавление валидатора сваггера
//		&mw.Options{
//			Options:      openapi3filter.Options{},
//			ErrorHandler: server.SwaggerErrorHandlerFunc, // добавление обработчика ошибок на уровне проверки сваггером
//		},
//	))
//	// Регистрируем все эндпоинты из OpenAPI
//	taskSrvStrict := server.NewStrictHandlerWithOptions(
//		ssi,
//		[]server.StrictMiddlewareFunc{middleware.AddRequestId, middleware.PanicRecover},
//		server.StrictHTTPServerOptions{
//			RequestErrorHandlerFunc:  server.RequestErrorHandlerFunc,
//			ResponseErrorHandlerFunc: server.ResponseErrorHandlerFunc,
//		},
//	)
//	server.HandlerFromMux(taskSrvStrict, router)
//	srv := &http.Server{
//		Addr:    ":8080",
//		Handler: router,
//	}
//	lc.Append(fx.Hook{
//		OnStart: func(ctx context.Context) error {
//			go func() {
//				log.Printf("Start server on port 8080")
//				if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
//					log.Printf("listen error: %s", err)
//				}
//			}()
//			return nil
//		},
//		OnStop: func(ctx context.Context) error {
//			log.Println("Start server shutdown ...")
//			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
//			defer cancel()
//			//time.Sleep(10 * time.Second)
//			if err := srv.Shutdown(ctx); err != nil {
//				log.Fatal("Server shutdown err:", err)
//				return err
//			}
//			log.Println("Server exiting")
//
//			runner.(task.AsyncRunner).WgRoot.Wait() // Ожидаем завершение загрузок, теоретически не должно зависнуть т.к. загрузки с таймаутом
//			log.Println("Background finish")
//			return nil
//		},
//	})
//
//	return srv
//}
