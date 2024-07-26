## monospace tasks import

Import scripts entries from projects package.json.

### Synopsis

Import scripts entries from projects package.json file as monospace pipeline task.

You can either pass a list of scripts to import using the syntax
'projectName#scriptName', or you can let the command prompt you for scripts to
import. You can narrow choices presented to you by using the --project-filter
flag. 

If running in interactive mode you will be able to edit the task before adding it
to the pipeline.

Pipeline will be checked for cyclic dependencies before saving any changes.


```
monospace tasks import [projectName#scriptName]... [flags]
```

### Options

```
  -h, --help                         help for import
  -y, --no-interactive               Prevent any interactive prompts by choosing default values (not always yes)
  -p, --project-filter strings       Filter projects by name
                                     This is like 'whitelisting' project in the list
                                     You can use 'root' for monospace root directory
  -P, --project-filter-out strings   Filter out by name
                                     Exclude projects from the list (blacklisting)
```

### Options inherited from parent commands

```
  -C, --no-color   Disable color output mode (you can also use env var NO_COLOR)
```

### SEE ALSO

* [monospace tasks](monospace_tasks.md)	 - List tasks defined in monospace.yml pipeline.

###### Auto generated by spf13/cobra on 19-Jun-2024