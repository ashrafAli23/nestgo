package core

import (
	"fmt"
	"os"
	"time"
)

// Logger is the logging abstraction for NestGo.
// Plug in any logger (zerolog, slog, zap, logrus) by implementing this interface.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

// Field is a structured log field (key-value pair).
type Field struct {
	Key   string
	Value interface{}
}

// F is a shorthand for creating a Field.
//
//	logger.Info("user created", core.F("user_id", 123), core.F("email", "a@b.com"))
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// ─── Default Logger ─────────────────────────────────────────────────────────

// defaultLogger is a simple logger that writes to stderr.
// Replace it with your own via SetLogger().
type defaultLogger struct{}

func (l *defaultLogger) Debug(msg string, fields ...Field) {
	l.log("DEBUG", msg, fields)
}

func (l *defaultLogger) Info(msg string, fields ...Field) {
	l.log("INFO", msg, fields)
}

func (l *defaultLogger) Warn(msg string, fields ...Field) {
	l.log("WARN", msg, fields)
}

func (l *defaultLogger) Error(msg string, fields ...Field) {
	l.log("ERROR", msg, fields)
}

func (l *defaultLogger) log(level, msg string, fields []Field) {
	ts := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	if len(fields) == 0 {
		fmt.Fprintf(os.Stderr, "[%s] %s %s\n", ts, level, msg)
		return
	}
	fieldStr := ""
	for _, f := range fields {
		fieldStr += fmt.Sprintf(" %s=%v", f.Key, f.Value)
	}
	fmt.Fprintf(os.Stderr, "[%s] %s %s%s\n", ts, level, msg, fieldStr)
}

// ─── Global Logger ──────────────────────────────────────────────────────────

var globalLogger Logger = &defaultLogger{}

// SetLogger replaces the global logger.
// Call this in main() before starting the server.
//
//	core.SetLogger(myZerologAdapter)
func SetLogger(l Logger) {
	if l != nil {
		globalLogger = l
	}
}

// Log returns the global logger instance.
//
//	core.Log().Info("server started", core.F("addr", ":3000"))
func Log() Logger {
	return globalLogger
}

// ─── Adapter Helpers ────────────────────────────────────────────────────────
// These let you wrap popular loggers in a few lines.

// LoggerFunc is a functional adapter for simple logging needs.
// Maps all levels to a single function.
type LoggerFunc func(level, msg string, fields []Field)

func (f LoggerFunc) Debug(msg string, fields ...Field) { f("DEBUG", msg, fields) }
func (f LoggerFunc) Info(msg string, fields ...Field)  { f("INFO", msg, fields) }
func (f LoggerFunc) Warn(msg string, fields ...Field)  { f("WARN", msg, fields) }
func (f LoggerFunc) Error(msg string, fields ...Field) { f("ERROR", msg, fields) }
