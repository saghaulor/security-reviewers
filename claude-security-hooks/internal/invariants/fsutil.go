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

// fsRoot is the base directory that workspace-relative paths are resolved against.
// When empty, paths resolve against the process working directory (the default,
// which preserves CWD-relative behavior for direct callers and tests).
//
// validate.Validate sets this once per invocation via SetFSRoot instead of calling
// os.Chdir. os.Chdir mutates process-global working-directory state that affects
// every goroutine and every relative file operation in the process; resolving
// against an explicit base avoids that blast radius (WR-05). The hooks binary
// processes a single event per invocation, so a package-scoped base is sufficient.
var fsRoot string

// SetFSRoot sets the base directory used to resolve workspace-relative paths in
// FileExists, DirExists, LineCount, and ResolveWorkspacePath. Pass "" to restore
// CWD-relative resolution.
func SetFSRoot(root string) { fsRoot = root }

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

// ResolveWorkspacePath cleans a workspace-relative path (rejecting absolute paths
// and ".." traversal) and joins it onto the configured fsRoot. Callers that read
// files directly — rather than through FileExists/DirExists/LineCount — use this so
// their reads honor fsRoot too.
func ResolveWorkspacePath(path string) (string, error) {
	cleaned, err := cleanWorkspaceRelative(path)
	if err != nil {
		return "", err
	}
	if fsRoot == "" {
		return cleaned, nil
	}
	return filepath.Join(fsRoot, cleaned), nil
}

// FileExists reports whether path resolves to an existing regular file.
// Returns false for directories, missing files, traversal escapes, and symlinks.
// Used by A7 / S1.
func FileExists(path string) bool {
	resolved, err := ResolveWorkspacePath(path)
	if err != nil {
		return false
	}
	info, err := os.Lstat(resolved)
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
	resolved, err := ResolveWorkspacePath(path)
	if err != nil {
		return false
	}
	info, err := os.Lstat(resolved)
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
	resolved, err := ResolveWorkspacePath(path)
	if err != nil {
		return 0, err
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return 0, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return 0, errPathEscapes
	}
	f, err := os.Open(resolved)
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
