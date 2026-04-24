package file

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestFsWalkerSingleFile(t *testing.T) {
	var visited []string
	err := fsWalker(fakeMalwareSample, false, func(path string, d os.DirEntry) {
		visited = append(visited, d.Name())
	})
	if err != nil {
		t.Fatalf("fsWalker failed: %s", err)
	}
	if len(visited) != 1 {
		t.Fatalf("expected 1 file, got %d", len(visited))
	}
	if visited[0] != "sample" {
		t.Fatalf("expected sample, got %s", visited[0])
	}
}

func TestFsWalkerDirectory(t *testing.T) {
	dir := t.TempDir()
	// create some files
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0600); err != nil {
			t.Fatalf("failed to create file: %s", err)
		}
	}
	// create a subdirectory with a file
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatalf("failed to create dir: %s", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "d.txt"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create file: %s", err)
	}

	var visited []string
	err := fsWalker(dir, false, func(path string, d os.DirEntry) {
		visited = append(visited, d.Name())
	})
	if err != nil {
		t.Fatalf("fsWalker failed: %s", err)
	}
	if len(visited) != 4 {
		t.Fatalf("expected 4 files, got %d: %v", len(visited), visited)
	}
}

func TestFsWalkerSkipsDirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create file: %s", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatalf("failed to create dir: %s", err)
	}

	var visited []string
	err := fsWalker(dir, false, func(path string, d os.DirEntry) {
		visited = append(visited, d.Name())
		if d.IsDir() {
			t.Fatalf("handler should not be called with directory, got %s", d.Name())
		}
	})
	if err != nil {
		t.Fatalf("fsWalker failed: %s", err)
	}
	if len(visited) != 1 {
		t.Fatalf("expected 1 file, got %d", len(visited))
	}
}

func TestFsWalkerNonExistentPath(t *testing.T) {
	err := fsWalker("/tmp/does_not_exist_at_all", false, func(path string, d os.DirEntry) {
		t.Fatal("should not be called")
	})
	if err == nil {
		t.Fatal("expected error for non-existent path")
	}
}

func TestFsWalkerPassesFilePath(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatalf("failed to create dir: %s", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "deep.txt"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create file: %s", err)
	}

	var paths []string
	err := fsWalker(dir, false, func(path string, d os.DirEntry) {
		paths = append(paths, path)
	})
	if err != nil {
		t.Fatalf("fsWalker failed: %s", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d: %v", len(paths), paths)
	}
	expected := filepath.Join(sub, "deep.txt")
	if paths[0] != expected {
		t.Fatalf("expected path %q, got %q", expected, paths[0])
	}
}

func TestFsWalkerIgnoreHidden(t *testing.T) {
	dir := t.TempDir()
	// visible file
	if err := os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create file: %s", err)
	}
	// hidden file
	if err := os.WriteFile(filepath.Join(dir, ".hidden"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create file: %s", err)
	}
	// hidden directory with a file inside
	hiddenDir := filepath.Join(dir, ".hiddendir")
	if err := os.Mkdir(hiddenDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %s", err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "inside.txt"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create file: %s", err)
	}

	var visited []string
	err := fsWalker(dir, true, func(path string, d os.DirEntry) {
		visited = append(visited, d.Name())
	})
	if err != nil {
		t.Fatalf("fsWalker failed: %s", err)
	}
	if len(visited) != 1 {
		t.Fatalf("expected 1 file (only visible.txt), got %d: %v", len(visited), visited)
	}
	if visited[0] != "visible.txt" {
		t.Fatalf("expected visible.txt, got %s", visited[0])
	}
}

func TestFsWalkerIgnoreHiddenDotfiles(t *testing.T) {
	dir := t.TempDir()
	// create common dotfiles that should be ignored
	for _, name := range []string{".gitignore", ".env", ".dockerignore"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0600); err != nil {
			t.Fatalf("failed to create file: %s", err)
		}
	}
	// create a visible file
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create file: %s", err)
	}
	// create nested dotfile inside a subdirectory
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatalf("failed to create dir: %s", err)
	}
	if err := os.WriteFile(filepath.Join(sub, ".gitignore"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create file: %s", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "readme.md"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create file: %s", err)
	}

	var visited []string
	err := fsWalker(dir, true, func(path string, d os.DirEntry) {
		visited = append(visited, d.Name())
	})
	if err != nil {
		t.Fatalf("fsWalker failed: %s", err)
	}
	sort.Strings(visited)
	if len(visited) != 2 {
		t.Fatalf("expected 2 files (main.go, readme.md), got %d: %v", len(visited), visited)
	}
	if visited[0] != "main.go" || visited[1] != "readme.md" {
		t.Fatalf("expected [main.go, readme.md], got %v", visited)
	}
}

func TestFsWalkerIgnoreHiddenSingleFile(t *testing.T) {
	dir := t.TempDir()
	hiddenFile := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(hiddenFile, []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create file: %s", err)
	}

	var visited []string
	err := fsWalker(hiddenFile, true, func(path string, d os.DirEntry) {
		visited = append(visited, d.Name())
	})
	if err != nil {
		t.Fatalf("fsWalker failed: %s", err)
	}
	if len(visited) != 0 {
		t.Fatalf("expected 0 files when ignoring hidden single file, got %d: %v", len(visited), visited)
	}
}

func TestFsWalkerIncludeHidden(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create file: %s", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".hidden"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create file: %s", err)
	}

	var visited []string
	err := fsWalker(dir, false, func(path string, d os.DirEntry) {
		visited = append(visited, d.Name())
	})
	if err != nil {
		t.Fatalf("fsWalker failed: %s", err)
	}
	sort.Strings(visited)
	if len(visited) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(visited), visited)
	}
}
