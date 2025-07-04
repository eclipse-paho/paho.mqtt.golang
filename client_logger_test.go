package mqtt

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

type mockLogger struct {
	output *bytes.Buffer
}

func (m *mockLogger) Println(v ...interface{}) {
	log.New(m.output, "", 0).Println(v...)
}

func (m *mockLogger) Printf(format string, v ...interface{}) {
	log.New(m.output, "", 0).Printf(format, v...)
}

func newMockLogger() *mockLogger {
	return &mockLogger{output: &bytes.Buffer{}}
}

func TestNewClientLogger(t *testing.T) {
	errorLogger := newMockLogger()
	criticalLogger := newMockLogger()
	warnLogger := newMockLogger()
	debugLogger := newMockLogger()
	clientID := "test-client-123"

	logger := NewClientLogger(clientID, errorLogger, criticalLogger, warnLogger, debugLogger)

	if logger == nil {
		t.Fatal("NewClientLogger returned nil")
	}

	if logger.ERROR.clientID != clientID {
		t.Errorf("ERROR logger clientID = %s, want %s", logger.ERROR.clientID, clientID)
	}
	if logger.CRITICAL.clientID != clientID {
		t.Errorf("CRITICAL logger clientID = %s, want %s", logger.CRITICAL.clientID, clientID)
	}
	if logger.WARN.clientID != clientID {
		t.Errorf("WARN logger clientID = %s, want %s", logger.WARN.clientID, clientID)
	}
	if logger.DEBUG.clientID != clientID {
		t.Errorf("DEBUG logger clientID = %s, want %s", logger.DEBUG.clientID, clientID)
	}
}

func TestClientErrorLogger_Println(t *testing.T) {
	mockLog := newMockLogger()
	clientID := "test-client"

	errorLogger := clientErrorLogger{
		clientID: clientID,
		logger:   mockLog,
	}

	errorLogger.Println("test message", "with multiple", "arguments")

	output := strings.TrimSpace(mockLog.output.String())
	expected := "[test-client] test message with multiple arguments"

	if output != expected {
		t.Errorf("Println output = %q, want %q", output, expected)
	}
}

func TestClientErrorLogger_Printf(t *testing.T) {
	mockLog := newMockLogger()
	clientID := "test-client"

	errorLogger := clientErrorLogger{
		clientID: clientID,
		logger:   mockLog,
	}

	errorLogger.Printf("error code: %d, message: %s", 404, "not found")

	output := strings.TrimSpace(mockLog.output.String())
	expected := "[test-client] error code: 404, message: not found"

	if output != expected {
		t.Errorf("Printf output = %q, want %q", output, expected)
	}
}

func TestClientCriticalLogger_Println(t *testing.T) {
	mockLog := newMockLogger()
	clientID := "critical-client"

	criticalLogger := clientCriticalLogger{
		clientID: clientID,
		logger:   mockLog,
	}

	criticalLogger.Println("critical system failure")

	output := strings.TrimSpace(mockLog.output.String())
	expected := "[critical-client] critical system failure"

	if output != expected {
		t.Errorf("Println output = %q, want %q", output, expected)
	}
}

func TestClientCriticalLogger_Printf(t *testing.T) {
	mockLog := newMockLogger()
	clientID := "critical-client"

	criticalLogger := clientCriticalLogger{
		clientID: clientID,
		logger:   mockLog,
	}

	criticalLogger.Printf("critical failure at %s", "2023-07-04")

	output := strings.TrimSpace(mockLog.output.String())
	expected := "[critical-client] critical failure at 2023-07-04"

	if output != expected {
		t.Errorf("Printf output = %q, want %q", output, expected)
	}
}

func TestClientWarnLogger_Println(t *testing.T) {
	mockLog := newMockLogger()
	clientID := "warn-client"

	warnLogger := clientWarnLogger{
		clientID: clientID,
		logger:   mockLog,
	}

	warnLogger.Println("warning message")

	output := strings.TrimSpace(mockLog.output.String())
	expected := "[warn-client] warning message"

	if output != expected {
		t.Errorf("Println output = %q, want %q", output, expected)
	}
}

func TestClientWarnLogger_Printf(t *testing.T) {
	mockLog := newMockLogger()
	clientID := "warn-client"

	warnLogger := clientWarnLogger{
		clientID: clientID,
		logger:   mockLog,
	}

	warnLogger.Printf("warning: %s is deprecated", "function")

	output := strings.TrimSpace(mockLog.output.String())
	expected := "[warn-client] warning: function is deprecated"

	if output != expected {
		t.Errorf("Printf output = %q, want %q", output, expected)
	}
}

func TestClientDebugLogger_Println(t *testing.T) {
	mockLog := newMockLogger()
	clientID := "debug-client"

	debugLogger := clientDebugLogger{
		clientID: clientID,
		logger:   mockLog,
	}

	debugLogger.Println("debug information", 123)

	output := strings.TrimSpace(mockLog.output.String())
	expected := "[debug-client] debug information 123"

	if output != expected {
		t.Errorf("Println output = %q, want %q", output, expected)
	}
}

func TestClientDebugLogger_Printf(t *testing.T) {
	mockLog := newMockLogger()
	clientID := "debug-client"

	debugLogger := clientDebugLogger{
		clientID: clientID,
		logger:   mockLog,
	}

	debugLogger.Printf("debug: connection count = %d", 42)

	output := strings.TrimSpace(mockLog.output.String())
	expected := "[debug-client] debug: connection count = 42"

	if output != expected {
		t.Errorf("Printf output = %q, want %q", output, expected)
	}
}

func TestClientLogger_EmptyClientID(t *testing.T) {
	mockLog := newMockLogger()

	logger := NewClientLogger("", mockLog, mockLog, mockLog, mockLog)
	logger.ERROR.Println("test")

	output := strings.TrimSpace(mockLog.output.String())
	expected := "[] test"

	if output != expected {
		t.Errorf("Empty clientID output = %q, want %q", output, expected)
	}
}
