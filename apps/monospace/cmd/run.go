/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/tasks"
	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [options] task1 [task2...]",
	Short: "Run given tasks in each project directory",
	Long: `Run given tasks defined in monospace.yml in each project directory concurrently. 

You can restrict the tasks execution to one or more projects
using the --project-filter flag.
You can pass additional arguments to tasks separating them with a double hyphen.

you can get a dependency graph of tasks to run by using the --graphviz flag.
It will output the dot representation in your terminal and open your browser
for visual online rendering.

A circular dependency check will be performed before the execution starts.`,
	Example: `  monospace run --project-filter modules/mymodule --project-filter modules/myothermodule test
  # or more concise
  monospace run -p modules/mymodule,modules/myothermodule test
  monospace run -p modules/mymodule,modules/myothermodule test -- additionalArg=value
  # run tasks on monospace root only
  monospace run -p root task
  # get some dependency graph
  monospace run task --graphviz
  # or for the entire pipeline
  monospace run --graphviz`,

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		config, err := app.ConfigGet()
		if err != nil {
			return nil, cobra.ShellCompDirectiveDefault
		}
		taskNames := []string{}
		for task := range config.Pipeline {
			if !strings.ContainsRune(task, '#') {
				taskNames = append(taskNames, task)
			} else {
				projectTask := strings.Split(task, "#")
				taskNames = append(taskNames, projectTask[1])
			}
		}
		return taskNames, cobra.ShellCompDirectiveDefault
	},
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		config := utils.CheckErrOrReturn(app.ConfigGet())
		graphviz := utils.CheckErrOrReturn(cmd.Flags().GetBool("graphviz"))

		// special cases either full graphviz or no task to run
		if len(args) == 0 {
			if graphviz {
				tasks.OpenGraphvizFull()
				return
			}
			utils.PrintError(errors.New("missing task to run"))
			cmd.Help()
			os.Exit(1)
		}

		filteredProjects := FlagGetFilteredProjects(cmd)
		// remove additional args from the command and populate additional args as job parameters
		additionalArgs := splitAdditionalArgs(&args)
		taskList := tasks.PrepareTaskList(args, filteredProjects)

		// we don't need to bother with ouput mode for graphviz output
		if graphviz {
			tasks.OpenGraphviz(taskList)
			return
		}

		// we make extra job to determine best output mode for the task
		flagOuputMode := utils.CheckErrOrReturn(cmd.Flags().GetString("output-mode"))
		var outputMode string
		if flagOuputMode != "" {
			// flag is set so we use it
			outputMode = flagOuputMode
		} else {
			// there's no user value we try to guess
			for _, taskArgName := range args {
				var tmpOutputMode string
				for taskStdName, task := range taskList.List {
					if strings.HasSuffix(taskStdName, "#"+taskArgName) && task.TaskDef.OutputMode != "" {
						if tmpOutputMode == "" {
							tmpOutputMode = task.TaskDef.OutputMode
						} else if tmpOutputMode != task.TaskDef.OutputMode { // ignore ambiguous output mode
							tmpOutputMode = ""
							break
						}
					}
				}
				// we found a value so we use it and stop looking
				if tmpOutputMode != "" {
					outputMode = tmpOutputMode
					break
				}
			}
		}

		if outputMode == "" { // we still don't have a value so we use global default
			outputMode = FlagGetOutputMode(cmd, config.PreferredOutputMode)
		}

		tasks.Run(taskList, additionalArgs, outputMode)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	FlagAddProjectFilter(runCmd, true)
	FlagAddOutputMode(runCmd)
	runCmd.Flags().BoolP("graphviz", "g", false, "Open a graph visualisation of the task execution plan instead of executing it")
}
