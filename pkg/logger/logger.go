package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog with additional context.
type Logger struct {
	log zerolog.Logger
}

// Config holds logger configuration.
type Config struct {
	Level      string
	Pretty     bool
	TimeFormat string
}

// DefaultConfig returns default logger configuration.
func DefaultConfig() Config {
	return Config{
		Level:      "info",
		Pretty:     true,
		TimeFormat: time.RFC3339,
	}
}

// New creates a new logger with the given configuration.
func New(cfg Config) *Logger {
	var output io.Writer = os.Stderr

	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: cfg.TimeFormat,
		}
	}

	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	log := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{log: log}
}

// Default creates a default logger.
func Default() *Logger {
	return New(DefaultConfig())
}

// With returns a new logger with additional fields.
func (l *Logger) With() *LogContext {
	return &LogContext{ctx: l.log.With()}
}

// Debug logs at debug level.
func (l *Logger) Debug() *zerolog.Event {
	return l.log.Debug()
}

// Info logs at info level.
func (l *Logger) Info() *zerolog.Event {
	return l.log.Info()
}

// Warn logs at warn level.
func (l *Logger) Warn() *zerolog.Event {
	return l.log.Warn()
}

// Error logs at error level.
func (l *Logger) Error() *zerolog.Event {
	return l.log.Error()
}

// Fatal logs at fatal level and exits.
func (l *Logger) Fatal() *zerolog.Event {
	return l.log.Fatal()
}

// LogContext wraps zerolog.Context for chaining.
type LogContext struct {
	ctx zerolog.Context
}

// Str adds a string field.
func (c *LogContext) Str(key, val string) *LogContext {
	c.ctx = c.ctx.Str(key, val)
	return c
}

// Int adds an int field.
func (c *LogContext) Int(key string, val int) *LogContext {
	c.ctx = c.ctx.Int(key, val)
	return c
}

// Bool adds a bool field.
func (c *LogContext) Bool(key string, val bool) *LogContext {
	c.ctx = c.ctx.Bool(key, val)
	return c
}

// Err adds an error field.
func (c *LogContext) Err(err error) *LogContext {
	c.ctx = c.ctx.Err(err)
	return c
}

// Logger returns the logger with the added context.
func (c *LogContext) Logger() *Logger {
	return &Logger{log: c.ctx.Logger()}
}

// Context key for request ID.
type contextKey string

const RequestIDKey contextKey = "request_id"

// FromContext returns a logger with context values.
func FromContext(ctx context.Context, l *Logger) *Logger {
	if l == nil {
		l = Default()
	}

	c := l.log.With()

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		c = c.Str("request_id", requestID)
	}

	return &Logger{log: c.Logger()}
}

// Global logger instance.
var global = Default()

// SetGlobal sets the global logger.
func SetGlobal(l *Logger) {
	global = l
}

// Global returns the global logger.
func Global() *Logger {
	return global
}

// Convenience functions using global logger.
func Debug() *zerolog.Event { return global.Debug() }
func Info() *zerolog.Event  { return global.Info() }
func Warn() *zerolog.Event  { return global.Warn() }
func Error() *zerolog.Event { return global.Error() }
func Fatal() *zerolog.Event { return global.Fatal() }
