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

	"github.com/software-t-rex/go-jobExecutor"
	"github.com/software-t-rex/monospace/colors"
	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Aliases: []string{"st"},
	Use:     "status",
	Short:   "Return aggregated git status information for all repositories in the monospace",
	Long: `Return aggregated git status information for all repositories in the monospace
`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		utils.CheckErr(utils.MonospaceChdir())
		projects := utils.ProjectsGetAll()
		executor := jobExecutor.NewExecutor().WithProgressBarOutput(len(projects), false, string(colors.BgBlack)+string(colors.BrightGreen))
		executor.AddNamedJobCmd("'monospace' git status", getStatusCommand(""))
		for _, p := range projects {
			project := p
			if p.Kind == utils.Internal {
				executor.AddNamedJobFn(fmt.Sprintf("'%s' is internal project -> skipped", project.Name), func() (string, error) {
					return "", nil
				})
				continue
			}
			executor.AddNamedJobCmd(fmt.Sprintf("'%s' git status", project.Name), getStatusCommand(project.Name))
		}
		executor.OnJobsDone(func(jobs jobExecutor.JobList) {
			output := []string{}
			for _, job := range jobs {
				if job.Res != "" {
					output = append(output, fmt.Sprintf("%s:\n%s", job.Name(), job.Res))
				} else {
					output = append(output, fmt.Sprintf("%s\n", job.Name()))
				}
			}
			fmt.Print(strings.Join(output, "\n"))
		})
		executor.Execute()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getStatusCommand(path string) *exec.Cmd {
	args := []string{"status", "--short"}
	if colors.ColorEnabled() {
		args = append([]string{"-c", "color.ui=always"}, args...)
	}
	cmd := exec.Command("git", args...)
	if path != "" {
		cmd.Dir = path
	}
	return cmd
}
