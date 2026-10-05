package test

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/opeteer/ace/internal/generator/installer"
)

func TestRunWithIdleTimeoutSuccess(t *testing.T) {
	cmd := exec.Command("echo", "hello world")
	err := installer.RunWithIdleTimeout(cmd, 2*time.Second)
	if err != nil {
		t.Fatalf("Expected fast echo command to succeed, got: %v", err)
	}
}

func TestRunWithIdleTimeoutStreamingContinues(t *testing.T) {
	// Outputs a line every 100ms for 600ms total.
	// Idle timeout is 300ms. Because data arrives every 100ms, the command should NOT time out!
	script := "for i in 1 2 3 4 5; do echo tick $i; sleep 0.1; done"
	cmd := exec.Command("bash", "-c", script)
	err := installer.RunWithIdleTimeout(cmd, 300*time.Millisecond)
	if err != nil {
		t.Fatalf("Expected streaming command to succeed as data was actively arriving, got: %v", err)
	}
}

func TestRunWithIdleTimeoutStalledCommandKilled(t *testing.T) {
	// A silent command that does not emit any data for 1.5 seconds.
	// Idle timeout is 300ms. It MUST be terminated due to no incoming data!
	cmd := exec.Command("sleep", "1.5")
	err := installer.RunWithIdleTimeout(cmd, 300*time.Millisecond)
	if err == nil {
		t.Fatal("Expected stalled command to return idle timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "stalled") && !strings.Contains(err.Error(), "no data received") {
		t.Errorf("Expected stalled/no data error message, got: %v", err)
	}
}
