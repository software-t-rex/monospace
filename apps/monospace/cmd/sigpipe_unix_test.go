//go:build unix

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

// TestHandleSIGPIPE runs a child process writing to a closed stdout pipe.
// Without the SIGPIPE handler the child is killed by the signal; with it the
// write returns EPIPE and the process exits normally.
func TestHandleSIGPIPE(t *testing.T) {
	if os.Getenv("MONOSPACE_TEST_SIGPIPE_CHILD") == "1" {
		handleSIGPIPE()
		r, w, err := os.Pipe()
		if err != nil {
			os.Exit(2)
		}
		_ = r.Close()
		os.Stdout = w
		fmt.Println("this write targets a closed pipe")
		_ = w.Close()
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestHandleSIGPIPE$")
	cmd.Env = append(os.Environ(), "MONOSPACE_TEST_SIGPIPE_CHILD=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child process was killed by SIGPIPE: %v\n%s", err, out)
	}
}
