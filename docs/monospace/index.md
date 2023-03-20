# monospace

![logo](../assets/logo-darkbg.png)

## requirements
- git command should be available in your Path
- js package manager defined in your config should also be available in your path
	(can be omit if you don't plan to work on javascript projects)
- go should be available in your path if you intend to work on go projects

## Installation

It's still early stage for monospace, so the installation process is not very fancy for now.

### download prebuilt binaries
Go to the [release page](https://github.com/software-t-rex/monospace/releases) and download the latest version for your platform, decompress the archive and make the monospace binary available from your $PATH.

### install from source
To install from source, if you already have a version of monospace installed you can ```monospace clone git@github.com:software-t-rex/monospace.git```
and then build the binary by issuing a ```go build main.go``` in the apps/monospace folder. Finally make the binary accessible from your $PATH.

If you don't have monospace installed, you have to first clone this repository and manually cloning all external repositories defined in .monospace/monospace.yml

> Other installation options are planned.

>At the moment monospace is only tested on linux platform, but it is compiled for other platforms too.
If you ran into issues with other platforms, please let us know, we will try our best to make it work. If you need other platform to be supported let us know and we will try to add it to the next release if doable.

## How to get started

### Brand new project/repo
Create a new dir then cd to that directory and issue a monospace [init command](./cli/md/monospace_init.md):
```bash
$ mkdir mynew-monospace
$ cd mynew-monospace
$ monospace init
```

### From a monorepo
From the root of your monorepo you need first to [init](./cli/md/monospace_init.md) a monospace.
```bash
$ cd my-monorepo
$ monospace init
```

You will then need to declare each of the subprojects your have using the [create command](./cli/md/monospace_create.md):

```bash
$ monospace create internal path/to/subprojects
```

once done you can decide to externalize some projects to their own repositories using the [externalize command](./cli/md/monospace_externalize.md)


### From polyrepo
There's multiple path here you can create a new repository for your monospace and start embedding
external projects inside the newly created monospace or from an existing repository that you want to turn into a monospace.

In both case go to the root folder of your future monospace  and use the [init command](./cli/md/monospace_init.md):
```bash
# from new repo
$ mkdir mynew-monospace
$ cd mynew-monospace
$ monospace init

# from existing repo
$ cd my-repo
$ monospace init
```

then you can start importing external projects inside the monospace using the [import command](./cli/md/monospace_import.md)
```bash
monospace import path/to/project/within/monospace git@your/external/repo/url.git
```

## Available commands
The best way to explore available commands is by using monospace itself with the -h flag.
Alternatively, you can find the list of available commands in the latest version and associated documentation [here](./cli/md/monospace.md).

## Configuration Options
Documentation regarding the monospace.yml configuration file can be found [here](./config/index.md)


## Some Default opinionated choices:
When initializing a new monospace it will declare some workspaces to your package manager:
- apps/* for applications
- packages/* for libraries

Default to add the following to monospace gitignore file (this will also be applied to local project created with monospace):
- node_modules
- .vscode
- .env
- dist
- coverage

Default package manager is pnpm 7, for now this is the only one tested but it should work correctly with yarn or npm, don't hesitate to report any issues with this package managers, they should be first citizen too.

monospace .npmrc will contains the following default settings
- auto-install-peers=true
- resolve-peers-from-workspace-root=true

If you think that this is not a good default feel free to contact us and explain why you think we should use other defaults. I'm always prone to change my mind about such decisions when there's good reasons to do so.
