module github.com/software-t-rex/monospace

go 1.21.6

// modules from the monospace
require (
	github.com/software-t-rex/go-jobExecutor/v2 v2.1.2
	github.com/software-t-rex/js-packagemanager v0.0.5
	github.com/software-t-rex/monospace/gomodules/colors v0.0.0
	github.com/software-t-rex/monospace/gomodules/scaffolders v0.0.0
	github.com/software-t-rex/monospace/gomodules/ui v0.0.0
	github.com/software-t-rex/monospace/gomodules/utils v0.0.0
	github.com/software-t-rex/packageJson v0.0.3
)

replace (
	// github.com/software-t-rex/go-jobExecutor/v2 => ../../gomodules/go-jobexecutor
	github.com/software-t-rex/monospace/gomodules/colors => ../../gomodules/colors
	github.com/software-t-rex/monospace/gomodules/scaffolders => ../../gomodules/scaffolders
	github.com/software-t-rex/monospace/gomodules/ui => ../../gomodules/ui
	github.com/software-t-rex/monospace/gomodules/utils => ../../gomodules/utils
)

require (
	github.com/spf13/cobra v1.6.1
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools/v3 v3.4.0
)

require (
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/term v0.19.0 // indirect
)

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.0 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.6.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)
