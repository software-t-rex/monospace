/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package cmd

import (
	"os/exec"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/software-t-rex/monospace/tasks"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var gitCmd = &cobra.Command{
	Use:   "git <git-command> [git-args...] [-- git-flags...]",
	Short: "Run given git command in each repository contained in this monospace",
	Long: `Run given git command in each repository contained in this monospace (including root) concurrently. 

You can ommit execution in one or more projects using --project-filter[-out] flags.

Output mode is grouped by default, and git flags must be passed after a double hyphen.
`,
	Example: `  monospace git status -- --porcelain
# excluding root
	monospace -P root git status -- --ignored
`,

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, err := app.ConfigGet()
		if err != nil {
			return nil, cobra.ShellCompDirectiveDefault
		}
		if len(args) == 1 { //@todo offer bettre completion than this
			comp := []string{
				"status",
				"fetch",
				"pull",
				"push",
				"checkout",
				"commit",
				"add",
				"reset",
				"rebase",
				"merge",
				"tag",
				"branch",
				"diff",
				"log",
				"show",
				"stash",
				"clean",
				"grep",
				"blame",
				"bisect",
				"reflog",
				"cherry-pick",
				"revert",
				"apply",
				"format-patch",
			}
			return comp, cobra.ShellCompDirectiveDefault
		}
		return nil, cobra.ShellCompDirectiveDefault
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		app.PopulateEnv(nil)
		config := utils.CheckErrOrReturn(app.ConfigGet())
		outputMode := FlagGetOutputMode(cmd, "grouped")
		aliases := config.GetProjectsAliases()
		projects := FlagGetFilteredProjectsWithRoot(cmd, config)
		executor := tasks.NewExecutor(outputMode)
		if ui.EnhancedEnabled() { // add colors to git commands if color is enabled
			args = append([]string{"-c", "color.ui=always"}, args...)
		}
		for _, p := range projects {
			if !p.IsGit() {
				continue
			}
			project := p
			cmd := exec.Command("git", args...)
			cmd.Dir = project.Path()
			switch outputMode {
			case "interleaved":
				alias, hasAlias := aliases[project.Name]
				executor.AddNamedJobCmd(utils.If(hasAlias, alias, project.Name), cmd)
			default:
				executor.AddNamedJobCmd(project.StyledString(), cmd)
			}
		}
		executor.Execute()
	},
}

func init() {
	RootCmd.AddCommand(gitCmd)
	FlagAddProjectFilter(gitCmd, true)
	FlagAddOutputMode(gitCmd)
}
