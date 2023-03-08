/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [options] task1 [task2...]",
	Short: "Run given tasks in each project directory",
	Long: `Run given command in each project directory concurrently.

You can restrict the tasks execution to one or more projects using the --project-filter
flag.
Example:
monospace run --project-filter modules/mymodule --project-filter modules/myothermodule test
or more concise
monospace run -p modules/mymodule,modules/myothermodule test
`,
	// ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// 	return nil, cobra.ShellCompDirectiveNoFileComp
	// },
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		CheckConfigFound(true)
		monoRoot := utils.MonospaceGetRoot()
		config := utils.CheckErrOrReturn(app.ConfigGet())
		// utils.CheckErr(utils.MonospaceChdir())

		// handle project aliases
		filter := utils.CheckErrOrReturn(cmd.Flags().GetStringSlice("project-filter"))
		if len(config.Aliases) > 0 {
			for i, f := range filter {
				alias := config.Aliases[f]
				if alias != "" {
					filter[i] = alias
				}
			}
		}
		var projects []utils.Project
		if len(filter) == 0 {
			projects = utils.ProjectsGetAll()
		} else {
			projects = utils.SliceFilter(utils.ProjectsGetAll(), func(p utils.Project) bool {
				return utils.SliceContains(filter, p.Name)
			})
		}
		// // create dependency graph
		// projectsTasks := make(map[string]PTask)
		// for _, p := range projects {
		// 	projectsTasks[p.Name] = PTask{
		// 		Name:           p.Name,
		// 		HasPackageJson: utils.FileExistsNoErr(filepath.Join(monoRoot, p.Name, "package.json")),
		// 	}
		// 	projectsTasks[p.Name].LoadInfos()
		// }
		// utils.Dump(projectsTasks)

		return
		// @TODO add a task planner
		// get tasks from pipeline
		// get tasks from package.json
		// populate dependsOn either by pipeline or package.json
		// topo ordering
		// make stages
		executor := utils.NewTaskExecutor()
		for _, p := range projects {
			project := p
			executor.AddNamedJobFn(fmt.Sprintf("%s: %s", utils.Bold(project.Name), strings.Join(args, " ")), func() (string, error) {
				cmd := exec.Command("ls")
				cmd.Dir = filepath.Join(monoRoot, project.Name)
				// cmd.Env = append(os.Environ(), "FORCE_COLOR=1", "FORCE_COLORS=1", "TERM=color", "COLORTERM=1", "_tty_out=True")
				resBytes, err := cmd.CombinedOutput()
				return string(resBytes), err
			})
		}
		executor.Execute()
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().StringSliceP("project-filter", "p", []string{}, "Filter projects by name")
	runCmd.RegisterFlagCompletionFunc("project-filter", utils.CompleteProjectFilter)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
