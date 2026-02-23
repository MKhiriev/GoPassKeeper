package logger

import (
	"context"
	"net/http"
	"os"
	"runtime"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	zerolog.Logger
}

func NewLogger(role string) *Logger {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return runtime.FuncForPC(pc).Name() // return function name
	}

	zerolog.CallerFieldName = "func"
	logger := zerolog.New(os.Stdout).With().
		Str("role", role).
		Timestamp().
		Caller().
		Logger()

	return &Logger{logger}
}

func Nop() *Logger {
	return &Logger{zerolog.Nop()}
}

func (l *Logger) GetChildLogger() *Logger {
	return &Logger{l.With().Logger()}
}

func FromRequest(r *http.Request) *Logger {
	return &Logger{*log.Ctx(r.Context())}
}

func FromContext(ctx context.Context) *Logger {
	return &Logger{*log.Ctx(ctx)}
}
