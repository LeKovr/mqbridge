package main

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SetupLog creates logger
func SetupLog(withDebug bool, opts ...zap.Option) logr.Logger {
	var log logr.Logger
	if withDebug {
		aa := zap.NewDevelopmentEncoderConfig()
		zo := append(opts, zap.AddCaller())
		aa.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapLog := zap.New(zapcore.NewCore(
			zapcore.NewConsoleEncoder(aa),
			zapcore.AddSync(colorable.NewColorableStdout()),
			zapcore.DebugLevel,
		),
			zo...,
		)
		log = zapr.NewLogger(zapLog)
	} else {
		zc := zap.NewProductionConfig()
		zapLog, _ := zc.Build(opts...)
		log = zapr.NewLogger(zapLog)
	}
	return log
}
