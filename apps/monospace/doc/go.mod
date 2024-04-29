module github.com/software-t-rex/monospace/doc

go 1.21.6

replace (
	github.com/software-t-rex/monospace => ../
	github.com/software-t-rex/monospace/gomodules/colors => ../../../gomodules/colors
	github.com/software-t-rex/monospace/gomodules/scaffolders => ../../../gomodules/scaffolders
	github.com/software-t-rex/monospace/gomodules/ui => ../../../gomodules/ui
	github.com/software-t-rex/monospace/gomodules/utils => ../../../gomodules/utils
)

require (
	github.com/software-t-rex/monospace v0.0.0-00010101000000-000000000000
	github.com/software-t-rex/monospace/gomodules/ui v0.0.0
	github.com/software-t-rex/monospace/gomodules/utils v0.0.0
	github.com/spf13/cobra v1.6.1
)

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/software-t-rex/go-jobExecutor/v2 v2.1.2 // indirect
	github.com/software-t-rex/js-packagemanager v0.0.5 // indirect
	github.com/software-t-rex/monospace/gomodules/colors v0.0.0 // indirect
	github.com/software-t-rex/monospace/gomodules/scaffolders v0.0.0 // indirect
	github.com/software-t-rex/packageJson v0.0.3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/term v0.19.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
