// ABOUTME: Integration tests for full workflow
// ABOUTME: Tests CLI commands end-to-end

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFullWorkflow(t *testing.T) {
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	binary := filepath.Join(projectRoot, "position")
	buildCmd := exec.Command("go", "build", "-o", binary, "./cmd/position")
	buildCmd.Dir = projectRoot
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build: %v\nOutput: %s", err, buildOutput)
	}
	defer func() { _ = os.Remove(binary) }()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	run := func(args ...string) (string, error) {
		fullArgs := append([]string{"--db", dbPath}, args...)
		cmd := exec.Command(binary, fullArgs...)
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	// Add a position
	output, err := run("add", "--label", "chicago", "harper", "41.8781", "-87.6298")
	if err != nil {
		t.Fatalf("Failed to add: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Added position") {
		t.Error("Expected success message")
	}

	// Get current
	output, err = run("current", "harper")
	if err != nil {
		t.Fatalf("Failed to get current: %v\n%s", err, output)
	}
	if !strings.Contains(output, "chicago") {
		t.Error("Expected chicago in output")
	}

	// Add another position
	_, err = run("add", "--label", "new york", "harper", "40.7128", "-74.0060")
	if err != nil {
		t.Fatalf("Failed to add second position: %v", err)
	}

	// Timeline should show both
	output, err = run("timeline", "harper")
	if err != nil {
		t.Fatalf("Failed to get timeline: %v\n%s", err, output)
	}
	if !strings.Contains(output, "new york") || !strings.Contains(output, "chicago") {
		t.Error("Expected both locations in timeline")
	}

	// List should show harper
	output, err = run("list")
	if err != nil {
		t.Fatalf("Failed to list: %v\n%s", err, output)
	}
	if !strings.Contains(output, "harper") {
		t.Error("Expected harper in list")
	}

	// Remove
	output, err = run("remove", "harper", "--confirm")
	if err != nil {
		t.Fatalf("Failed to remove: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Removed") {
		t.Error("Expected removal confirmation")
	}

	// List should be empty
	output, err = run("list")
	if err != nil {
		t.Fatalf("Failed to list: %v\n%s", err, output)
	}
	if strings.Contains(output, "harper") {
		t.Error("harper should be removed")
	}

	t.Log("Integration test passed!")
}
