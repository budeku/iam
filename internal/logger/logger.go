package logger

import (
	"budeku/iam/internal/conf"
	"os"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger atomic.Pointer[zap.SugaredLogger]

func init() {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	globalLogger.Store(l.Sugar())
}

func Init(cfg conf.Logger) error {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return err
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	var cores []zapcore.Core

	rotator := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    orDefault(cfg.MaxSizeMB, 100),
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}
	cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(rotator), level))

	if conf.Debug {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
	}

	core := zapcore.NewTee(cores...)
	l := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	globalLogger.Store(l.Sugar())
	return nil
}

func Sync() error {
	return globalLogger.Load().Sync()
}

func Debug(format string, args ...any) { globalLogger.Load().Debugf(format, args...) }
func Info(format string, args ...any)  { globalLogger.Load().Infof(format, args...) }
func Warn(format string, args ...any)  { globalLogger.Load().Warnf(format, args...) }
func Error(format string, args ...any) { globalLogger.Load().Errorf(format, args...) }
func Fatal(format string, args ...any) { globalLogger.Load().Fatalf(format, args...) }

func parseLevel(level string) (zapcore.Level, error) {
	if level == "" {
		return zapcore.InfoLevel, nil
	}
	var l zapcore.Level
	// 这里忽略，因为配置文件中已有验证
	_ = l.UnmarshalText([]byte(level))
	return l, nil
}

func orDefault(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}
