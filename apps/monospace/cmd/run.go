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
	Long: `Run given command in each project directory concurrently.

You can restrict the tasks execution to one or more projects
using the --project-filter flag.
You can pass additional arguments to tasks separating them with a double hyphen.

` + underline("Example:") + `
` + italic(`  monospace run --project-filter modules/mymodule --project-filter modules/myothermodule test`) + `
or more concise
` + italic(`  monospace run -p modules/mymodule,modules/myothermodule test`) + `
` + italic(`  monospace run -p modules/mymodule,modules/myothermodule test -- additionalArg=value `) + `




you can get a dependency graph of tasks to run by using the --graphviz flag.
It will output the dot representation in your terminal and open your browser
for visual online rendering.

` + italic(`  monospace run task --graphviz`) + `
or for the entire pipeline
` + italic(`  monospace run --graphviz`),

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
		outputMode := FlagGetOutputMode(cmd)
		graphviz := utils.CheckErrOrReturn(cmd.Flags().GetBool("graphviz"))
		filteredProjects := FlagGetFilteredProjects(cmd, false)

		if len(args) == 0 && !graphviz {
			utils.PrintError(errors.New("missing task to run"))
			cmd.Help()
			os.Exit(1)
		}
		// remove additional args from the command and populate additional args as job parameters
		additionalArgs := splitAdditionalArgs(&args)

		CheckConfigFound(true)
		config := utils.CheckErrOrReturn(app.ConfigGet())
		if outputMode == "" && config.PreferedOutputMode != "" {
			outputMode = config.PreferedOutputMode
		}

		if graphviz && len(args) == 0 {
			tasks.OpenGraphvizFull()
			return
		}

		if graphviz {
			tasks.OpenGraphviz(args, filteredProjects)
			return
		}

		tasks.Run(args, filteredProjects, additionalArgs, outputMode)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	FlagAddProjectFilter(runCmd)
	FlagAddOutputMode(runCmd)
	runCmd.Flags().BoolP("graphviz", "g", false, "Open a graph visualisation of the task execution plan instead of executing it")
}
