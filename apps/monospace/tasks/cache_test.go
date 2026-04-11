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
		Mode:          app.CacheModeSkip,
		Strategy:      app.CacheStrategyContent,
		MonospaceRoot: monospaceRoot,
	}
}

func baseTaskDef(cmd []string) app.MonospaceConfigTask {
	return app.MonospaceConfigTask{
		Cmd:   cmd,
		Cache: app.CacheModeSkip,
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
	opts.Mode = app.CacheModeRestore
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
	opts.Mode = app.CacheModeRestore
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
