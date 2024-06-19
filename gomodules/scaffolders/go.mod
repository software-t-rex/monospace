module github.com/software-t-rex/monospace/gomodules/scaffolders

go 1.21.6

require github.com/software-t-rex/monospace/gomodules/ui v0.0.0

replace (
	github.com/software-t-rex/monospace/gomodules/ui => ../ui
	github.com/software-t-rex/monospace/gomodules/utils => ../utils
)

require (
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/term v0.13.0 // indirect
)
