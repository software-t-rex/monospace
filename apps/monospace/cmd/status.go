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
	"github.com/software-t-rex/monospace/gomodules/colors"
	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Aliases: []string{"st"},
	Use:     "status",
	Short:   "Return aggregated git status information for all repositories in the monospace",
	Long: `Return aggregated git status information for all repositories in the monospace:

You can pass args to git status by separating them with double hyphen '--'
` + underline("Example:") + `
  monospace status -- --porcelain

` + italic(`monospace st`) + ` is an alias of this command but will add the --short and --branch
flags to the underlying git status command.`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		utils.CheckErr(utils.MonospaceChdir())

		// add --short flag if called with 'st' alias
		if cmd.CalledAs() == "st" {
			args = append(args, "--short", "--branch")
		}

		isShort := utils.SliceContains(args, "--short") || utils.SliceContains(args, "--porcelain")
		nameStyle := colors.Style(colors.Bold)
		projects := utils.ProjectsGetAll()
		internals := []string{}
		executor := jobExecutor.NewExecutor().WithProgressBarOutput(
			10, false, string(colors.BgBlack)+string(colors.BrightGreen),
		)
		executor.AddNamedJobCmd("monospace", getStatusCommand("", args))
		for _, p := range projects {
			project := p
			if p.Kind == utils.Internal {
				if !isShort {
					internals = append(internals, p.Name)
				}
				continue
			}
			executor.AddNamedJobCmd(project.Name, getStatusCommand(project.Name, args))
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getStatusCommand(path string, args []string) *exec.Cmd {
	args = append([]string{"status"}, args...)
	if colors.ColorEnabled() {
		args = append([]string{"-c", "color.ui=always"}, args...)
	}
	cmd := exec.Command("git", args...)
	if path != "" {
		cmd.Dir = path
	}
	return cmd
}
