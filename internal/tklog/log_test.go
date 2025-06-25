package tklog

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func resetLogger() {
	logger = nil
}

func TestEnsureLoggerCreatesFile(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(dir)
	resetLogger()

	l := ensureLogger()
	if l == nil {
		t.Fatalf("expected logger not nil")
	}
	if _, err := os.Stat("tked.log"); err != nil {
		t.Fatalf("expected log file created: %v", err)
	}
}

func TestLoggingFunctionsWrite(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(dir)
	resetLogger()

	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")

	data, err := os.ReadFile("tked.log")
	if err != nil {
		t.Fatalf("unexpected error reading log: %v", err)
	}
	out := string(data)
	if !strings.Contains(out, "DBG: debug") {
		t.Fatalf("debug message not found: %q", out)
	}
	if !strings.Contains(out, "INF: info") {
		t.Fatalf("info message not found: %q", out)
	}
	if !strings.Contains(out, "WRN: warn") {
		t.Fatalf("warn message not found: %q", out)
	}
	if !strings.Contains(out, "ERR: error") {
		t.Fatalf("error message not found: %q", out)
	}
}

func TestPanicLogsAndPanics(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(dir)
	resetLogger()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
		data, err := os.ReadFile("tked.log")
		if err != nil {
			t.Fatalf("unexpected error reading log: %v", err)
		}
		if !strings.Contains(string(data), "PAN: boom") {
			t.Fatalf("log message not found")
		}
	}()

	Panic("boom")
}

func TestFatalLogsAndExits(t *testing.T) {
	if os.Getenv("TKLOG_FATAL") == "1" {
		resetLogger()
		Fatal("goodbye")
		return
	}

	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestFatalLogsAndExits$")
	cmd.Env = append(os.Environ(), "TKLOG_FATAL=1")
	cmd.Dir = dir
	err := cmd.Run()
	if err == nil {
		t.Fatalf("process succeeded, want exit 1")
	}
	if ee, ok := err.(*exec.ExitError); ok {
		if ee.ExitCode() != 1 {
			t.Fatalf("expected exit code 1 got %d", ee.ExitCode())
		}
	} else {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "tked.log"))
	if err != nil {
		t.Fatalf("unexpected error reading log: %v", err)
	}
	if !strings.Contains(string(data), "FTL: goodbye") {
		t.Fatalf("log message not found")
	}
}
