/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package cmd

import (
	"os/exec"

	"github.com/software-t-rex/monospace/gomodules/colors"
	"github.com/spf13/cobra"
)

// repoCmd represents the git command
var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Perform operations agains repositories in the monospace",
	Long: `Perform operations agains repositories in the monospace

monospace repo check
`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	CheckConfigFound(true)
	// 	monoRoot := utils.MonospaceGetRoot()
	// 	config := utils.CheckErrOrReturn(app.ConfigGet())
	// 	outputMode := ValidateFlagOutputMode(cmd)

	// 	switch args[0] {
	// 	case "check":
	// 	}

	// 	projects := utils.ProjectsGetAll()
	// 	internals := []string{}
	// 	executor := NewTaskExecutor(outputMode)

	// 	executor.AddNamedJobCmd("monospace", getStatusCommand("", args))

	// },
}

func init() {
	RootCmd.AddCommand(repoCmd)
	AddPersistentFlagProjectFilter(repoCmd)
	AddPersistentFlagOutputMode(repoCmd)
}

func getGitCommand(path string, args []string) *exec.Cmd {
	if colors.ColorEnabled() {
		args = append([]string{"-c", "color.ui=always"}, args...)
	}
	cmd := exec.Command("git", args...)
	if path != "" {
		cmd.Dir = path
	}
	return cmd
}
