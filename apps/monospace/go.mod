module github.com/software-t-rex/monospace

go 1.20

// modules from the monospace
require github.com/software-t-rex/monospace/colors v0.0.0

require github.com/software-t-rex/monospace/scaffolders v0.0.0

require github.com/software-t-rex/js-packagemanager v0.0.0

require github.com/software-t-rex/packageJson v0.0.3

require (
	github.com/software-t-rex/go-jobExecutor/v2 v2.0.0
	github.com/spf13/cobra v1.6.1
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/software-t-rex/monospace/colors => ../../gomodules/colors

replace github.com/software-t-rex/monospace/scaffolders => ../../gomodules/scaffolders

replace github.com/software-t-rex/packageJson => ../../gomodules/packageJson

replace github.com/software-t-rex/go-jobExecutor/v2 => ../../gomodules/go-jobexecutor

replace github.com/software-t-rex/js-packagemanager => ../../gomodules/js-packagemanager

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.6.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)
