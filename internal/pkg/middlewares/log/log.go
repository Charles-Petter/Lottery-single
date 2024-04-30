package log

import (
	"context"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"lottery_single/internal/pkg/constant"
	"os"
)

var (
	log *zap.Logger
)

// Logger 默认日志组件使用zap
type Logger interface {
	Error(v ...interface{})
	Warn(v ...interface{})
	Info(v ...interface{})
	Debug(v ...interface{})
	Errorf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

var (
	logger Logger
)

func Init(opts ...Option) {
	// 初始化
	logger = newSugarLogger(newOptions(opts...))
}

// Options 选项配置
type Options struct {
	logPath    string // 日志路径
	fileName   string // 日志名称
	logLevel   string // 日志级别
	maxSize    int    // 日志保留大小，以 M 为单位
	maxBackups int    // 保留文件个数
}

// Option 选项方法
type Option func(*Options)

// newOptions 初始化
func newOptions(opts ...Option) Options {
	// 默认配置
	options := Options{
		fileName:   "lottery_single.log",
		logLevel:   "info",
		maxSize:    100,
		maxBackups: 3,
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// WithLogLevel 日志级别
func WithLogLevel(level string) Option {
	return func(o *Options) {
		o.logLevel = level
	}
}

func WithFileName(fileName string) Option {
	return func(o *Options) {
		o.fileName = fileName
	}
}

func WithLogPath(logPath string) Option {
	return func(o *Options) {
		o.logPath = logPath
	}
}

func WithMaxSize(maxSize int) Option {
	return func(o *Options) {
		o.maxSize = maxSize
	}
}
func WithMaxBackups(maxBackups int) Option {
	return func(o *Options) {
		o.maxBackups = maxBackups
	}
}

type zapLoggerWrapper struct {
	*zap.SugaredLogger
	options Options
}

func newSugarLogger(options Options) *zapLoggerWrapper {
	w := &zapLoggerWrapper{options: options}
	encoder := w.getEncoder()
	w.setSugaredLogger(encoder)
	return w
}
func (w *zapLoggerWrapper) setSugaredLogger(encoder zapcore.Encoder) {
	var coreArr []zapcore.Core
	// info文件writeSyncer
	// 日志级别
	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { // error级别
		return lev >= zap.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { // info和debug级别,debug级别是最低的
		if w.options.logLevel == "debug" {
			return lev < zap.ErrorLevel && lev >= zap.DebugLevel
		} else {
			return lev < zap.ErrorLevel && lev >= zap.InfoLevel
		}
	})
	infoFileWriteSyncer := w.getLogWriter("info_")
	infoFileCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(infoFileWriteSyncer, zapcore.AddSync(os.Stdout)), lowPriority)
	errorFileWriteSyncer := w.getLogWriter("error_")
	errorFileCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(errorFileWriteSyncer, zapcore.AddSync(os.Stdout)), highPriority)
	coreArr = append(coreArr, infoFileCore)
	coreArr = append(coreArr, errorFileCore)
	log = zap.New(zapcore.NewTee(coreArr...), zap.AddCaller(), zap.AddCallerSkip(1)) // zap.AddCaller()为显示文件名和行号，可省略
	w.SugaredLogger = log.Sugar()
}
func (w *zapLoggerWrapper) getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 在日志文件中使用大写字母记录日志级别
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// NewConsoleEncoder 打印更符合人们观察的方式
	return zapcore.NewConsoleEncoder(encoderConfig)
}
func (w *zapLoggerWrapper) getLogWriter(typeName string) zapcore.WriteSyncer {
	logf, err := rotatelogs.New(
		w.options.logPath+"/"+typeName+"_%Y-%m-%d_"+w.options.fileName,
		//rotatelogs.WithMaxAge(24*time.Hour),
		rotatelogs.WithRotationCount(uint(w.options.maxBackups)),
		//rotatelogs.WithRotationTime(time.Minute),
		rotatelogs.WithRotationSize(int64(w.options.maxSize*1024*1024)),
	)

	if err != nil {
		panic(err)
	}

	return zapcore.AddSync(logf)
}

// getDefaultLogger 获取默认日志实现
func getDefaultLogger() Logger {
	return logger
}

// Debugf 打印 Debug 日志
func Debugf(format string, args ...interface{}) {
	getDefaultLogger().Debugf(format, args...)
}

// Infof 打印 Info 日志
func Infof(format string, args ...interface{}) {
	getDefaultLogger().Infof(format, args...)
}

// Warnf 打印 Warn 日志
func Warnf(format string, args ...interface{}) {
	getDefaultLogger().Warnf(format, args...)
}

// Errorf 打印 Error 日志
func Errorf(format string, args ...interface{}) {
	getDefaultLogger().Errorf(format, args...)
}

// DebugContextf 打印 Debug 日志
func DebugContextf(ctx context.Context, format string, args ...interface{}) {
	value := ctx.Value(constant.ReqID)
	args = append([]interface{}{value}, args...)
	getDefaultLogger().Debugf(constant.ReqID+":%s|"+format, args...)
}

// InfoContext 打印 Info 日志
func InfoContext(ctx context.Context, args ...interface{}) {
	getDefaultLogger().Info(args...)
}

// InfoContextf 打印 Info 日志
func InfoContextf(ctx context.Context, format string, args ...interface{}) {
	value := ctx.Value(constant.ReqID)
	args = append([]interface{}{value}, args...)
	getDefaultLogger().Infof(constant.ReqID+":%s|"+format, args...)
}

// WarnContext 打印 Warn 日志
func WarnContext(ctx context.Context, args ...interface{}) {
	getDefaultLogger().Warn(args...)
}

// WarnContextf 打印 Warn 日志
func WarnContextf(ctx context.Context, format string, args ...interface{}) {
	value := ctx.Value(constant.ReqID)
	args = append([]interface{}{value}, args...)
	getDefaultLogger().Warnf(constant.ReqID+":%s|"+format, args...)
}

func ErrorContextf(ctx context.Context, format string, args ...interface{}) {
	value := ctx.Value(constant.ReqID)
	args = append([]interface{}{value}, args...)
	getDefaultLogger().Errorf(constant.ReqID+":%s|"+format, args...)
}
func Fatalf(format string, args ...interface{}) {
	Errorf(format, args...)
}
