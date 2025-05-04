package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level string) (*zap.Logger, error) {
	var l zapcore.Level
	err := l.UnmarshalText([]byte(level))
	if err != nil {
		return nil, err
	}
	
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(l)
	
	return config.Build()
}