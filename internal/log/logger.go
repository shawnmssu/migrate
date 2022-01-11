package log

import (
	"github.com/ucloud/migrate/internal/conf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"path"
)

var Logger *zap.Logger

func InitLogger(config *conf.Log) error {
	var err error
	var level zapcore.Level
	if config == nil {
		configStr := zap.NewProductionConfig()
		configStr.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		Logger, err = configStr.Build()
		if err != nil {
			return err
		}
		return nil
	}

	switch config.Level {
	case "DEBUG":
		level = zap.DebugLevel
	case "INFO":
		level = zap.InfoLevel
	case "WARN":
		level = zap.WarnLevel
	case "ERROR":
		level = zap.ErrorLevel
	default:
		level = zap.DebugLevel
	}

	if !config.IsStdout {
		dir := config.Dir
		name := config.Name
		if dir == "" {
			dir = "./build"
		}
		if name == "" {
			name = "migrate"
		}
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename: path.Join(dir, name),
		})
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			w,
			level,
		)
		Logger = zap.New(core)
	} else {
		configStr := zap.NewProductionConfig()
		configStr.Level = zap.NewAtomicLevelAt(level)
		Logger, err = configStr.Build()
		if err != nil {
			return err
		}
	}

	defer func() {
		_ = Logger.Sync()
	}()

	return nil
}
