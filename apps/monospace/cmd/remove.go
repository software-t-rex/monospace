/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package cmd

import (
	"github.com/software-t-rex/monospace/utils"

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove projectName",
	Short: "Remove a project from the monospace",
	Long: `Remove the given project from the monospace:

It will:
- remove the project from the .monospace/monospace.yml config
- remove the project from the monospace .gitignore for non 'internal' projects
- delete the corresponding directory if --rmdir or -r flag is set

` + underline("First argument:") + ` is the relative path (from monospace root) of the project to remove.

` + underline("Example:") + `
` + italic("  monospace remove apps/my-app"),
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return utils.ProjectsGetAllNameOnly(), cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		rmDir := utils.CheckErrOrReturn(cmd.Flags().GetBool("rmdir"))
		utils.ProjectRemove(args[0], true, !rmDir)
	},
}

func init() {
	RootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolP("rmdir", "r", false, "Remove the project directory without confirm")
}
