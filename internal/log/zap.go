package log

import (
	"fmt"
	"io"
	"os"

	"github.com/yixy/tiny-photograph/common/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger
var W zapcore.WriteSyncer

func init() {
	logDir := fmt.Sprintf(fmt.Sprintf("%s/logs", env.Workdir))
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		fmt.Printf("Error when mkdir %s", logDir)
		return
	}
	InitLogger(fmt.Sprintf("%s/tiny-photograph.log", logDir))
}

func InitLogger(logfile string) {
	Logger = Init(logfile)
}

func Init(logfile string) *zap.Logger {
	var output io.Writer
	if logfile != "" {
		// lumberjack.Logger is already safe for concurrent use, so we don't need to
		// lock it.
		output = &lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    128, // megabytes for MB
			MaxBackups: 2,
			MaxAge:     365,  // days
			Compress:   true, // disabled by default
		}
	} else {
		output = os.Stdout
	}

	W = zapcore.AddSync(output)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		W,
		zap.InfoLevel,
	)
	// dev mod
	caller := zap.AddCaller()
	// filename and line
	development := zap.Development()
	// initial app name
	filed := zap.Fields(zap.String("serviceName", env.AppName))
	logger := zap.New(core, caller, development, filed)
	defer logger.Sync()
	return logger
}
