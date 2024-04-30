module github.com/software-t-rex/monospace/gomodules/utils

go 1.21.6

// modules from the monospace
require github.com/software-t-rex/monospace/gomodules/ui v0.0.0

require (
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/term v0.13.0 // indirect
)

replace github.com/software-t-rex/monospace/gomodules/ui => ../../gomodules/ui
