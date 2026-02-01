// ABOUTME: Tests for the install-skill command
// ABOUTME: Verifies skill installation, directory creation, and file content

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// installSkillToDir installs the skill to a custom directory for testing.
// This allows us to test the installation logic without modifying the real home directory.
func installSkillToDir(homeDir string) error {
	skillDir := filepath.Join(homeDir, ".claude", "skills", "position")
	skillPath := filepath.Join(skillDir, "SKILL.md")

	// Read embedded skill file
	content, err := skillFS.ReadFile("skill/SKILL.md")
	if err != nil {
		return err
	}

	// Create directory
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}

	// Write skill file
	if err := os.WriteFile(skillPath, content, 0644); err != nil {
		return err
	}

	return nil
}

func TestSkillInstall_Success(t *testing.T) {
	tmpDir := t.TempDir()

	err := installSkillToDir(tmpDir)
	if err != nil {
		t.Fatalf("installSkillToDir failed: %v", err)
	}

	skillPath := filepath.Join(tmpDir, ".claude", "skills", "position", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Error("skill file was not created")
	}
}

func TestSkillInstall_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()

	err := installSkillToDir(tmpDir)
	if err != nil {
		t.Fatalf("installSkillToDir failed: %v", err)
	}

	// Verify all directories in the path were created
	expectedDirs := []string{
		filepath.Join(tmpDir, ".claude"),
		filepath.Join(tmpDir, ".claude", "skills"),
		filepath.Join(tmpDir, ".claude", "skills", "position"),
	}

	for _, dir := range expectedDirs {
		info, err := os.Stat(dir)
		if os.IsNotExist(err) {
			t.Errorf("directory was not created: %s", dir)
			continue
		}
		if err != nil {
			t.Errorf("error checking directory %s: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected directory, got file: %s", dir)
		}
	}
}

func TestSkillInstall_FileContent(t *testing.T) {
	tmpDir := t.TempDir()

	err := installSkillToDir(tmpDir)
	if err != nil {
		t.Fatalf("installSkillToDir failed: %v", err)
	}

	skillPath := filepath.Join(tmpDir, ".claude", "skills", "position", "SKILL.md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("failed to read skill file: %v", err)
	}

	// Check expected content markers
	contentStr := string(content)

	expectedStrings := []string{
		"name: position",
		"# position - Location Tracking",
		"mcp__position__add_position",
		"mcp__position__get_current",
		"mcp__position__get_timeline",
		"mcp__position__list_entities",
		"mcp__position__remove_entity",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("skill file missing expected content: %q", expected)
		}
	}
}

func TestSkillInstall_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()

	// First installation
	err := installSkillToDir(tmpDir)
	if err != nil {
		t.Fatalf("first installSkillToDir failed: %v", err)
	}

	skillPath := filepath.Join(tmpDir, ".claude", "skills", "position", "SKILL.md")

	// Get original file info
	origInfo, err := os.Stat(skillPath)
	if err != nil {
		t.Fatalf("failed to stat original file: %v", err)
	}
	origSize := origInfo.Size()

	// Modify the file to verify overwrite
	testContent := []byte("this is test content that should be overwritten")
	if err := os.WriteFile(skillPath, testContent, 0644); err != nil {
		t.Fatalf("failed to write test content: %v", err)
	}

	// Verify test content was written
	modContent, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("failed to read modified file: %v", err)
	}
	if string(modContent) != string(testContent) {
		t.Fatal("test content was not written correctly")
	}

	// Second installation should overwrite
	err = installSkillToDir(tmpDir)
	if err != nil {
		t.Fatalf("second installSkillToDir failed: %v", err)
	}

	// Verify file was overwritten with original content
	newInfo, err := os.Stat(skillPath)
	if err != nil {
		t.Fatalf("failed to stat new file: %v", err)
	}
	newSize := newInfo.Size()

	if newSize != origSize {
		t.Errorf("file size after overwrite: got %d, want %d", newSize, origSize)
	}

	// Verify content is the embedded skill file, not test content
	finalContent, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("failed to read final file: %v", err)
	}
	if strings.Contains(string(finalContent), "this is test content") {
		t.Error("file was not overwritten - still contains test content")
	}
	if !strings.Contains(string(finalContent), "name: position") {
		t.Error("file was not overwritten with correct skill content")
	}
}

func TestSkillInstall_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	err := installSkillToDir(tmpDir)
	if err != nil {
		t.Fatalf("installSkillToDir failed: %v", err)
	}

	skillPath := filepath.Join(tmpDir, ".claude", "skills", "position", "SKILL.md")
	info, err := os.Stat(skillPath)
	if err != nil {
		t.Fatalf("failed to stat skill file: %v", err)
	}

	// Check file permissions (should be 0644)
	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Errorf("unexpected file permissions: got %o, want 0644", perm)
	}
}

func TestSkillInstall_DirectoryPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	err := installSkillToDir(tmpDir)
	if err != nil {
		t.Fatalf("installSkillToDir failed: %v", err)
	}

	skillDir := filepath.Join(tmpDir, ".claude", "skills", "position")
	info, err := os.Stat(skillDir)
	if err != nil {
		t.Fatalf("failed to stat skill directory: %v", err)
	}

	// Check directory permissions (should be 0755)
	perm := info.Mode().Perm()
	if perm != 0755 {
		t.Errorf("unexpected directory permissions: got %o, want 0755", perm)
	}
}

func TestSkillInstall_EmbeddedFileExists(t *testing.T) {
	// Verify the embedded file can be read
	content, err := skillFS.ReadFile("skill/SKILL.md")
	if err != nil {
		t.Fatalf("failed to read embedded skill file: %v", err)
	}

	if len(content) == 0 {
		t.Error("embedded skill file is empty")
	}
}

func TestSkillSkipConfirmFlag(t *testing.T) {
	// Verify the flag exists on the command
	flag := installSkillCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("--yes flag not found on install-skill command")
	}

	if flag.Shorthand != "y" {
		t.Errorf("--yes flag shorthand: got %q, want %q", flag.Shorthand, "y")
	}

	if flag.DefValue != "false" {
		t.Errorf("--yes flag default value: got %q, want %q", flag.DefValue, "false")
	}
}

func TestInstallSkillCmd_Metadata(t *testing.T) {
	if installSkillCmd.Use != "install-skill" {
		t.Errorf("command Use: got %q, want %q", installSkillCmd.Use, "install-skill")
	}

	if installSkillCmd.Short != "Install Claude Code skill" {
		t.Errorf("command Short: got %q, want %q", installSkillCmd.Short, "Install Claude Code skill")
	}

	if !strings.Contains(installSkillCmd.Long, "~/.claude/skills/position/") {
		t.Error("command Long description should mention installation path")
	}
}

func TestInstallSkillCmd_RunE(t *testing.T) {
	// Verify that the command's RunE function exists and is callable
	if installSkillCmd.RunE == nil {
		t.Fatal("installSkillCmd.RunE should not be nil")
	}
}
