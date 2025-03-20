package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger to provide a simpler interface
type Logger struct {
	*zap.SugaredLogger
}

// Config holds logger configuration
type Config struct {
	Level      string `json:"level"`
	OutputPath string `json:"output_path"`
	Format     string `json:"format"` // "json" or "console"
}

// New creates a new logger instance
func New(config *Config) (*Logger, error) {
	level := zap.NewAtomicLevel()
	err := level.UnmarshalText([]byte(config.Level))
	if err != nil {
		level.SetLevel(zapcore.InfoLevel)
	}

	zapConfig := zap.Config{
		Level:            level,
		OutputPaths:      []string{config.OutputPath},
		ErrorOutputPaths: []string{config.OutputPath},
		Encoding:         config.Format,
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:    "msg",
			LevelKey:      "level",
			TimeKey:       "time",
			NameKey:       "logger",
			CallerKey:     "caller",
			FunctionKey:   zapcore.OmitKey,
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.LowercaseLevelEncoder,
			EncodeTime:    zapcore.ISO8601TimeEncoder,
			EncodeCaller:  zapcore.ShortCallerEncoder,
		},
	}

	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, err
	}

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}, nil
}

// NewDefault creates a new logger with default configuration
func NewDefault() *Logger {
	config := &Config{
		Level:      "debug",
		OutputPath: "stdout",
		Format:     "console",
	}

	logger, err := New(config)
	if err != nil {
		// Fallback to a basic logger if configuration fails
		zapLogger, _ := zap.NewProduction()
		return &Logger{
			SugaredLogger: zapLogger.Sugar(),
		}
	}

	return logger
}

// With adds structured context to the logger
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{
		SugaredLogger: l.SugaredLogger.With(args...),
	}
}

// Debug logs a message at debug level
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Debugw(msg, keysAndValues...)
}

// Info logs a message at info level
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Infow(msg, keysAndValues...)
}

// Warn logs a message at warn level
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Warnw(msg, keysAndValues...)
}

// Error logs a message at error level
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Errorw(msg, keysAndValues...)
}

// Fatal logs a message at fatal level and then calls os.Exit(1)
func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Fatalw(msg, keysAndValues...)
}
