# monospace configuration

To configure your monospace you can edit file .monospace/monospace.yml at the root of your monospace directory.
Filename MUST be monospace.**yml** not .monospace.yaml.

> This documentation may be late at describing options as the configuration options evolve. Latest options will always be described in [monospace.schema.json](https://raw.githubusercontent.com/software-t-rex/monospace/main/apps/monospace/schemas/monospace.schema.json)

## js_package_manager (string)
**defaults**: ^pnpm@8.8.0

this is used to define you package manager of choice. It will be used when you will want to run tasks against javascript projects.
the syntax is similar to the packageManager property in package.json files.

## go_mod_prefix (string)
**defaults**: example.com

Prefix to use when creating new go modules.

## preferred_output_mode (string)
**default**: grouped

Default output mode to use for run and exec commands.
- none: no output from tasks
- interleaved: print prefixed output from tasks as they arrive
- grouped: print combined output of tasks as they complete.
- status-only: display running status summary
- errors-only: is like interleaved output but displaying only what comes on stderr

> You can always override this with the --output-mode option of the run or exec command

## projects (object)
It is preferred to use monospace create/import/externalize/remove commands to edit projects settings.
But it can be sometimes useful to edit it manually, if you know what you are doing.
It is a Key=value pair object with relative path to projects from the root of your monospace as keys, and associated repositories URL as values.

The repository URL can take following values:
- "internal" for projects that use the monospace root repository
- "local" for projects that don't have a remote repository
- a valid git repository URL

## aliases (object)
You should prefer to use the **aliases** command instead of editing manually this setting.
It is simply a list of aliases as keys associated with relative project path in your monospace.

Aliases can be used when defining tasks in the pipeline, or when filtering projects for various commands.

## pipeline (object)

### taskName (string)
the key inside pipeline define the task name, but there's a little more than that. You can prefix the task name with a project name (relative path in the monospace) or alias like this: ```path/to/project#mytask```

### cmd (array of string)
For js projects you don't need to specify a command, it will call the script with the same task name in the package.json of the project. For other languages, or for specific needs you can define a custom command to run. Here's some important things to know about this:
- the working directory will be the project directory
	- you can call script in the project directory by prefixing them with **./**
	- it's not real shell command but go exec.Command so the same limitations apply, for example you can't use redirection or logical operators. If you need fancy stuff like that you should create a shell script which you will point to in the **cmd** property.
- the dir .monospace/bin/ will be added to your $PATH. So you can put global executable there and you will be able to run them for each project. (beware that it will be looked up after your already defined $PATH environment variable so if you put a binary with a name that is already accessible in your path, the one in .monospace/bin won't be called)
- env variables ($Var or ${VAR}) will be replaced, monospace will populate a bunch of them such as:
	- MONOSPACE_ROOT: absolute path of to your monospace
	- MONOSPACE_VERSION: the version of your monospace cli
	- MONOSPACE_JSPM: the value of _js_package_manager_
	- MONOSPACE_GOPREFIX: the value of _go_mod_prefix_

### dependsOn (array)
List of other tasks that need to complete before executing this one.
Task names in the list that are not prefixed will match a task defined for the same project.
You can specify another task from another project by prefixing the task name with the project name
(relative path to the project within the monospace directory) and a sharp.

> Dependencies are not magically resolved and must be explicitly set in the config file. This is, at least for now a design choice!

Here's an example:
```yaml
pipeline:
	test: {} // you can't leave a task blank but you can define it as an empty object literal this is useful for package.json tasks
	myproject#build:
		- dependsOn: [test, myOtherProject#build]
```
In this configuration myproject#build will depend on the tests for myproject and the build of myOtherProject to be successful.

### persistent (boolean)
**default** false
Persistent tasks are long-running process such as server, or watchers that will not exit unless manually stopped.
> It is important to mark such tasks as persistent, doing so monospace will prevent other tasks to depend on them and will inform you of configuration problem when calling the run command or when performing a ```monospace check```.

### output_mode (string)
**default**: to preferred_output_mode

If set monospace will try to respect this setting when calling run command.
If run launches multiple tasks with different settings it will default to the preferred_output_mode
You can always override this with the --output-mode option of the run command
Acceptable values are the same as preferred_output_mode.


