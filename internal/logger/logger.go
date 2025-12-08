package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

var levelColors = map[LogLevel]string{
	DEBUG: "\033[36m", // Cyan
	INFO:  "\033[32m", // Green
	WARN:  "\033[33m", // Yellow
	ERROR: "\033[31m", // Red
	FATAL: "\033[35m", // Magenta
}

const colorReset = "\033[0m"

// Logger is a custom logger with level support
type Logger struct {
	level      LogLevel
	logger     *log.Logger
	useColors  bool
	showCaller bool
}

var defaultLogger *Logger

// Init initializes the default logger
func Init(environment string, output io.Writer) {
	level := INFO
	if strings.ToLower(environment) == "dev" || strings.ToLower(environment) == "development" {
		level = DEBUG
	}

	defaultLogger = &Logger{
		level:      level,
		logger:     log.New(output, "", 0),
		useColors:  isTerminal(output),
		showCaller: level == DEBUG,
	}
}

// SetLevel sets the logging level
func SetLevel(level LogLevel) {
	if defaultLogger != nil {
		defaultLogger.level = level
		defaultLogger.showCaller = level == DEBUG
	}
}

// GetLevel returns current logging level
func GetLevel() LogLevel {
	if defaultLogger != nil {
		return defaultLogger.level
	}
	return INFO
}

// isTerminal checks if output is a terminal
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return false
		}
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// formatMessage formats the log message with timestamp and level
func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := levelNames[level]
	message := fmt.Sprintf(format, args...)

	var output string
	if l.useColors {
		color := levelColors[level]
		output = fmt.Sprintf("%s [%s%s%s] %s", timestamp, color, levelStr, colorReset, message)
	} else {
		output = fmt.Sprintf("%s [%s] %s", timestamp, levelStr, message)
	}

	// Add caller information for DEBUG level
	if l.showCaller && level == DEBUG {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			// Get only filename without full path
			parts := strings.Split(file, "/")
			filename := parts[len(parts)-1]
			output = fmt.Sprintf("%s [%s:%d]", output, filename, line)
		}
	}

	return output
}

// log writes a log message if the level is appropriate
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if l.level <= level {
		message := l.formatMessage(level, format, args...)
		l.logger.Println(message)

		// Exit on FATAL
		if level == FATAL {
			os.Exit(1)
		}
	}
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(DEBUG, format, args...)
	}
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(INFO, format, args...)
	}
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(WARN, format, args...)
	}
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(ERROR, format, args...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(FATAL, format, args...)
	}
}

// Debugf is an alias for Debug
func Debugf(format string, args ...interface{}) {
	Debug(format, args...)
}

// Infof is an alias for Info
func Infof(format string, args ...interface{}) {
	Info(format, args...)
}

// Warnf is an alias for Warn
func Warnf(format string, args ...interface{}) {
	Warn(format, args...)
}

// Errorf is an alias for Error
func Errorf(format string, args ...interface{}) {
	Error(format, args...)
}

// Fatalf is an alias for Fatal
func Fatalf(format string, args ...interface{}) {
	Fatal(format, args...)
}

// WithFields returns a formatted string with fields
func WithFields(message string, fields map[string]interface{}) string {
	parts := []string{message}
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(parts, " ")
}
