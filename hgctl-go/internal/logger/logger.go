package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Title(msg string)
	Sugar() *zap.SugaredLogger
}

type ZapLogger struct {
	*zap.Logger
	writer io.Writer
}

// LoggerOptions configures the logger
type LoggerOptions struct {
	Verbose bool
	Writer  io.Writer
}

// NewLogger creates a logger with default options (writes to stderr)
func NewLogger(verbose bool) Logger {
	return NewLoggerWithOptions(LoggerOptions{
		Verbose: verbose,
		Writer:  os.Stderr,
	})
}

// NewLoggerWithWriter creates a logger with a custom writer
func NewLoggerWithWriter(verbose bool, w io.Writer) Logger {
	return NewLoggerWithOptions(LoggerOptions{
		Verbose: verbose,
		Writer:  w,
	})
}

// NewLoggerWithOptions creates a logger with full configuration options
func NewLoggerWithOptions(opts LoggerOptions) Logger {
	// Default to stderr if no writer provided
	if opts.Writer == nil {
		opts.Writer = os.Stderr
	}

	// Custom encoder config to match TypeScript format
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    coloredLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create console encoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// Set log level
	level := zapcore.InfoLevel
	if opts.Verbose {
		level = zapcore.DebugLevel
	}

	// Create core with custom writer
	core := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(opts.Writer),
		level,
	)

	// Create logger
	logger := zap.New(core)

	return &ZapLogger{
		Logger: logger,
		writer: opts.Writer,
	}
}

func (l *ZapLogger) Title(msg string) {
	fmt.Fprintln(l.writer)
	color.New(color.FgCyan, color.Bold).Fprintln(l.writer, msg)
	fmt.Fprintln(l.writer)
}

// Custom encoder for colored level output
func coloredLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var levelColor *color.Color
	var levelText string

	switch l {
	case zapcore.DebugLevel:
		levelColor = color.New(color.FgWhite)
		levelText = "DEBUG"
	case zapcore.InfoLevel:
		levelColor = color.New(color.FgBlue)
		levelText = "INFO"
	case zapcore.WarnLevel:
		levelColor = color.New(color.FgYellow)
		levelText = "WARN"
	case zapcore.ErrorLevel:
		levelColor = color.New(color.FgRed)
		levelText = "ERROR"
	case zapcore.FatalLevel:
		levelColor = color.New(color.FgRed, color.Bold)
		levelText = "FATAL"
	default:
		levelColor = color.New(color.FgWhite)
		levelText = l.String()
	}

	enc.AppendString(levelColor.Sprint(levelText))
}

// Custom time encoder to match TypeScript format
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(color.New(color.FgWhite).Sprintf("[%s]", t.Format("15:04:05")))
}

// Global logger instance
var globalLogger Logger

// InitGlobalLogger initializes the global logger (backward compatible)
func InitGlobalLogger(verbose bool) {
	globalLogger = NewLogger(verbose)
}

// InitGlobalLoggerWithWriter initializes the global logger with a custom writer
func InitGlobalLoggerWithWriter(verbose bool, w io.Writer) {
	globalLogger = NewLoggerWithWriter(verbose, w)
}

// InitGlobalLoggerWithOptions initializes the global logger with full options
func InitGlobalLoggerWithOptions(opts LoggerOptions) {
	globalLogger = NewLoggerWithOptions(opts)
}

// GetLogger returns the global logger
func GetLogger() Logger {
	if globalLogger == nil {
		globalLogger = NewLogger(false)
	}
	return globalLogger
}
