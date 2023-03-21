## monospace run

Run given tasks in each project directory

### Synopsis

Run given command in each project directory concurrently.

You can restrict the tasks execution to one or more projects
using the --project-filter flag.

Example:
```
  monospace run --project-filter modules/mymodule --project-filter modules/myothermodule test
```
or more concise
```
  monospace run -p modules/mymodule,modules/myothermodule test
```

you can get a dependency graph of tasks to run by using the --graphviz flag.
It will output the dot representation in your terminal and open your browser
for visual online rendering.

```
  monospace run task --graphviz
```
or for the entire pipeline
```
  monospace run --graphviz
```

```
monospace run [options] task1 [task2...] [flags]
```

### Options

```
  -g, --graphviz                 Open a graph visualisation of the task execution plan instead of executing it
  -h, --help                     help for run
  -p, --project-filter strings   Filter projects by name
```

### Options inherited from parent commands

```
  -C, --no-color   Disable color output mode (you can also use env var NO_COLOR)
```

### SEE ALSO

* [monospace](monospace.md)	 - monospace is not monorepo

###### Auto generated by spf13/cobra on 21-Mar-2023