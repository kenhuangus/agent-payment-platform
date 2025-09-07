package common

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

// LogLevel represents logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger provides structured logging
type Logger struct {
	level  LogLevel
	writer io.Writer
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		writer: os.Stdout,
	}
}

// NewLoggerWithWriter creates a new logger with custom writer
func NewLoggerWithWriter(level LogLevel, writer io.Writer) *Logger {
	return &Logger{
		level:  level,
		writer: writer,
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// log writes a log message with level and context
func (l *Logger) log(level LogLevel, message string, args ...interface{}) {
	if level < l.level {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Format the message
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := l.levelString(level)

	// Create the log message
	logMessage := fmt.Sprintf("[%s] %s %s:%d - %s",
		timestamp,
		levelStr,
		file,
		line,
		fmt.Sprintf(message, args...),
	)

	// Write to the configured writer
	fmt.Fprintln(l.writer, logMessage)

	// If it's a fatal error, exit the program
	if level == FATAL {
		os.Exit(1)
	}
}

// levelString converts LogLevel to string
func (l *Logger) levelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, args ...interface{}) {
	l.log(DEBUG, message, args...)
}

// Info logs an info message
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(INFO, message, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(WARN, message, args...)
}

// Error logs an error message
func (l *Logger) Error(message string, args ...interface{}) {
	l.log(ERROR, message, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, args ...interface{}) {
	l.log(FATAL, message, args...)
}

// Global logger instance
var defaultLogger *Logger

// init initializes the default logger
func init() {
	defaultLogger = NewLogger(INFO)
}

// SetGlobalLogLevel sets the global logging level
func SetGlobalLogLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// Debug logs a debug message using the global logger
func Debug(message string, args ...interface{}) {
	defaultLogger.Debug(message, args...)
}

// Info logs an info message using the global logger
func Info(message string, args ...interface{}) {
	defaultLogger.Info(message, args...)
}

// Warn logs a warning message using the global logger
func Warn(message string, args ...interface{}) {
	defaultLogger.Warn(message, args...)
}

// Error logs an error message using the global logger
func Error(message string, args ...interface{}) {
	defaultLogger.Error(message, args...)
}

// Fatal logs a fatal message using the global logger
func Fatal(message string, args ...interface{}) {
	defaultLogger.Fatal(message, args...)
}

// SetupDefaultLogger sets up the default Go logger to use our structured format
func SetupDefaultLogger() {
	log.SetFlags(0)
	log.SetOutput(defaultLogger.writer)
}

// LogWithContext logs a message with additional context
func (l *Logger) LogWithContext(level LogLevel, context map[string]interface{}, message string, args ...interface{}) {
	if level < l.level {
		return
	}

	// Build context string
	var contextStr string
	for key, value := range context {
		contextStr += fmt.Sprintf(" %s=%v", key, value)
	}

	// Format the message with context
	formattedMessage := fmt.Sprintf("%s%s", fmt.Sprintf(message, args...), contextStr)
	l.log(level, formattedMessage)
}

// InfoWithContext logs an info message with context
func (l *Logger) InfoWithContext(context map[string]interface{}, message string, args ...interface{}) {
	l.LogWithContext(INFO, context, message, args...)
}

// ErrorWithContext logs an error message with context
func (l *Logger) ErrorWithContext(context map[string]interface{}, message string, args ...interface{}) {
	l.LogWithContext(ERROR, context, message, args...)
}

// Performance logging
func (l *Logger) LogPerformance(operation string, start time.Time, err error) {
	duration := time.Since(start)
	context := map[string]interface{}{
		"operation": operation,
		"duration":  duration.String(),
		"success":   err == nil,
	}

	if err != nil {
		context["error"] = err.Error()
		l.ErrorWithContext(context, "Operation completed with error")
	} else {
		l.InfoWithContext(context, "Operation completed successfully")
	}
}
