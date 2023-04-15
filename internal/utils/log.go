package utils

import "go.uber.org/zap"

var Logger *zap.Logger

func InitializeLogger() {
	logger, _ := zap.NewProduction()
	Logger = logger
}
