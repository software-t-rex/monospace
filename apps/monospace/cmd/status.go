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
	"strings"

	"github.com/software-t-rex/go-jobExecutor/v2"
	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/software-t-rex/monospace/mono"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Aliases: []string{"st"},
	Use:     "status",
	Short:   "Return aggregated git status information for all repositories in the monospace",
	Long: `Return aggregated git status information for all repositories in the monospace:

You can pass args to git status by separating them with double hyphen '--'`,
	Example: `  monospace status
  # passing git status args
  monospace status -- --porcelain
  # st is an alias
  # it will add the --short and --branch flags to underlying git status command.
  monospace st
  # is the same as
  monospace status -- --short --branch`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		utils.CheckErr(mono.SpaceChdir())

		// add --short flag if called with 'st' alias
		if cmd.CalledAs() == "st" {
			args = append(args, "--short", "--branch")
		}

		isShort := utils.SliceContains(args, "--short") || utils.SliceContains(args, "--porcelain")
		nameStyle := ui.NewStyler(ui.Bold)
		projects := mono.ProjectsGetAll()
		internals := []string{}
		barBg := ui.AdaptiveColor{Dark: ui.Black, Light: ui.White}.Background()
		executor := jobExecutor.NewExecutor().WithProgressBarOutput(
			40, false, ui.SGREscapeSequence(barBg, theme.Config.AccentColor.Foreground()),
		)
		executor.AddNamedJobCmd(mono.RootProject.StyledString(), getStatusCommand("", args))
		for _, p := range projects {
			project := p
			if p.Kind == mono.Internal {
				if !isShort {
					internals = append(internals, p.StyledString())
				}
				continue
			}
			executor.AddNamedJobCmd(project.StyledString(), getStatusCommand(project.Path(), args))
		}

		executor.OnJobsDone(func(jobs jobExecutor.JobList) {
			output := []string{}
			if len(internals) > 0 && !isShort {
				fmt.Print(nameStyle("Skipped internal projects:"), "\n - ", strings.Join(internals, "\n - "), "\n\n")
			}
			for _, job := range jobs {
				output = append(output, fmt.Sprintf("%s:\n%s", nameStyle(job.Name()), utils.Indent(job.Res, "  ")))
			}
			fmt.Print(strings.Join(output, ""))
		})
		executor.Execute()
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}

func getStatusCommand(path string, args []string) *exec.Cmd {
	args = append([]string{"status"}, args...)
	if ui.EnhancedEnabled() {
		args = append([]string{"-c", "color.ui=always"}, args...)
	}
	cmd := exec.Command("git", args...)
	if path != "" {
		cmd.Dir = path
	}
	return cmd
}
