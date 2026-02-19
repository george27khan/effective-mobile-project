package logger

import (
	"context"
	"effective-mobile-project/internal/domain/entity"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type keyType int

const key = keyType(0)

type HandlerMiddlware struct {
	next slog.Handler
}

func NewHandlerMiddleware(next slog.Handler) *HandlerMiddlware {
	return &HandlerMiddlware{next: next}
}

func (h *HandlerMiddlware) Enabled(ctx context.Context, rec slog.Level) bool {
	return h.next.Enabled(ctx, rec)
}

func (h *HandlerMiddlware) Handle(ctx context.Context, rec slog.Record) error {
	if c, ok := ctx.Value(key).(logCtx); ok {
		if c.ID != 0 {
			rec.Add("IdSubs", c.ID)
		}
		if c.ServiceName != "" {
			rec.Add("ServiceName", c.ServiceName)
		}
		if c.Price != nil {
			rec.Add("Price", c.Price)
		}
		if c.UserId != [16]byte{} {
			rec.Add("UserId", uuid.UUID(c.UserId).String())
		}
		if !c.StartDate.IsZero() {
			rec.Add("StartDate", c.StartDate)
		}
		if c.EndDate != nil {
			rec.Add("EndDate", c.EndDate)
		}
		if c.IsDelete != false {
			rec.Add("IsDelete", c.IsDelete)
		}
		if c.Stack != nil {
			rec.Add("Stack", c.Stack)
		}
	}
	return h.next.Handle(ctx, rec)
}

func (h *HandlerMiddlware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &HandlerMiddlware{next: h.next.WithAttrs(attrs)} // не забыть обернуть, но осторожно
}

func (h *HandlerMiddlware) WithGroup(name string) slog.Handler {
	return &HandlerMiddlware{next: h.next.WithGroup(name)} // не забыть обернуть, но осторожно
}

type logCtx struct {
	ID          entity.IdSubs
	ServiceName entity.ServiceName
	Price       entity.Price
	UserId      entity.UserId
	StartDate   time.Time
	EndDate     *time.Time
	IsDelete    entity.IsDelete
	Stack       []string
}

func WithID(ctx context.Context, id entity.IdSubs) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.ID = id
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{ID: id})
}

func WithUserID(ctx context.Context, userId entity.UserId) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.UserId = userId
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{UserId: userId})
}

func WithServiceName(ctx context.Context, serviceName entity.ServiceName) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.ServiceName = serviceName
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{ServiceName: serviceName})
}

func WithSubs(ctx context.Context, subs entity.Subscription) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.ID = subs.Id
		c.ServiceName = subs.ServiceName
		c.Price = subs.Price
		c.UserId = subs.UserId
		c.StartDate = subs.StartDate
		c.EndDate = subs.EndDate
		c.IsDelete = subs.IsDelete

		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{
		ID:          subs.Id,
		ServiceName: subs.ServiceName,
		Price:       subs.Price,
		UserId:      subs.UserId,
		StartDate:   subs.StartDate,
		EndDate:     subs.EndDate,
		IsDelete:    subs.IsDelete,
	})
}

func WithStack(ctx context.Context, stack []string) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.Stack = stack
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{Stack: stack})
}

func ShortStack(skip int) []string {
	const depth = 10
	var pcs [depth]uintptr

	n := runtime.Callers(skip, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var stack []string
	for {
		frame, more := frames.Next()

		stack = append(stack,
			fmt.Sprintf("%s:%d %s",
				frame.File, frame.Line, frame.Function),
		)

		if !more {
			break
		}
	}
	return stack
}

// -----------------------------------------------

type errorWithLogCtx struct {
	next error
	ctx  logCtx
}

// Error для реализации интерфейса
func (e *errorWithLogCtx) Error() string {
	return e.next.Error()
}

// Unwrap для реализации интерфейса
func (e *errorWithLogCtx) Unwrap() error {
	return e.next
}

func WrapError(ctx context.Context, err error) error {
	c := logCtx{}
	if x, ok := ctx.Value(key).(logCtx); ok {
		c = x
	}
	return &errorWithLogCtx{
		next: err,
		ctx:  c,
	}
}

func ErrorCtx(ctx context.Context, err error) context.Context {
	if e, ok := err.(*errorWithLogCtx); ok { // в реальной жизни используйте error.As
		return context.WithValue(ctx, key, e.ctx)
	}
	return ctx
}

// -----------------------------------------------

func InitLogging() *slog.Logger {
	handler := slog.Handler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	handler = NewHandlerMiddleware(handler)
	//slog.SetDefault(slog.New(handler))
	return slog.New(handler)
}
