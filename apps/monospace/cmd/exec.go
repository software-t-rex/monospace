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
	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/software-t-rex/monospace/mono"
	"github.com/software-t-rex/monospace/tasks"
	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec [options] -- cmd [args...]",
	Short: "execute given command in each project directory",
	Long: `execute given command in each project directory concurrently.

` + ui.ApplyStyle("execute options and command options must be separated by '--'", ui.Bold) + `
You can restrict the command to one or more projects using flag --project-filter.`,
	Example: `  monospace exec --project-filter modules/mymodule --project-filter modules/myothermodule -- ls -la
  # or more concise
  monospace exec -p modules/mymodule,modules/myothermodule -- ls -la
  # create a branch on all git projects at once (including root)
  monospace exec --git -r -- git checkout -b my-new-branch
  # fetching only external projects
  monospace exec --external -- git fetch`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		app.PopulateEnv(nil)
		config := utils.CheckErrOrReturn(app.ConfigGet())
		cmdBin := args[0]
		cmdArgs := args[1:]
		outputMode := FlagGetOutputMode(cmd, config.PreferredOutputMode)
		filterGitOnly := FlagGetBool(cmd, "git")
		filterExternal := FlagGetBool(cmd, "external")
		filterInternal := FlagGetBool(cmd, "internal")
		filterLocal := FlagGetBool(cmd, "local")
		includeRoot := FlagGetBool(cmd, "include-root")
		hasKindFilter := filterExternal || filterInternal || filterLocal

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
			// check project match --git/--external/--internal/--local flags
			if filterGitOnly && !p.IsGit() {
				continue
			}
			if hasKindFilter && !((filterExternal && p.Kind == mono.External) ||
				(filterInternal && p.Kind == mono.Internal) ||
				(filterLocal && p.Kind == mono.Local) ||
				(includeRoot && p.Kind == mono.Root)) {
				continue
			}
			os.Setenv("MONOSPACE_PROJECT_KIND", project.Kind.String())
			os.Setenv("MONOSPACE_PROJECT_NAME", project.Name)
			os.Setenv("MONOSPACE_PROJECT_PATH", project.Path())
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
	execCmd.Flags().Bool("git", false, "Execute command in git projects only (root has to be include with -r)")
	execCmd.Flags().Bool("external", false, "Execute command in all external projects")
	execCmd.Flags().Bool("internal", false, "Execute command in all internal projects (root has to be include with -r)")
	execCmd.Flags().Bool("local", false, "Execute command in all local projects (root has to be include with -r)")

}
