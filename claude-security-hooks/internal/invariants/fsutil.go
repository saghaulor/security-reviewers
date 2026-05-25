package invariants

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// errPathEscapes is returned by guarded reads when the input path escapes the
// workspace via "..", absolute prefix, or symlink. Callers convert to Violation.
var errPathEscapes = errors.New("path_escapes_workspace")

func cleanWorkspaceRelative(path string) (string, error) {
	if filepath.IsAbs(path) {
		return "", errPathEscapes
	}
	cleaned := filepath.Clean(path)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", errPathEscapes
	}
	return cleaned, nil
}

// FileExists reports whether path resolves to an existing regular file.
// Returns false for directories, missing files, traversal escapes, and symlinks.
// Used by A7 / S1.
func FileExists(path string) bool {
	cleaned, err := cleanWorkspaceRelative(path)
	if err != nil {
		return false
	}
	info, err := os.Lstat(cleaned)
	if err != nil {
		return false
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false
	}
	return info.Mode().IsRegular()
}

// DirExists reports whether path resolves to an existing directory.
func DirExists(path string) bool {
	cleaned, err := cleanWorkspaceRelative(path)
	if err != nil {
		return false
	}
	info, err := os.Lstat(cleaned)
	if err != nil {
		return false
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false
	}
	return info.IsDir()
}

// LineCount returns the number of newline-delimited lines in the file at path.
// Streams via bufio.Scanner; rejects traversal and symlinks. Used by A8 / T9 / AZ4 / OA4 / IC3.
func LineCount(path string) (int, error) {
	cleaned, err := cleanWorkspaceRelative(path)
	if err != nil {
		return 0, err
	}
	info, err := os.Lstat(cleaned)
	if err != nil {
		return 0, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return 0, errPathEscapes
	}
	f, err := os.Open(cleaned)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	n := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		n++
	}
	return n, sc.Err()
}

// ResolveWorkspaceRoot consults $CLAUDE_PROJECT_DIR first (per RESEARCH RQ-8),
// then the process CWD, accepting the first directory that contains one of the
// workspace markers go.mod / .claude / .planning. Returns "" if none qualify;
// callers treat "" as block reason "workspace_root_not_found" (D-11).
func ResolveWorkspaceRoot() string {
	if dir := os.Getenv("CLAUDE_PROJECT_DIR"); dir != "" && hasWorkspaceMarkers(dir) {
		return dir
	}
	if cwd, err := os.Getwd(); err == nil && hasWorkspaceMarkers(cwd) {
		return cwd
	}
	return ""
}

func hasWorkspaceMarkers(dir string) bool {
	for _, marker := range []string{"go.mod", ".claude", ".planning"} {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}
