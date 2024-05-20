package utils

import (
    "go.uber.org/zap"
)

var logger *zap.Logger

func InitLogger() {
    var err error
    logger, err = zap.NewProduction()
    if err != nil {
        panic(err)
    }
}

func Logger() *zap.Logger {
    return logger
}
