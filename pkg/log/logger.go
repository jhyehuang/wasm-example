package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var GinLogger *zap.Logger
var Logger *zap.SugaredLogger

func init() {

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder   // 修改时间戳的格式
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 日志级别使用大写显示
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder //

	writer := zapcore.AddSync(os.Stdout)
	encoder := zapcore.NewConsoleEncoder(encoderConfig) // 设置 console 编码器
	core := zapcore.NewCore(encoder, writer, zapcore.DebugLevel)

	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
	GinLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

}

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	Logger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	Logger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	Logger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	Logger.Errorf(template, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	Logger.Fatalf(template, args...)
}
