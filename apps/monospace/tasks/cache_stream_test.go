/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package tasks

import (
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	jobExecutor "github.com/software-t-rex/go-jobExecutor/v2"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

// streamCmd builds a command echoing the given lines on stdout.
func streamCmd(lines ...string) *exec.Cmd {
	script := ""
	for _, l := range lines {
		script += "echo " + l + "; "
	}
	script += "true"
	return exec.Command("sh", "-c", script)
}

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")

func stripAnsi(s string) string { return ansiRe.ReplaceAllString(s, "") }

// captureStdout runs fn with os.Stdout redirected to a pipe and returns what
// was written to it.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating stdout pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("closing stdout pipe: %v", err)
	}
	os.Stdout = orig
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("reading captured stdout: %v", err)
	}
	return buf.String()
}

// ─── runCmdCaptured ───────────────────────────────────────────────────────────

func TestRunCmdCaptured_TeesToLiveWriter(t *testing.T) {
	var live bytes.Buffer
	out, err := runCmdCaptured(streamCmd("hello", "world"), &live)
	if err != nil {
		t.Fatalf("runCmdCaptured: unexpected error: %v", err)
	}
	if want := "hello\nworld\n"; out != want {
		t.Errorf("runCmdCaptured: captured output = %q, want %q", out, want)
	}
	if out != live.String() {
		t.Errorf("runCmdCaptured: live output %q differs from captured output %q", live.String(), out)
	}
}

func TestRunCmdCaptured_NilWriterUnchangedBehaviour(t *testing.T) {
	out, err := runCmdCaptured(streamCmd("hello"), nil)
	if err != nil {
		t.Fatalf("runCmdCaptured: unexpected error: %v", err)
	}
	if want := "hello\n"; out != want {
		t.Errorf("runCmdCaptured: captured output = %q, want %q", out, want)
	}
}

// TestRunCmdCaptured_StderrIsCapturedToo ensures the tee also covers stderr:
// both streams must land in the capture buffer (cache source of truth).
func TestRunCmdCaptured_StderrIsCapturedToo(t *testing.T) {
	var live bytes.Buffer
	cmd := exec.Command("sh", "-c", "echo on-stdout; echo on-stderr >&2")
	out, err := runCmdCaptured(cmd, &live)
	if err != nil {
		t.Fatalf("runCmdCaptured: unexpected error: %v", err)
	}
	for _, want := range []string{"on-stdout", "on-stderr"} {
		if !strings.Contains(out, want) {
			t.Errorf("captured output %q missing %q", out, want)
		}
		if !strings.Contains(live.String(), want) {
			t.Errorf("live output %q missing %q", live.String(), want)
		}
	}
}

// TestRunCmdCaptured_ConcurrentStreamsNoRace exercises heavy interleaved
// stdout/stderr traffic: it must not race on the capture buffer (guarded by
// running the suite with -race) and must not lose any byte.
func TestRunCmdCaptured_ConcurrentStreamsNoRace(t *testing.T) {
	var live bytes.Buffer
	cmd := exec.Command("sh", "-c",
		"i=0; while [ $i -lt 200 ]; do echo out-$i; echo err-$i >&2; i=$((i+1)); done")
	out, err := runCmdCaptured(cmd, &live)
	if err != nil {
		t.Fatalf("runCmdCaptured: unexpected error: %v", err)
	}
	if n := strings.Count(out, "out-"); n != 200 {
		t.Errorf("captured stdout lines = %d, want 200", n)
	}
	if n := strings.Count(out, "err-"); n != 200 {
		t.Errorf("captured stderr lines = %d, want 200", n)
	}
	if out != live.String() {
		t.Errorf("live output differs from captured output (%d vs %d bytes)", len(live.String()), len(out))
	}
}

// ─── runWithCache: cache miss ─────────────────────────────────────────────────

// TestRunWithCache_Miss_StreamsAndSavesCompleteOutput is the core acceptance
// test: on a miss the output must be streamed live AND fully saved to the cache,
// while the returned string stays empty to prevent a double dump.
func TestRunWithCache_Miss_StreamsAndSavesCompleteOutput(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)
	taskDef := baseTaskDef([]string{"build"})

	hash, err := ComputeHash(opts, taskDef)
	if err != nil {
		t.Fatalf("ComputeHash: %v", err)
	}

	var live bytes.Buffer
	out, runErr := runWithCache(opts, taskDef, streamCmd("line1", "line2"), &live)
	if runErr != nil {
		t.Fatalf("runWithCache: unexpected error: %v", runErr)
	}
	if want := "line1\nline2\n"; live.String() != want {
		t.Errorf("live output = %q, want %q", live.String(), want)
	}
	// returned "" => executor will not dump the output a second time
	if out != "" {
		t.Errorf("returned output = %q, want empty string when streaming (anti double dump)", out)
	}

	// the cache must still hold the COMPLETE output for the next hit
	result, checkErr := Check(opts, hash)
	if checkErr != nil {
		t.Fatalf("Check: %v", checkErr)
	}
	if !result.Hit {
		t.Fatal("cache entry missing after a successful miss")
	}
	cached, readErr := readCachedOutput(result.CacheDir)
	if readErr != nil {
		t.Fatalf("readCachedOutput: %v", readErr)
	}
	if want := "line1\nline2\n"; cached != want {
		t.Errorf("cached output = %q, want %q", cached, want)
	}
}

// TestRunWithCache_Miss_NoLiveWriter_ReturnsFullOutput is the anti-regression
// guard for grouped/none/errors-only modes: without a live writer the behaviour
// must be exactly the previous one.
func TestRunWithCache_Miss_NoLiveWriter_ReturnsFullOutput(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)
	taskDef := baseTaskDef([]string{"build"})

	hash, err := ComputeHash(opts, taskDef)
	if err != nil {
		t.Fatalf("ComputeHash: %v", err)
	}

	out, runErr := runWithCache(opts, taskDef, streamCmd("line1", "line2"), nil)
	if runErr != nil {
		t.Fatalf("runWithCache: unexpected error: %v", runErr)
	}
	if want := "line1\nline2\n"; out != want {
		t.Errorf("returned output = %q, want %q", out, want)
	}

	result, _ := Check(opts, hash)
	cached, _ := readCachedOutput(result.CacheDir)
	if want := "line1\nline2\n"; cached != want {
		t.Errorf("cached output = %q, want %q", cached, want)
	}
}

// TestRunWithCache_Miss_FailedCommandNotSaved ensures a failing command is
// streamed but never cached.
func TestRunWithCache_Miss_FailedCommandNotSaved(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)
	taskDef := baseTaskDef([]string{"build"})

	hash, err := ComputeHash(opts, taskDef)
	if err != nil {
		t.Fatalf("ComputeHash: %v", err)
	}

	var live bytes.Buffer
	cmd := exec.Command("sh", "-c", "echo boom; exit 3")
	out, runErr := runWithCache(opts, taskDef, cmd, &live)
	if runErr == nil {
		t.Fatal("runWithCache: expected an error for a failing command")
	}
	if !strings.Contains(live.String(), "boom") {
		t.Errorf("live output = %q, want it to contain %q", live.String(), "boom")
	}
	if out != "" {
		t.Errorf("returned output = %q, want empty string when streaming", out)
	}
	if result, _ := Check(opts, hash); result.Hit {
		t.Error("a failed command must not be cached")
	}
}

// ─── runWithCache: cache hit ──────────────────────────────────────────────────

func TestRunWithCache_Hit_ReplaysCachedOutput(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)
	taskDef := baseTaskDef([]string{"build"})

	hash, err := ComputeHash(opts, taskDef)
	if err != nil {
		t.Fatalf("ComputeHash: %v", err)
	}
	if err := Save(opts, hash, "cached-output\n"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var live bytes.Buffer
	out, runErr := runWithCache(opts, taskDef, streamCmd("from-cmd"), &live)
	if runErr != nil {
		t.Fatalf("runWithCache: unexpected error: %v", runErr)
	}
	want := "[cache hit: " + hash[:8] + "]\ncached-output\n"
	if out != want {
		t.Errorf("replay = %q, want %q", out, want)
	}
	if live.Len() != 0 {
		t.Errorf("cache hit must not stream to the live writer, got %q", live.String())
	}
}

func TestRunWithCache_Hit_RestoreMode(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)
	opts.Mode = "restore"
	taskDef := baseTaskDef([]string{"build"})

	hash, err := ComputeHash(opts, taskDef)
	if err != nil {
		t.Fatalf("ComputeHash: %v", err)
	}
	if err := Save(opts, hash, "cached-output\n"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var live bytes.Buffer
	out, runErr := runWithCache(opts, taskDef, streamCmd("from-cmd"), &live)
	if runErr != nil {
		t.Fatalf("runWithCache: unexpected error: %v", runErr)
	}
	want := "[cache hit: " + hash[:8] + ", restored]\ncached-output\n"
	if out != want {
		t.Errorf("replay = %q, want %q", out, want)
	}
	if live.Len() != 0 {
		t.Errorf("cache hit must not stream to the live writer, got %q", live.String())
	}
}

// ─── runWithCache: hash failure ───────────────────────────────────────────────

// TestRunWithCache_HashFailure_StreamsAndSkipsSave covers the "hash failure is
// non fatal" path: the task still runs, streams live, returns "" and caches
// nothing.
func TestRunWithCache_HashFailure_StreamsAndSkipsSave(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	opts := baseOpts(dir, root)
	opts.Inputs = []string{"["} // invalid glob => ComputeHash fails
	taskDef := baseTaskDef([]string{"build"})

	if _, err := ComputeHash(opts, taskDef); err == nil {
		t.Skip("invalid glob pattern is not rejected by the glob implementation")
	}

	var live bytes.Buffer
	out, runErr := runWithCache(opts, taskDef, streamCmd("line1"), &live)
	if runErr != nil {
		t.Fatalf("runWithCache: unexpected error: %v", runErr)
	}
	if want := "line1\n"; live.String() != want {
		t.Errorf("live output = %q, want %q", live.String(), want)
	}
	if out != "" {
		t.Errorf("returned output = %q, want empty string when streaming", out)
	}
	if _, err := os.Stat(cacheTaskDir(root, opts.ProjectName, opts.TaskName)); !os.IsNotExist(err) {
		t.Errorf("no cache entry must be created on hash failure (stat err = %v)", err)
	}
}

// ─── end to end: executor + cache closure ────────────────────────────────────

// TestInterleaved_CacheMiss_StreamsExactlyOnce runs the real interleaved
// executor around a cache closure and checks the user facing contract: the task
// output appears exactly once, prefixed by the task name, and the cache entry is
// filled for the next run.
func TestInterleaved_CacheMiss_StreamsExactlyOnce(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	cacheOpts := baseOpts(dir, root)
	taskDef := baseTaskDef([]string{"build"})
	const taskName = "myproject#build"
	const marker = "STREAMED-MARKER"

	hash, err := ComputeHash(cacheOpts, taskDef)
	if err != nil {
		t.Fatalf("ComputeHash: %v", err)
	}

	e := NewExecutor("interleaved")
	e.AddJob(jobExecutor.NamedJob{Name: taskName, Job: func() (string, error) {
		// same wiring as wrapWithCache: writer built at run time, prefix lazily
		// resolved from the job display name.
		live := jobExecutor.NewPrefixedWriter(os.Stdout, taskName+": ")
		return runWithCache(cacheOpts, taskDef, streamCmd(marker), live)
	}})

	captured := captureStdout(t, func() {
		if errs := e.DagExecute(); errs.Len() > 0 {
			t.Errorf("DagExecute: unexpected errors: %v", errs)
		}
	})
	clean := stripAnsi(captured)

	if n := strings.Count(clean, marker); n != 1 {
		t.Errorf("output appeared %d times, want exactly 1 (no double dump):\n%s", n, clean)
	}
	if !strings.Contains(clean, taskName+": "+marker) {
		t.Errorf("streamed output is not prefixed with %q:\n%s", taskName+": ", clean)
	}

	result, _ := Check(cacheOpts, hash)
	if !result.Hit {
		t.Fatal("cache entry missing after an interleaved cache miss")
	}
	cached, _ := readCachedOutput(result.CacheDir)
	if !strings.Contains(cached, marker) {
		t.Errorf("cached output %q does not contain %q", cached, marker)
	}
}

// TestInterleaved_PrefixIsTheFinalDisplayName reproduces the production wiring
// shape (job handle assigned *after* AddJob, prefix resolved inside the closure)
// and proves the lazy prefix resolution: the streamed prefix must be the display
// name as rewritten by the executor at OnJobsStart (last "/" segment), not the
// raw registered name.
//
// Regression guard for the prefix bug identified in the plan: a non lazy prefix
// would stream "apps/localapp#build: " while cache-hit replays and error dumps
// use "localapp#build: ".
func TestInterleaved_PrefixIsTheFinalDisplayName(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	cacheOpts := baseOpts(dir, root)
	taskDef := baseTaskDef([]string{"build"})
	const registeredName = "apps/localapp#build"
	const expectedPrefix = "localapp#build: "
	const marker = "LAZY-MARKER"

	e := NewExecutor("interleaved")
	var j jobExecutor.Job
	var livePrefix func() string
	livePrefix = func() string { return j.Name() }

	j = e.AddJob(jobExecutor.NamedJob{Name: registeredName, Job: func() (string, error) {
		// mirrored from wrapWithCache: the writer is built at run time
		live := jobExecutor.NewPrefixedWriter(os.Stdout, livePrefix()+": ")
		return runWithCache(cacheOpts, taskDef, streamCmd(marker), live)
	}})

	captured := captureStdout(t, func() {
		if errs := e.DagExecute(); errs.Len() > 0 {
			t.Errorf("DagExecute: unexpected errors: %v", errs)
		}
	})
	clean := stripAnsi(captured)

	if !strings.Contains(clean, expectedPrefix+marker) {
		t.Errorf("streamed output is not prefixed with the resolved display name %q:\n%s", expectedPrefix, clean)
	}
	if strings.Contains(clean, registeredName+": ") {
		t.Errorf("streamed output still uses the raw registered name %q (lazy prefix resolution broken):\n%s", registeredName+": ", clean)
	}
}

// TestInterleaved_CacheHit_ReplaysThroughExecutor checks the hit path end to
// end: the executor prints the replay string, which therefore appears once.
func TestInterleaved_CacheHit_ReplaysThroughExecutor(t *testing.T) {
	root := t.TempDir()
	dir := makeTestProject(t, map[string]string{"a.go": "a"})
	cacheOpts := baseOpts(dir, root)
	taskDef := baseTaskDef([]string{"build"})
	const taskName = "myproject#build"
	const marker = "CACHED-MARKER"

	hash, err := ComputeHash(cacheOpts, taskDef)
	if err != nil {
		t.Fatalf("ComputeHash: %v", err)
	}
	if err := Save(cacheOpts, hash, marker+"\n"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	e := NewExecutor("interleaved")
	e.AddJob(jobExecutor.NamedJob{Name: taskName, Job: func() (string, error) {
		live := jobExecutor.NewPrefixedWriter(os.Stdout, taskName+": ")
		return runWithCache(cacheOpts, taskDef, streamCmd("from-cmd"), live)
	}})

	captured := captureStdout(t, func() {
		if errs := e.DagExecute(); errs.Len() > 0 {
			t.Errorf("DagExecute: unexpected errors: %v", errs)
		}
	})
	clean := stripAnsi(captured)

	if n := strings.Count(clean, marker); n != 1 {
		t.Errorf("replayed output appeared %d times, want exactly 1:\n%s", n, clean)
	}
	if strings.Contains(clean, "from-cmd") {
		t.Errorf("command must not run on a cache hit:\n%s", clean)
	}
}
