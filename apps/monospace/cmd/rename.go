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

	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra"
)

// renameCmd represents the rename command
var renameCmd = &cobra.Command{
	Use:   "rename projectName newProjectName",
	Short: "Rename a project",
	Long: `This will rename a project inside the monospace:
will update the monospace gitignore and .monospaced.yml files accordingly.
`,
	Args: cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveDefault
		}
		return utils.ProjectsGetAllNameOnly(), cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		utils.CheckErr(os.Chdir(utils.MonospaceGetRoot()))

		oldName := args[0]
		newName := args[1]

		if !utils.MonospaceHasProject(oldName) {
			utils.Exit(fmt.Sprintf("Unkwown project %s", oldName))
		} else if utils.FileExistsNoErr(newName) {
			utils.Exit(fmt.Sprintf("%s already exists", oldName))
		}

		project := utils.CheckErrOrReturn(utils.ProjectGetByName(oldName))

		utils.CheckErr(utils.ProjectRemoveFromConfig(project, true))
		utils.CheckErr(utils.MonospaceAddProject(newName, project.RepoUrl))

		if project.Kind != utils.Internal {
			utils.CheckErr(utils.ProjectRemoveFromGitignore(project, true))
			utils.CheckErr(utils.MonospaceAddProjectToGitignore(newName))
		}

		utils.CheckErr(os.Rename(oldName, newName))

		fmt.Println(utils.Success("Done"))
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// renameCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// renameCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
