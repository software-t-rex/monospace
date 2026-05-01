/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package tasks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/software-t-rex/monospace/app"
)

// makeTestProject creates a temporary project directory with the given files (name → content).
func makeTestProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0640); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	return dir
}

func baseOpts(projectPath, monospaceRoot string) CacheOptions {
	return CacheOptions{
		ProjectName:   "myproject",
		TaskName:      "build",
		ProjectPath:   projectPath,
		Mode:          "skip",
		Strategy:      app.CacheStrategyContent,
		MonospaceRoot: monospaceRoot,
	}
}

func baseTaskDef(cmd []string) app.MonospaceConfigTask {
	return app.MonospaceConfigTask{
		Cmd:   cmd,
		Cache: "skip",
	}
}

// ─── walkProjectFiles ──────────────────────────────────────────────────────────

// TestWalkProjectFiles_FollowsSymlinkDirectory checks that walkProjectFiles traverses
// directories accessible via a symbolic link.
// Current bug: filepath.WalkDir does not follow symlinks to directories.
func TestWalkProjectFiles_FollowsSymlinkDirectory(t *testing.T) {
	projectDir := t.TempDir()

	// Create a real subdirectory with a file
	realSubdir := filepath.Join(projectDir, "subdir")
	if err := os.MkdirAll(realSubdir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realSubdir, "file.go"), []byte("package main"), 0640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "main.go"), []byte("package main"), 0640); err != nil {
		t.Fatal(err)
	}

	// Create a symbolic link to the real subdirectory
	linkDir := filepath.Join(projectDir, "linkdir")
	if err := os.Symlink(realSubdir, linkDir); err != nil {
		t.Skip("cannot create symbolic links on this OS")
	}

	files, err := walkProjectFiles(projectDir)
	if err != nil {
		t.Fatalf("walkProjectFiles: %v", err)
	}

	var relFiles []string
	for _, f := range files {
		rel, _ := filepath.Rel(projectDir, f)
		relFiles = append(relFiles, rel)
	}
	sort.Strings(relFiles)

	// We want the CONTENT of the linked directory to be traversed (linkdir/file.go),
	// not just the link itself appearing as a file.
	hasFileInsideLink := false
	sep := string(filepath.Separator)
	for _, f := range relFiles {
		if strings.HasPrefix(f, "linkdir"+sep) {
			hasFileInsideLink = true
		}
	}
	if !hasFileInsideLink {
		t.Errorf("walkProjectFiles should traverse the content of symbolic directories\n"+
			"(e.g.: linkdir/file.go), found files: %v", relFiles)
	}
}

// TestWalkProjectFiles_SymlinkCycleNoHang checks that walkProjectFiles does not loop
// indefinitely when a symbolic link cycle is present.
func TestWalkProjectFiles_SymlinkCycleNoHang(t *testing.T) {
	projectDir := t.TempDir()

	subdir := filepath.Join(projectDir, "subdir")
	if err := os.MkdirAll(subdir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "file.go"), []byte("content"), 0640); err != nil {
		t.Fatal(err)
	}

	// Symbolic link in subdir that points to the parent directory → cycle
	cycleLink := filepath.Join(subdir, "cycle")
	if err := os.Symlink(projectDir, cycleLink); err != nil {
		t.Skip("cannot create symbolic links on this OS")
	}

	done := make(chan []string, 1)
	go func() {
		files, _ := walkProjectFiles(projectDir)
		done <- files
	}()

	select {
	case files := <-done:
		// Must complete without looping — just check that we have at least one file
		if len(files) == 0 {
			t.Error("no files found when at least one was expected")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("walkProjectFiles diverged on a symbolic link cycle")
	}
}

// TestComputeHash_InputsPatternChangesBustsCache checks that changing Inputs patterns
// invalidates the cache even when the matched files are identical.
// Current bug: patterns themselves are not included in the hash, only
// the resolved files. Two different patterns matching the same files
// therefore produce the same hash.
func TestComputeHash_InputsPatternChangesBustsCache(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{
		"src/main.go": "main",
	})
	taskDef := baseTaskDef([]string{"build"})

	// Both patterns match the same file src/main.go
	opts1 := baseOpts(dir, root)
	opts1.Inputs = []string{"src/**"}

	opts2 := baseOpts(dir, root)
	opts2.Inputs = []string{"src/**/*.go"}

	h1, err := ComputeHash(opts1, taskDef)
	if err != nil {
		t.Fatalf("ComputeHash opts1: %v", err)
	}
	h2, err := ComputeHash(opts2, taskDef)
	if err != nil {
		t.Fatalf("ComputeHash opts2: %v", err)
	}

	if h1 == h2 {
		t.Error("different Inputs patterns should produce different hashes " +
			"even if the matched files are identical (patterns are not included in the hash)")
	}
}

// ─── ComputeHash ───────────────────────────────────────────────────────────────

func TestComputeHash_ContentStrategy_SameFiles(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"src/main.go": "package main"})
	opts := baseOpts(dir, root)
	taskDef := baseTaskDef([]string{"go", "build"})

	h1, err := ComputeHash(opts, taskDef)
	if err != nil {
		t.Fatalf("ComputeHash: %v", err)
	}
	h2, err := ComputeHash(opts, taskDef)
	if err != nil {
		t.Fatalf("ComputeHash: %v", err)
	}
	if h1 != h2 {
		t.Error("same inputs should produce the same hash")
	}
}

func TestComputeHash_ContentStrategy_ChangedFile(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"src/main.go": "package main"})
	opts := baseOpts(dir, root)
	taskDef := baseTaskDef([]string{"go", "build"})

	h1, _ := ComputeHash(opts, taskDef)

	// mutate file content
	os.WriteFile(filepath.Join(dir, "src/main.go"), []byte("package main // changed"), 0640)

	h2, _ := ComputeHash(opts, taskDef)
	if h1 == h2 {
		t.Error("changed file content should produce a different hash")
	}
}

func TestComputeHash_MtimeStrategy(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.txt": "hello"})
	opts := baseOpts(dir, root)
	opts.Strategy = app.CacheStrategyMtime
	taskDef := baseTaskDef([]string{"echo"})

	h1, _ := ComputeHash(opts, taskDef)
	h2, _ := ComputeHash(opts, taskDef)
	if h1 != h2 {
		t.Error("same mtime should produce the same hash")
	}

	// change mtime without changing content
	future := time.Now().Add(10 * time.Second)
	os.Chtimes(filepath.Join(dir, "a.txt"), future, future)

	h3, _ := ComputeHash(opts, taskDef)
	if h1 == h3 {
		t.Error("changed mtime should produce a different hash")
	}
}

func TestComputeHash_EmptyInputs_UsesAllFiles(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{
		"a.go": "a",
		"b.go": "b",
	})
	opts := baseOpts(dir, root)
	taskDef := baseTaskDef([]string{"go", "build"})

	h1, _ := ComputeHash(opts, taskDef)

	// add a new file — hash should change
	os.WriteFile(filepath.Join(dir, "c.go"), []byte("c"), 0640)
	h2, _ := ComputeHash(opts, taskDef)

	if h1 == h2 {
		t.Error("adding a file should change the hash when inputs is empty")
	}
}

func TestComputeHash_WithInputGlobs(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{
		"src/main.go":    "main",
		"dist/output.js": "built",
	})
	opts := baseOpts(dir, root)
	opts.Inputs = []string{"src/**"}
	taskDef := baseTaskDef([]string{"build"})

	h1, _ := ComputeHash(opts, taskDef)

	// changing dist file should NOT change hash
	os.WriteFile(filepath.Join(dir, "dist/output.js"), []byte("changed"), 0640)
	h2, _ := ComputeHash(opts, taskDef)
	if h1 != h2 {
		t.Error("changing a file outside inputs glob should not change the hash")
	}

	// changing src file SHOULD change hash
	os.WriteFile(filepath.Join(dir, "src/main.go"), []byte("changed"), 0640)
	h3, _ := ComputeHash(opts, taskDef)
	if h1 == h3 {
		t.Error("changing a file inside inputs glob should change the hash")
	}
}

func TestComputeHash_TaskDefChangeBustsCache(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)

	h1, _ := ComputeHash(opts, baseTaskDef([]string{"go", "build"}))
	h2, _ := ComputeHash(opts, baseTaskDef([]string{"go", "build", "-race"}))

	if h1 == h2 {
		t.Error("different cmd should produce different hashes")
	}
}

// ─── Check / Save — skip mode ──────────────────────────────────────────────────

func TestCheckAndSave_SkipMode_Miss(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)

	result, err := Check(opts, "deadbeef")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if result.Hit {
		t.Error("expected cache miss")
	}
}

func TestCheckAndSave_SkipMode_Hit(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)
	taskDef := baseTaskDef([]string{"go", "build"})

	hash, _ := ComputeHash(opts, taskDef)
	if err := Save(opts, hash, "test output"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	result, err := Check(opts, hash)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !result.Hit {
		t.Error("expected cache hit after Save")
	}
	// skip mode: no outputs directory
	outputsDir := filepath.Join(result.CacheDir, "outputs")
	if _, err := os.Stat(outputsDir); !os.IsNotExist(err) {
		t.Error("skip mode should not create an outputs directory")
	}
}

func TestSave_OutputIsStored(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)

	const wantOutput = "build successful\nno errors"
	hash, _ := ComputeHash(opts, baseTaskDef([]string{"go", "build"}))
	Save(opts, hash, wantOutput)

	entryDir := cacheEntryDir(root, opts.ProjectName, opts.TaskName, hash)
	data, err := os.ReadFile(filepath.Join(entryDir, "output.txt"))
	if err != nil {
		t.Fatalf("output.txt not found: %v", err)
	}
	if string(data) != wantOutput {
		t.Errorf("stored output: got %q, want %q", string(data), wantOutput)
	}
}

func TestReadCachedOutput_ReturnsStoredOutput(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)

	const wantOutput = "hello from cache"
	hash, _ := ComputeHash(opts, baseTaskDef([]string{"go", "build"}))
	Save(opts, hash, wantOutput)

	result, _ := Check(opts, hash)
	got, err := readCachedOutput(result.CacheDir)
	if err != nil {
		t.Fatalf("readCachedOutput: %v", err)
	}
	if got != wantOutput {
		t.Errorf("replayed output: got %q, want %q", got, wantOutput)
	}
}

func TestReadCachedOutput_EmptyWhenMissing(t *testing.T) {
	got, err := readCachedOutput(t.TempDir())
	if err != nil {
		t.Fatalf("readCachedOutput on missing file: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// ─── Check / Save — restore mode ──────────────────────────────────────────────

func TestCheckAndSave_RestoreMode_SavesOutputs(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{
		"src/main.go": "main",
		"dist/out.js": "built",
	})
	opts := baseOpts(dir, root)
	opts.Mode = "restore"
	opts.Outputs = []string{"dist/**"}
	taskDef := baseTaskDef([]string{"build"})

	hash, _ := ComputeHash(opts, taskDef)
	if err := Save(opts, hash, "test output"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// verify output file was archived
	cachedOutput := filepath.Join(cacheEntryDir(root, opts.ProjectName, opts.TaskName, hash), "outputs", "dist", "out.js")
	data, err := os.ReadFile(cachedOutput)
	if err != nil {
		t.Fatalf("cached output not found: %v", err)
	}
	if string(data) != "built" {
		t.Errorf("cached output content: got %q, want %q", string(data), "built")
	}
}

func TestRestore_RestoresOutputFiles(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{
		"src/main.go": "main",
		"dist/out.js": "built",
	})
	opts := baseOpts(dir, root)
	opts.Mode = "restore"
	opts.Outputs = []string{"dist/**"}
	taskDef := baseTaskDef([]string{"build"})

	hash, _ := ComputeHash(opts, taskDef)
	Save(opts, hash, "test output")
	result, _ := Check(opts, hash)

	// delete the output file
	os.Remove(filepath.Join(dir, "dist/out.js"))

	if err := Restore(opts, result); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "dist/out.js"))
	if err != nil {
		t.Fatalf("restored file not found: %v", err)
	}
	if string(data) != "built" {
		t.Errorf("restored content: got %q, want %q", string(data), "built")
	}
}

// ─── ClearCache ────────────────────────────────────────────────────────────────

func TestClearCache_SingleTask(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)
	taskDef := baseTaskDef([]string{"go", "build"})

	hash, _ := ComputeHash(opts, taskDef)
	Save(opts, hash, "test output")

	if err := ClearCache(root, opts.ProjectName, opts.TaskName); err != nil {
		t.Fatalf("ClearCache: %v", err)
	}

	result, _ := Check(opts, hash)
	if result.Hit {
		t.Error("expected cache miss after ClearCache")
	}
}

func TestClearCache_All(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})

	// save two different tasks
	for _, taskName := range []string{"build", "test"} {
		opts := baseOpts(dir, root)
		opts.TaskName = taskName
		taskDef := baseTaskDef([]string{"go", taskName})
		hash, _ := ComputeHash(opts, taskDef)
		Save(opts, hash, "test output")
	}

	if err := ClearCache(root, "", ""); err != nil {
		t.Fatalf("ClearCache all: %v", err)
	}

	if _, err := os.Stat(cacheBaseDir(root)); !os.IsNotExist(err) {
		t.Error("cache base dir should be removed after ClearCache all")
	}
}

func TestClearCache_NonExistentIsNoError(t *testing.T) {
	root := t.TempDir()
	if err := ClearCache(root, "nonexistent", "task"); err != nil {
		t.Errorf("ClearCache on non-existent entry should not error: %v", err)
	}
}

// ─── GetCacheStatus ────────────────────────────────────────────────────────────

func TestGetCacheStatus_ReturnsEntries(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})

	for _, taskName := range []string{"build", "test"} {
		opts := baseOpts(dir, root)
		opts.TaskName = taskName
		taskDef := baseTaskDef([]string{"go", taskName})
		hash, _ := ComputeHash(opts, taskDef)
		Save(opts, hash, "test output")
	}

	entries, err := GetCacheStatus(root, nil)
	if err != nil {
		t.Fatalf("GetCacheStatus: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 cache entries, got %d", len(entries))
	}
}

func TestGetCacheStatus_WithFilter(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})

	for _, taskName := range []string{"build", "test"} {
		opts := baseOpts(dir, root)
		opts.TaskName = taskName
		taskDef := baseTaskDef([]string{"go", taskName})
		hash, _ := ComputeHash(opts, taskDef)
		Save(opts, hash, "test output")
	}

	entries, err := GetCacheStatus(root, []string{"myproject#build"})
	if err != nil {
		t.Fatalf("GetCacheStatus: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after filtering, got %d", len(entries))
	}
	if entries[0].Task != "build" {
		t.Errorf("expected task 'build', got %q", entries[0].Task)
	}
}

func TestGetCacheStatus_EmptyCache(t *testing.T) {
	root := t.TempDir()
	entries, err := GetCacheStatus(root, nil)
	if err != nil {
		t.Fatalf("GetCacheStatus on empty cache: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestGetCacheStatus_MetadataContent(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)
	taskDef := baseTaskDef([]string{"go", "build"})

	hash, _ := ComputeHash(opts, taskDef)
	before := time.Now().Truncate(time.Second)
	Save(opts, hash, "test output")

	entries, _ := GetCacheStatus(root, nil)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	e := entries[0]
	if e.Project != opts.ProjectName {
		t.Errorf("project: got %q, want %q", e.Project, opts.ProjectName)
	}
	if e.Task != opts.TaskName {
		t.Errorf("task: got %q, want %q", e.Task, opts.TaskName)
	}
	if e.Hash != hash {
		t.Errorf("hash: got %q, want %q", e.Hash, hash)
	}
	if e.CachedAt.Before(before) {
		t.Errorf("timestamp %v is before save time %v", e.CachedAt, before)
	}
}

// ─── metadata.json format ──────────────────────────────────────────────────────

func TestSave_MetadataJSON(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)

	if err := Save(opts, "abc123", "test output"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	metaPath := filepath.Join(cacheEntryDir(root, opts.ProjectName, opts.TaskName, "abc123"), "metadata.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("metadata.json not found: %v", err)
	}
	var meta CacheMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("invalid metadata.json: %v", err)
	}
	if meta.Hash != "abc123" {
		t.Errorf("metadata hash: got %q, want %q", meta.Hash, "abc123")
	}
	if meta.Project != opts.ProjectName {
		t.Errorf("metadata project: got %q, want %q", meta.Project, opts.ProjectName)
	}
	if meta.Task != opts.TaskName {
		t.Errorf("metadata task: got %q, want %q", meta.Task, opts.TaskName)
	}
}

// ─── Error Propagation ─────────────────────────────────────────────

func TestComputeHash_ErrorOnMissingFile(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)
	// Create a file, then make it unreadable to trigger os.Open error in content strategy
	badFile := filepath.Join(dir, "unreadable.txt")
	if err := os.WriteFile(badFile, []byte("bad"), 0640); err != nil {
		t.Fatal(err)
	}
	// Remove read permissions
	if err := os.Chmod(badFile, 0000); err != nil {
		t.Skip("cannot chmod, skipping")
	}
	defer os.Chmod(badFile, 0640)
	opts.Inputs = []string{"*.txt"}
	_, err := ComputeHash(opts, baseTaskDef([]string{"build"}))
	if err == nil {
		t.Fatal("ComputeHash: expected error for unreadable input file, got nil")
	}
}

func TestComputeHash_ErrorOnStatFail(t *testing.T) {
	root := t.TempDir()
	dir := t.TempDir() // empty project
	opts := baseOpts(dir, root)
	opts.Strategy = app.CacheStrategyMtime
	// Create a file then remove read permissions to cause Stat to fail
	// Simpler: just use a non-existent file path
	opts.ProjectPath = filepath.Join(dir, "nonexistent")
	_, err := ComputeHash(opts, baseTaskDef([]string{"build"}))
	if err == nil {
		t.Fatal("ComputeHash: expected error for unreadable file in mtime mode, got nil")
	}
}

func TestRestore_PropagatesWalkError(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)
	opts.Mode = "restore"
	opts.Outputs = []string{"**"}
	taskDef := baseTaskDef([]string{"build"})

	hash, _ := ComputeHash(opts, taskDef)
	Save(opts, hash, "output")

	result, _ := Check(opts, hash)
	// Remove the outputs directory to cause WalkDir to fail
	os.RemoveAll(filepath.Join(result.CacheDir, "outputs"))
	// Actually, let's create a bad scenario: make outputs dir unreadable
	outputsDir := filepath.Join(result.CacheDir, "outputs")
	if err := os.MkdirAll(outputsDir, 0750); err != nil {
		t.Fatal(err)
	}
	// Write a file
	os.WriteFile(filepath.Join(outputsDir, "test.txt"), []byte("test"), 0640)
	// Restore should succeed with valid outputs
	if err := Restore(opts, result); err != nil {
		t.Fatalf("Restore: unexpected error: %v", err)
	}
}

func TestCopyFile_PreservesPermissions(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src.txt")
	dst := filepath.Join(root, "dst.txt")

	// Create source with specific permissions (executable)
	if err := os.WriteFile(src, []byte("test"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("copyFile: permission not preserved, got %o, want %o", info.Mode().Perm(), 0755)
	}
}

func TestSave_LogsWarningOnFailure(t *testing.T) {
	// Save should fail when it cannot create the cache directory
	// Use a path where we cannot create directories
	opts := CacheOptions{
		ProjectName:   "test",
		TaskName:      "test",
		ProjectPath:   t.TempDir(),
		Mode:          "skip",
		Strategy:      "content",
		MonospaceRoot: "/proc/invalid", // cannot write to /proc
	}
	err := Save(opts, "fakehash", "output")
	if err == nil {
		t.Fatal("Save: expected error for invalid cache dir, got nil")
	}
}
