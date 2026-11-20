//go:build !unix

package cmd

// handleSIGPIPE is a no-op on platforms without SIGPIPE.
func handleSIGPIPE() {}
