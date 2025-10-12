package mqtt

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestLogWrapperOutput(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	wrapper := NewLogWrapper(handler)

	logger := slog.New(wrapper)

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warning")
	logger.Error("test", slog.String("error", errors.New("error msg").Error()))

	out := buf.String()

	if !strings.Contains(out, "level=DEBUG msg=debug") {
		t.Errorf("Expected output to contain 'level=DEBUG msg=debug', got: %s", out)
	}
	if !strings.Contains(out, "level=INFO msg=info") {
		t.Errorf("Expected output to contain 'level=INFO msg=info', got: %s", out)
	}
	if !strings.Contains(out, "level=WARN msg=warning") {
		t.Errorf("Expected output to contain 'level=WARN msg=warning', got: %s", out)
	}
	if !strings.Contains(out, "level=ERROR msg=test error=\"error msg\"") {
		t.Errorf(`Expected output to contain 'level=ERROR msg=test error="error msg"', got: %s`, out)
	}
}
