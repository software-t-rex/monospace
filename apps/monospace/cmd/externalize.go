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

// externalizeCmd represents the externalize command
var externalizeCmd = &cobra.Command{
	Use:   "externalize projectName [repoUrl]",
	Short: "Turn an 'internal' project into an 'external' project",
	Long: `Turn an 'internal' project into an 'external' project.

Here what this command will really do:
- Check that the project path is clean
	- Propose to stash uncommited changes (--no-interactive will say yes)
- Use git subtree split to extract the project directory and related history
- Remove extracted directory from the monospace
- Init a new repository in the project directory
	- optionally set initial branch name (--initial-branch=name, default to 'master')
	- optionally set the new repo origin if repoUrl was given
- merge the extracted branch from root repository into the newly created repo
	- optionally push -u to the new repo origin (--push)
- Add the projectName to the monospace .gitignore
- Mark the project as external (if repoUrl is set) or local in monospace.yml
- remove the temporary subtree branch from root repository if there's no error

You can then review changes in the monospace root repository and commit them.

` + utils.Warning(`Beware that the operation will remove all files in the project directory before re-creating them.
You should check that there's no untracked files before proceeding as they will be lost.`),
	Example: `monospace externalize packages/osslib
monospace externalize packages/osslib git@github.com:user/osslib.git --push --initial-branch=main --commit
`,
	Args: cobra.RangeArgs(1, 2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return utils.ProjectsGetAllNameOnly(), cobra.ShellCompDirectiveNoFileComp
	},

	Run: func(cmd *cobra.Command, args []string) {

		CheckConfigFound(true)
		monoRoot := utils.MonospaceGetRoot()
		projectName := args[0]
		flagPush := utils.CheckErrOrReturn(cmd.Flags().GetBool("push"))
		initBranch := utils.CheckErrOrReturn(cmd.Flags().GetString("initial-branch"))
		noConfirm := FlagGetNoInteractive(cmd)
		opts := utils.GitExternalizeOptions{
			PushOrigin: flagPush,
		}
		if initBranch != "" {
			opts.InitialBranch = initBranch
		}
		if len(args) == 2 {
			opts.Origin = args[1]
		} else if opts.PushOrigin {
			utils.Exit("you must provide a repo url when using --push")
		}

		isClean := utils.GitIsClean(monoRoot, projectName)
		if !isClean {
			if noConfirm {
				opts.AllowStash = true
			} else {
				fmt.Println(utils.Warning("This project directory is not clean, monospace will stash the directory"))
				fmt.Println(utils.Warning("and attempt to restore it afterwards, but this can lead to data lost."))
				opts.AllowStash = utils.Confirm("are you sure you want to continue ?", false)
				if !opts.AllowStash {
					fmt.Println("Exiting on user request")
					os.Exit(0)
				}
			}
		}
		utils.CheckErr(utils.GitExternalize(monoRoot, projectName, opts))

		fmt.Println(utils.Success("Externalization done"))
	},
}

func init() {
	RootCmd.AddCommand(externalizeCmd)
	externalizeCmd.Flags().StringP("initial-branch", "b", "", "set the default branch name (default to your git default setting)")
	externalizeCmd.Flags().BoolP("push", "p", false, "push initial branch and set upstream to origin")
	FlagAddNoInteractive(externalizeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// externalizeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// externalizeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
