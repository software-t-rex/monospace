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

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/gomodules/colors"
	"github.com/software-t-rex/monospace/tasks"
	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec [options] -- cmd [args...]",
	Short: "execute given command in each project directory",
	Long: `execute given command in each project directory concurrently.

` + utils.Bold("execute options and command options must be separated by '--'") + `
You can restrict the command to one or more projects using flag --project-filter.`,
	Example: `  monospace exec --project-filter modules/mymodule --project-filter modules/myothermodule -- ls -la
  # or more concise
  monospace exec -p modules/mymodule,modules/myothermodule -- ls -la`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		config := utils.CheckErrOrReturn(app.ConfigGet())
		cmdBin := args[0]
		cmdArgs := args[1:]
		outputMode := FlagGetOutputMode(cmd, config.PreferredOutputMode)

		if cmdBin != "" && cmdBin[0] == '.' { // make relative path relative to projects
			cmdBin = filepath.Join(utils.CheckErrOrReturn(os.Getwd()), cmdBin)
		}
		if cmdBin == "git" && colors.ColorEnabled() { // add colors to git commands if color is enabled
			cmdArgs = append([]string{"-c", "color.ui=always"}, cmdArgs...)
		}

		projects := FlagGetFilteredProjects(cmd)

		executor := tasks.NewExecutor(outputMode)
		for _, p := range projects {
			project := p
			cmd := exec.Command(cmdBin, cmdArgs...)
			cmd.Dir = project.Path()
			switch outputMode {
			case "interleaved":
				executor.AddNamedJobCmd(project.Name, cmd)
			default:
				executor.AddNamedJobCmd(fmt.Sprintf("%s: %s", project.StyledString(), strings.Join(args, " ")), cmd)
			}
		}
		executor.Execute()
	},
}

func init() {
	RootCmd.AddCommand(execCmd)
	FlagAddProjectFilter(execCmd, false)
	FlagAddOutputMode(execCmd)

}
