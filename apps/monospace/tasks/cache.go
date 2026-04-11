/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package tasks

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/software-t-rex/monospace/app"
)

// CacheOptions holds everything the cache layer needs to operate on a single task.
type CacheOptions struct {
	ProjectName   string
	TaskName      string
	ProjectPath   string   // absolute path to the project directory
	Mode          string   // "skip" | "restore"
	Strategy      string   // "content" | "mtime"
	Inputs        []string // glob patterns relative to ProjectPath (empty = all files)
	Outputs       []string // glob patterns relative to ProjectPath
	MonospaceRoot string
	MaxEntries    int // maximum number of hash entries to keep per task (0 = no limit)
}

// CacheMetadata is stored alongside a cache entry to record when it was created.
type CacheMetadata struct {
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
	Project   string    `json:"project"`
	Task      string    `json:"task"`
	Strategy  string    `json:"strategy"`
}

// CacheResult is returned by Check.
type CacheResult struct {
	Hit      bool
	Hash     string
	CacheDir string // full path to .monospace/.cache/{proj}#{task}/{hash}
}

// CacheStatusEntry describes the cache state of a given task.
type CacheStatusEntry struct {
	Project  string
	Task     string
	Hash     string
	CachedAt time.Time
}

// cacheBaseDir returns the base cache directory for this monospace.
func cacheBaseDir(monospaceRoot string) string {
	return filepath.Join(monospaceRoot, ".monospace", ".cache")
}

// cacheTaskDir returns the per-task cache directory (all hash entries live here).
func cacheTaskDir(monospaceRoot, project, task string) string {
	projectKey := strings.ReplaceAll(project, "/", "__")
	return filepath.Join(cacheBaseDir(monospaceRoot), projectKey+"#"+task)
}

// cacheEntryDir returns the path for a specific hash entry.
func cacheEntryDir(monospaceRoot, project, task, hash string) string {
	return filepath.Join(cacheTaskDir(monospaceRoot, project, task), hash)
}

// ComputeHash returns a hex-encoded SHA256 hash over the resolved input files
// for the given task. The hash also covers the task command and env vars so
// that changes to the task definition itself bust the cache.
//
// Strategy "content" (default) hashes the actual file bytes.
// Strategy "mtime" hashes "{relpath}:{size}:{modtime_unix}\n" for each file.
func ComputeHash(opts CacheOptions, taskDef app.MonospaceConfigTask) (string, error) {
	h := sha256.New()

	// 1. Stable task identity
	fmt.Fprintf(h, "%s|%s|%s\n", opts.TaskName, opts.ProjectName, strings.Join(taskDef.Cmd, " "))

	// 2. Sorted env key=value pairs
	envKeys := make([]string, 0, len(taskDef.Env))
	for k := range taskDef.Env {
		envKeys = append(envKeys, k)
	}
	sort.Strings(envKeys)
	for _, k := range envKeys {
		fmt.Fprintf(h, "%s=%s\n", k, taskDef.Env[k])
	}

	// 3. Resolve input file list
	files, err := resolveInputFiles(opts)
	if err != nil {
		return "", fmt.Errorf("resolving input files for %s#%s: %w", opts.ProjectName, opts.TaskName, err)
	}
	sort.Strings(files)

	// 4. Hash each file according to the chosen strategy
	strategy := opts.Strategy
	if strategy == "" {
		strategy = app.CacheStrategyContent
	}
	for _, file := range files {
		rel, _ := filepath.Rel(opts.ProjectPath, file)
		if strategy == app.CacheStrategyMtime {
			info, err := os.Stat(file)
			if err != nil {
				continue
			}
			fmt.Fprintf(h, "%s:%d:%d\n", rel, info.Size(), info.ModTime().Unix())
		} else {
			// content strategy
			f, err := os.Open(file)
			if err != nil {
				continue
			}
			fmt.Fprintf(h, "%s:", rel)
			io.Copy(h, f) //nolint:errcheck
			f.Close()
			fmt.Fprint(h, "\n")
		}
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// resolveInputFiles returns the list of files to include in the cache key.
// If no inputs are configured all files in the project directory are used
// (excluding .git and .monospace directories).
func resolveInputFiles(opts CacheOptions) ([]string, error) {
	if len(opts.Inputs) == 0 {
		return walkProjectFiles(opts.ProjectPath)
	}

	seen := make(map[string]struct{})
	var files []string
	fsys := os.DirFS(opts.ProjectPath)
	for _, pattern := range opts.Inputs {
		matches, err := doublestar.Glob(fsys, pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}
		for _, match := range matches {
			abs := filepath.Join(opts.ProjectPath, match)
			info, err := os.Stat(abs)
			if err != nil || info.IsDir() {
				continue
			}
			if _, exists := seen[abs]; !exists {
				seen[abs] = struct{}{}
				files = append(files, abs)
			}
		}
	}
	return files, nil
}

// walkProjectFiles recursively collects all regular files in dir,
// skipping .git and .monospace directories.
func walkProjectFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == ".monospace" {
				return filepath.SkipDir
			}
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

// Check looks up whether a cache entry exists for the given hash.
func Check(opts CacheOptions, hash string) (CacheResult, error) {
	entryDir := cacheEntryDir(opts.MonospaceRoot, opts.ProjectName, opts.TaskName, hash)
	metaPath := filepath.Join(entryDir, "metadata.json")

	data, err := os.ReadFile(metaPath)
	if err != nil {
		return CacheResult{Hit: false, Hash: hash}, nil
	}
	var meta CacheMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return CacheResult{Hit: false, Hash: hash}, nil
	}
	return CacheResult{Hit: true, Hash: hash, CacheDir: entryDir}, nil
}

// Save writes metadata, the task output, and (in restore mode) archives the
// output files. Should be called after a task completes successfully.
// The output string is stored so it can be replayed on subsequent cache hits.
func Save(opts CacheOptions, hash string, output string) error {
	entryDir := cacheEntryDir(opts.MonospaceRoot, opts.ProjectName, opts.TaskName, hash)
	if err := os.MkdirAll(entryDir, 0750); err != nil {
		return fmt.Errorf("creating cache entry dir: %w", err)
	}

	meta := CacheMetadata{
		Hash:      hash,
		Timestamp: time.Now(),
		Project:   opts.ProjectName,
		Task:      opts.TaskName,
		Strategy:  opts.Strategy,
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshaling cache metadata: %w", err)
	}
	if err := os.WriteFile(filepath.Join(entryDir, "metadata.json"), data, 0640); err != nil {
		return fmt.Errorf("writing cache metadata: %w", err)
	}

	// Always persist the command output for replay on cache hits.
	if err := os.WriteFile(filepath.Join(entryDir, "output.txt"), []byte(output), 0640); err != nil {
		return fmt.Errorf("writing cached output: %w", err)
	}

	// Prune old entries beyond the configured limit.
	if opts.MaxEntries > 0 {
		_ = PruneTaskCache(opts.MonospaceRoot, opts.ProjectName, opts.TaskName, opts.MaxEntries)
	}

	// Save output files only in restore mode
	if opts.Mode == app.CacheModeRestore && len(opts.Outputs) > 0 {
		outputsDir := filepath.Join(entryDir, "outputs")
		if err := os.MkdirAll(outputsDir, 0750); err != nil {
			return fmt.Errorf("creating outputs dir: %w", err)
		}
		fsys := os.DirFS(opts.ProjectPath)
		for _, pattern := range opts.Outputs {
			matches, err := doublestar.Glob(fsys, pattern)
			if err != nil {
				continue
			}
			for _, match := range matches {
				src := filepath.Join(opts.ProjectPath, match)
				info, err := os.Stat(src)
				if err != nil || info.IsDir() {
					continue
				}
				dst := filepath.Join(outputsDir, match)
				if err := copyFile(src, dst); err != nil {
					return fmt.Errorf("caching output file %s: %w", match, err)
				}
			}
		}
	}
	return nil
}

// Restore copies cached output files back to the project directory.
// Only useful in restore mode on a cache hit.
func Restore(opts CacheOptions, result CacheResult) error {
	outputsDir := filepath.Join(result.CacheDir, "outputs")
	if _, err := os.Stat(outputsDir); os.IsNotExist(err) {
		return nil // nothing to restore
	}
	return filepath.WalkDir(outputsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(outputsDir, path)
		if err != nil {
			return nil
		}
		dst := filepath.Join(opts.ProjectPath, rel)
		return copyFile(path, dst)
	})
}

// readCachedOutput reads the output stored by Save for the given cache entry.
// Returns an empty string (no error) if no output was recorded.
func readCachedOutput(cacheDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(cacheDir, "output.txt"))
	if os.IsNotExist(err) {
		return "", nil
	}
	return string(data), err
}

// ClearCache removes all cache entries for a given project+task pair.
// If project and task are both empty, the entire cache directory is removed.
func ClearCache(monospaceRoot, project, task string) error {
	var dir string
	if project == "" && task == "" {
		dir = cacheBaseDir(monospaceRoot)
	} else {
		dir = cacheTaskDir(monospaceRoot, project, task)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dir)
}

// GetCacheStatus returns the cache status entries for all cached tasks,
// optionally filtered by a list of "project#task" strings.
func GetCacheStatus(monospaceRoot string, filters []string) ([]CacheStatusEntry, error) {
	base := cacheBaseDir(monospaceRoot)
	entries, err := os.ReadDir(base)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading cache dir: %w", err)
	}

	var result []CacheStatusEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// directory name is "{project_encoded}#{task}"
		dirName := entry.Name()
		lastHash := strings.LastIndex(dirName, "#")
		if lastHash < 0 {
			continue
		}
		project := strings.ReplaceAll(dirName[:lastHash], "__", "/")
		taskName := dirName[lastHash+1:]
		taskKey := project + "#" + taskName

		// apply filter
		if len(filters) > 0 {
			found := false
			for _, f := range filters {
				if f == taskKey || f == project || f == taskName {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// enumerate hash subdirectories to find the most recent entry
		taskDir := filepath.Join(base, dirName)
		hashEntries, err := os.ReadDir(taskDir)
		if err != nil {
			continue
		}
		for _, hashEntry := range hashEntries {
			if !hashEntry.IsDir() {
				continue
			}
			metaPath := filepath.Join(taskDir, hashEntry.Name(), "metadata.json")
			data, err := os.ReadFile(metaPath)
			if err != nil {
				continue
			}
			var meta CacheMetadata
			if err := json.Unmarshal(data, &meta); err != nil {
				continue
			}
			result = append(result, CacheStatusEntry{
				Project:  project,
				Task:     taskName,
				Hash:     meta.Hash,
				CachedAt: meta.Timestamp,
			})
		}
	}
	return result, nil
}

// PruneTaskCache removes the oldest cache entries for a given project+task pair,
// keeping at most maxEntries. Entries are sorted by their recorded timestamp;
// the newest ones are kept.
func PruneTaskCache(monospaceRoot, project, task string, maxEntries int) error {
	taskDir := cacheTaskDir(monospaceRoot, project, task)
	rawEntries, err := os.ReadDir(taskDir)
	if os.IsNotExist(err) || len(rawEntries) <= maxEntries {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading cache task dir: %w", err)
	}

	type hashEntry struct {
		dir       string
		timestamp time.Time
	}
	var entries []hashEntry
	for _, e := range rawEntries {
		if !e.IsDir() {
			continue
		}
		metaPath := filepath.Join(taskDir, e.Name(), "metadata.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}
		var meta CacheMetadata
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		entries = append(entries, hashEntry{
			dir:       filepath.Join(taskDir, e.Name()),
			timestamp: meta.Timestamp,
		})
	}

	if len(entries) <= maxEntries {
		return nil
	}

	// Keep the newest entries; remove the rest.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].timestamp.After(entries[j].timestamp)
	})
	for _, e := range entries[maxEntries:] {
		_ = os.RemoveAll(e.dir)
	}
	return nil
}

// copyFile copies a file from src to dst, creating parent directories as needed.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
