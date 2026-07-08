package lib

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/pkg/file"
)

// Logger is the application logger. It is a plain *slog.Logger so business code
// depends only on the standard library — the concrete backend (currently zap +
// lumberjack, wired below via zapslog) can be swapped by touching this file
// alone. See docs/infra-abstraction-plan.md §3.
//
// NOTE: this is an alias, so `lib.Logger` and `*slog.Logger` are the same type.
// Dropping the alias name in favour of writing `*slog.Logger` everywhere is a
// mechanical follow-up; keeping it here avoids churning 50+ pass-through
// signatures in this pass.
type Logger = *slog.Logger

// NewLogger builds the application logger. slog is the front end; a zap core
// (JSON/console encoder + stdout + lumberjack rotation) is the back end, bridged
// by zapslog so the on-disk format and rotation are unchanged from the previous
// zap-only setup.
func NewLogger(config Config) Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeTime:     localTimeEncoder,
	}

	var encoder zapcore.Encoder
	if config.Log.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	level := zap.NewAtomicLevelAt(toLevel(config.Log.Level))
	core := zapcore.NewCore(encoder, toWriter(config), level)

	handler := zapslog.NewHandler(core,
		zapslog.WithCaller(true),
		zapslog.AddStacktraceAt(slog.LevelWarn),
	)
	return slog.New(handler)
}

// NopLogger returns a logger that discards all records. Intended for tests.
func NopLogger() Logger {
	return slog.New(slog.DiscardHandler)
}

func localTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(constants.TimeFormat))
}

func toLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dpanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func toWriter(config Config) zapcore.WriteSyncer {
	fp := ""
	sp := string(filepath.Separator)

	fp, _ = filepath.Abs(filepath.Dir(filepath.Join(".")))
	fp += sp + "logs" + sp

	if config.Log.Directory != "" {
		if err := file.EnsureDirRW(config.Log.Directory); err == nil {
			fp = config.Log.Directory
		}
	}

	return zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout),
		zapcore.AddSync(&lumberjack.Logger{ // 文件切割
			Filename:   filepath.Join(fp, config.Name) + ".log",
			MaxSize:    100,
			MaxAge:     30,
			MaxBackups: 10,
			LocalTime:  true,
			Compress:   true,
		}),
	)
}
