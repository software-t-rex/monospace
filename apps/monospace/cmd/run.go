/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/software-t-rex/monospace/colors"
	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [options] -- cmd [args...]",
	Short: "Run given command in each project directory",
	Long: `Run given command in each project directory in parallel.
run options and the command must be separated by '--'

You can restrict the command to one or more projects using the --project-filter
flag.
Example:
monospace run --project-filter modules/mymodule --project-filter modules/myothermodule -- ls -la
or more concise
monospace run -p modules/mymodule -p modules/myothermodule -- ls -la
`,
	// ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// 	return nil, cobra.ShellCompDirectiveNoFileComp
	// },
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		CheckConfigFound(true)
		monoRoot := utils.MonospaceGetRoot()
		cmdBin := args[0]
		cmdArgs := args[1:]
		if cmdBin != "" && cmdBin[0] == '.' {
			cmdBin = filepath.Join(utils.CheckErrOrReturn(os.Getwd()), cmdBin)
		}
		if cmdBin == "git" && colors.ColorEnabled() { // add colors to git commands if color is enabled
			cmdArgs = append([]string{"-c", "color.ui=always"}, cmdArgs...)
		}
		// utils.CheckErr(utils.MonospaceChdir())
		projects := utils.ProjectsGetAll()
		filter := utils.CheckErrOrReturn(cmd.Flags().GetStringArray("project-filter"))
		executor := utils.NewTaskExecutor()
		for _, p := range projects {
			project := p
			if !filterProject(project.Name, filter) {
				continue
			}
			executor.AddNamedJobFn(fmt.Sprintf("%s: %s", utils.Bold(project.Name), strings.Join(args, " ")), func() (string, error) {
				cmd := exec.Command(cmdBin, cmdArgs...)
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
	runCmd.Flags().StringArrayP("project-filter", "p", []string{}, "Filter projects by name")
	runCmd.RegisterFlagCompletionFunc("project-filter", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return utils.ProjectsGetAllNameOnly(), cobra.ShellCompDirectiveDefault
	})

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func filterProject(projectName string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if filter == projectName {
			return true
		}
	}
	return false
}
