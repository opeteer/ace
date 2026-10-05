package test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/opeteer/ace/internal/doctor"
)

func TestDiagnoseSystem(t *testing.T) {
	diagnostics := doctor.DiagnoseSystem()
	if len(diagnostics) == 0 {
		t.Fatal("Expected at least one diagnostic result, got 0")
	}

	foundGit := false
	for _, d := range diagnostics {
		if d.Name == "" || d.Category == "" || d.Binary == "" {
			t.Errorf("Diagnostic item has missing required fields: %+v", d)
		}
		if d.Status != doctor.StatusOK && d.Status != doctor.StatusMissing {
			t.Errorf("Unexpected status %s for tool %s", d.Status, d.Name)
		}
		if d.Binary == "git" {
			foundGit = true
			if d.Status != doctor.StatusOK {
				t.Errorf("Expected Git to be OK on host, got %s", d.Status)
			}
			if d.Version == "" {
				t.Error("Expected Git version to be non-empty")
			}
		}
	}

	if !foundGit {
		t.Error("Git was not included in audited tools list")
	}
}

func TestDiagnoseSystemJSON(t *testing.T) {
	jsonStr, err := doctor.DiagnoseSystemJSON()
	if err != nil {
		t.Fatalf("DiagnoseSystemJSON failed: %v", err)
	}

	var parsed []doctor.ToolDiagnostic
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if len(parsed) == 0 {
		t.Error("Expected parsed diagnostics JSON to have items")
	}
}

func TestRenderReport(t *testing.T) {
	diagnostics := doctor.DiagnoseSystem()
	rendered := doctor.RenderReport(diagnostics)

	if !strings.Contains(rendered, "Ace Toolchain & Environment Readiness Doctor") {
		t.Error("Expected report to contain header title")
	}
	if !strings.Contains(rendered, "Audit Summary:") {
		t.Error("Expected report to contain Audit Summary")
	}
	if !strings.Contains(rendered, "Pro-Tip:") {
		t.Error("Expected report to contain Pro-Tip")
	}
}
