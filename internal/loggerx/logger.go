package loggerx

import (
	"log"

	"go.uber.org/zap"
)

func MustNewLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("while creating logger: %s", err.Error())
	}
	return logger
}
