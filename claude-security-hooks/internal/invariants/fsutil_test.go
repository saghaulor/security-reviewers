package invariants_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saghaulor/claude-security-hooks/internal/invariants"
)

func TestFileExists_RegularFile(t *testing.T) {
	path := filepath.Join("testdata", "workspace", "handler.go")
	if !invariants.FileExists(path) {
		t.Errorf("FileExists(%q) = false, want true", path)
	}
}

func TestFileExists_MissingFile(t *testing.T) {
	path := filepath.Join("testdata", "workspace", "missing.go")
	if invariants.FileExists(path) {
		t.Errorf("FileExists(%q) = true, want false", path)
	}
}

func TestFileExists_Directory(t *testing.T) {
	path := filepath.Join("testdata", "workspace")
	if invariants.FileExists(path) {
		t.Errorf("FileExists(%q) = true, want false (directories are not files)", path)
	}
}

func TestDirExists_ExistingDir(t *testing.T) {
	path := filepath.Join("testdata", "workspace")
	if !invariants.DirExists(path) {
		t.Errorf("DirExists(%q) = false, want true", path)
	}
}

func TestDirExists_MissingDir(t *testing.T) {
	path := filepath.Join("testdata", "workspace", "nope")
	if invariants.DirExists(path) {
		t.Errorf("DirExists(%q) = true, want false", path)
	}
}

func TestLineCount_KnownCounts(t *testing.T) {
	tests := []struct {
		file  string
		lines int
	}{
		{filepath.Join("testdata", "workspace", "handler.go"), 27},
		{filepath.Join("testdata", "workspace", "service.go"), 11},
		{filepath.Join("testdata", "workspace", "empty.go"), 0},
		{filepath.Join("testdata", "workspace", "noeol.go"), 1},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			n, err := invariants.LineCount(tt.file)
			if err != nil {
				t.Fatalf("LineCount(%q) error = %v, want nil", tt.file, err)
			}
			if n != tt.lines {
				t.Errorf("LineCount(%q) = %d, want %d", tt.file, n, tt.lines)
			}
		})
	}
}

func TestLineCount_MissingFile(t *testing.T) {
	path := filepath.Join("testdata", "workspace", "missing.go")
	_, err := invariants.LineCount(path)
	if err == nil {
		t.Errorf("LineCount(%q) error = nil, want non-nil", path)
	}
}

func TestLineCount_RejectsSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.go")
	if err := os.WriteFile(targetFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	linkFile := filepath.Join(tmpDir, "link.go")
	if err := os.Symlink(targetFile, linkFile); err != nil {
		t.Fatalf("Symlink failed: %v", err)
	}

	// Change to tmpDir so the relative path works
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	_, err := invariants.LineCount("link.go")
	if err == nil {
		t.Errorf("LineCount(symlink) error = nil, want non-nil")
	}
}

func TestLineCount_RejectsTraversal(t *testing.T) {
	tests := []struct {
		path string
		name string
	}{
		{"../../etc/passwd", "leading traversal"},
		{"/etc/passwd", "absolute path"},
		{"a/b/../../../c.go", "midpoint traversal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := invariants.LineCount(tt.path)
			if err == nil {
				t.Errorf("LineCount(%q) error = nil, want non-nil", tt.path)
			}
		})
	}
}

func TestResolveWorkspaceRoot_EnvSet(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a go.mod marker
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	t.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	root := invariants.ResolveWorkspaceRoot()
	if root != tmpDir {
		t.Errorf("ResolveWorkspaceRoot() = %q, want %q", root, tmpDir)
	}
}

func TestResolveWorkspaceRoot_EnvSetButMissingMarkers(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create any markers

	t.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	root := invariants.ResolveWorkspaceRoot()
	// Should fall back to CWD. Since test may be run from various locations,
	// we just check that it doesn't return the tmpDir.
	if root == tmpDir {
		t.Errorf("ResolveWorkspaceRoot() = %q (tmpDir), want fallback or empty", tmpDir)
	}
}

func TestResolveWorkspaceRoot_EnvUnsetCWDValid(t *testing.T) {
	// Save original env
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	// Use actual project root for this test
	os.Chdir("/home/saghaulor/code/security_reviewer")
	t.Setenv("CLAUDE_PROJECT_DIR", "")

	root := invariants.ResolveWorkspaceRoot()
	if root == "" {
		t.Errorf("ResolveWorkspaceRoot() = empty, want non-empty (project has markers)")
	}
}

func TestResolveWorkspaceRoot_EnvUnsetCWDInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	os.Chdir(tmpDir)
	t.Setenv("CLAUDE_PROJECT_DIR", "")

	root := invariants.ResolveWorkspaceRoot()
	if root != "" {
		t.Errorf("ResolveWorkspaceRoot() = %q, want empty", root)
	}
}
