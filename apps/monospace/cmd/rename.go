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

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/software-t-rex/monospace/mono"
	"github.com/spf13/cobra"
)

// renameCmd represents the rename command
var renameCmd = &cobra.Command{
	Use:   "rename projectName newProjectName",
	Short: "Rename a project",
	Long: `This will rename a project inside the monospace:
will update the monospace gitignore and .monospace.yml files accordingly.
It will also rename the project pipeline tasks and update tasks depending on them.`,
	Args: cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveDefault
		}
		return mono.ProjectsGetAllNameOnly(), cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		utils.CheckErr(os.Chdir(mono.SpaceGetRoot()))

		oldName := args[0]
		newName := args[1]

		if !mono.SpaceHasProject(oldName) {
			utils.Exit(fmt.Sprintf("Unkwown project %s", oldName))
		} else if utils.FileExistsNoErr(newName) {
			utils.Exit(fmt.Sprintf("%s already exists", oldName))
		} else if !mono.ProjectIsValidName(newName) {
			utils.Exit(fmt.Sprintf("%s is not a valid project name", newName))
		}

		project := utils.CheckErrOrReturn(mono.ProjectGetByName(oldName))

		// ConfigRenameProject rename pipeline tasks and their references so no task is lost
		utils.CheckErr(app.ConfigRenameProject(project.Name, newName, true))

		if project.Kind != mono.Internal {
			utils.CheckErr(mono.ProjectRemoveFromGitignore(project, true))
			utils.CheckErr(mono.SpaceAddProjectToGitignore(newName))
		}

		utils.CheckErr(os.Rename(oldName, newName))

		fmt.Println(theme.Success("Done"))
	},
}

func init() {
	RootCmd.AddCommand(renameCmd)
}
