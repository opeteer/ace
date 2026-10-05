package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opeteer/ace/internal/doctor"
)

func TestProvisionerRecipes(t *testing.T) {
	expectedTools := []string{"composer", "bun", "pnpm", "rust", "cargo", "dotnet"}

	for _, tool := range expectedTools {
		if !doctor.IsProvisionable(tool) {
			t.Errorf("Expected tool '%s' to be provisionable", tool)
		}
	}

	if doctor.IsProvisionable("non-existent-tool-xyz") {
		t.Error("Expected fake tool to NOT be provisionable")
	}

	tools := doctor.GetAvailableToolNames()
	if len(tools) == 0 {
		t.Fatal("Expected available tool names to be non-empty")
	}
}

func TestGetUserLocalBin(t *testing.T) {
	localBin, err := doctor.GetUserLocalBin()
	if err != nil {
		t.Fatalf("GetUserLocalBin failed: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".local", "bin")
	if localBin != expected {
		t.Errorf("Expected localBin '%s', got '%s'", expected, localBin)
	}

	stat, err := os.Stat(localBin)
	if err != nil || !stat.IsDir() {
		t.Errorf("Expected localBin '%s' to exist as a directory", localBin)
	}
}
