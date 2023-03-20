/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package cmd

import (
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

Example:
monospace run --project-filter modules/mymodule --project-filter modules/myothermodule test
or more concise
monospace run -p modules/mymodule,modules/myothermodule test


`,
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
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		config := utils.CheckErrOrReturn(app.ConfigGet())

		// handle project aliases in filters
		filter := utils.CheckErrOrReturn(cmd.Flags().GetStringSlice("project-filter"))
		if len(config.Aliases) > 0 {
			for i, f := range filter {
				alias := config.Aliases[f]
				if alias != "" {
					filter[i] = alias
				}
			}
		}
		if graphviz, _ := cmd.Flags().GetBool("graphviz"); graphviz {
			tasks.OpenGraphviz(args, filter)
			return
		}
		tasks.Run(args, filter)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().StringSliceP("project-filter", "p", []string{}, "Filter projects by name")
	runCmd.Flags().BoolP("graphviz", "g", false, "Open a graph visualisation of the task execution plan instead of executing it")
	runCmd.RegisterFlagCompletionFunc("project-filter", utils.CompleteProjectFilter)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
