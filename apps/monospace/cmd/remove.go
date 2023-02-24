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

ProjectName is relative path of project from the root of the monospace.

	It will:
- remove the project from the .monospace.yml config
- remove the project from the monospace .gitignore for non 'internal' projects
- delete the corresponding directory if --rmdir or -r flag is set

` + underline("First argument:") + ` is the path of the project to remove relative to monospace root.
` + underline("example:") + " " + italic("monospace remove apps/my-app") + `
`,
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return utils.ProjectsGetAllNameOnly(), cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound()
		utils.ProjectRemove(args[0], true, !flagRemoveRmDir)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolVarP(&flagRemoveRmDir, "rmdir", "r", false, "Remove the project directory without confirm")
	// Here you will define your flags and configuration settings.

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
