package infra

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestStdLogger_Infof(t *testing.T) {
	// Перехватываем вывод log
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := NewStdLogger()
	logger.Infof("test message %s", "value")

	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Fatalf("expected [INFO] in output, got: %s", output)
	}
	if !strings.Contains(output, "test message value") {
		t.Fatalf("expected message in output, got: %s", output)
	}
}

func TestStdLogger_Errorf(t *testing.T) {
	// Перехватываем вывод log
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := NewStdLogger()
	logger.Errorf("error message %s", "error")

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Fatalf("expected [ERROR] in output, got: %s", output)
	}
	if !strings.Contains(output, "error message error") {
		t.Fatalf("expected message in output, got: %s", output)
	}
}

func TestNewStdLogger(t *testing.T) {
	logger := NewStdLogger()
	if logger == nil {
		t.Fatalf("expected non-nil logger")
	}

	// Проверяем, что методы работают
	logger.Infof("test")
	logger.Errorf("test")
}
