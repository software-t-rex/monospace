//go:build unix

package cmd

import (
	"os"
	"os/signal"
	"syscall"
)

// Keep a reference to the channel: it is intentionally never read. Handling
// SIGPIPE only prevents the default process termination; subsequent writes to
// the closed stdout return EPIPE, which output writers ignore.
var sigpipe = make(chan os.Signal, 1)

// handleSIGPIPE prevents the process from being killed when stdout is closed
// (for example when piping `monospace run` into `head`). Writes then return
// EPIPE, which output writers ignore, so tasks can finish and fill the cache.
//
// signal.Notify is used instead of signal.Ignore so the caught disposition is
// reset to the default for child processes on exec.
func handleSIGPIPE() {
	signal.Notify(sigpipe, syscall.SIGPIPE)
}
