package logger

import (
	"shared/app"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RequestMdc struct {
	Subsystem uuid.UUID
	Request   uuid.UUID
}

func NewLogger(application app.Application) (*zap.Logger, error) {
	switch application.Environment {
	case "prod":
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		return cfg.Build()
	case "dev":
		return zap.NewDevelopment()
	}
	return zap.NewExample(), nil
}

func withMdc(l *zap.Logger, mdc *RequestMdc) *zap.Logger {
	return l.With(zap.Any("subsystem", mdc.Subsystem), zap.Any("request", mdc.Request))
}

func LogInfo(l *zap.Logger, mdc *RequestMdc, message string) {
	withMdc(l, mdc).Info(message)
}

func LogError(l *zap.Logger, mdc *RequestMdc, message string, err error) {
	withMdc(l, mdc).Error(message, zap.Error(err))
}

func LogWarn(l *zap.Logger, mdc *RequestMdc, message string, err error) {
	withMdc(l, mdc).Warn(message, zap.Error(err))
}

func LogDebug(l *zap.Logger, mdc *RequestMdc, message string) {
	withMdc(l, mdc).Debug(message)
}
