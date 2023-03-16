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
	github.com/spf13/viper v1.15.0
	sigs.k8s.io/yaml v1.3.0
)

replace github.com/software-t-rex/monospace/colors => ../../gomodules/colors

replace github.com/software-t-rex/monospace/scaffolders => ../../gomodules/scaffolders

replace github.com/software-t-rex/packageJson => ../../gomodules/packageJson

replace github.com/software-t-rex/go-jobExecutor/v2 => ../../gomodules/go-jobexecutor

replace github.com/software-t-rex/js-packagemanager => ../../gomodules/js-packagemanager

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
