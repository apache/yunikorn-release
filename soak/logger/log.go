package logger

import (
	"github.com/apache/yunikorn-core/pkg/log"
	"go.uber.org/zap"
	"strconv"
)

var Logger *zap.Logger = log.Log(log.Test)

func SetLogLevel(level int) {
	log.UpdateLoggingConfig(map[string]string{
		"log.level": strconv.Itoa(level),
	})
}
