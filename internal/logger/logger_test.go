package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerInit(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expectedLvl LogLevel
	}{
		{"Production", "production", INFO},
		{"Dev", "dev", DEBUG},
		{"Development", "development", DEBUG},
		{"DEV uppercase", "DEV", DEBUG},
		{"Unknown", "unknown", INFO},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			Init(tt.environment, &buf)

			if GetLevel() != tt.expectedLvl {
				t.Errorf("Expected level %v, got %v", tt.expectedLvl, GetLevel())
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	Init("production", &buf)

	tests := []struct {
		name      string
		logFunc   func(string, ...interface{})
		level     string
		shouldLog bool
	}{
		{"DEBUG message", func(f string, args ...interface{}) { Debug(f, args...) }, "DEBUG", false},
		{"INFO message", func(f string, args ...interface{}) { Info(f, args...) }, "INFO", true},
		{"WARN message", func(f string, args ...interface{}) { Warn(f, args...) }, "WARN", true},
		{"ERROR message", func(f string, args ...interface{}) { Error(f, args...) }, "ERROR", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc("test message")

			output := buf.String()
			contains := strings.Contains(output, tt.level)

			if tt.shouldLog && !contains {
				t.Errorf("Expected log to contain %s, got: %s", tt.level, output)
			}
			if !tt.shouldLog && contains {
				t.Errorf("Expected log NOT to contain %s, got: %s", tt.level, output)
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	Init("production", &buf)

	SetLevel(DEBUG)
	if GetLevel() != DEBUG {
		t.Errorf("SetLevel failed")
	}

	SetLevel(INFO)
	if GetLevel() != INFO {
		t.Errorf("SetLevel failed")
	}
}

func TestWithFields(t *testing.T) {
	fields := map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	}

	msg := WithFields("User action", fields)

	if !strings.Contains(msg, "User action") {
		t.Errorf("Message not found in output: %s", msg)
	}
	if !strings.Contains(msg, "user_id=123") {
		t.Errorf("Field user_id not found in output: %s", msg)
	}
	if !strings.Contains(msg, "action=login") {
		t.Errorf("Field action not found in output: %s", msg)
	}
}
