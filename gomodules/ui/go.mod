module github.com/software-t-rex/monospace/gomodules/ui

go 1.21.6

replace github.com/software-t-rex/monospace/gomodules/utils => ../utils

require golang.org/x/term v0.13.0

require (
	golang.org/x/sys v0.13.0
	gotest.tools/v3 v3.5.1
)

require github.com/google/go-cmp v0.5.9 // indirect
