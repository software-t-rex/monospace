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
	"github.com/software-t-rex/monospace/mono"
	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra"
)

// aliasesCmd represents the aliases command
var aliasesCmd = &cobra.Command{
	Use:   "aliases",
	Short: "List, Add or Remove project aliases",
	Long: `List, Add or Remove project aliases:
Filtering on projects by their full project path can sometimes be a bit cumbersome.
You can add aliases to a project by using the 'alias add' command

Aliases should only contain letters, numbers, underscores and hyphens and start
with a letter or an underscore.

without arguments this command will return the list of current aliases`,
	Example: `  monospace aliases list
  monospace aliases add packages/mypackage myalias
  monospace aliases remove myalias`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{"add", "remove", "list"}, cobra.ShellCompDirectiveDefault
		}
		if len(args) == 1 {
			switch args[0] {
			case "add":
				return mono.ProjectsGetAllNameOnly(), cobra.ShellCompDirectiveDefault
			case "remove":
				return mono.ProjectsGetAliasesNameOnly(), cobra.ShellCompDirectiveDefault
			default:
				return nil, cobra.ShellCompDirectiveError
			}
		}
		if len(args) == 2 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveDefault
	},
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		if len(args) == 0 {
			args = append(args, "list")
		}
		switch args[0] {
		case "ls":
			fallthrough
		case "list":
			config := utils.CheckErrOrReturn(app.ConfigGet())
			if len(config.Aliases) == 0 {
				fmt.Println("No aliases defined")
				os.Exit(0)
			}
			for alias, projectName := range config.Aliases {
				fmt.Printf("%s: %s\n", alias, projectName)
			}
		case "add":
			if len(args) != 3 {
				utils.Exit("Bad number of arguments, try: monospace add project/path alias")
			}
			utils.CheckErr(app.ConfigAddProjectAlias(args[1], args[2], true))
		case "remove":
			if len(args) != 2 {
				utils.Exit("Bad number of arguments, try: monospace remove alias")
			}
			utils.CheckErr(app.ConfigRemoveProjectAlias(args[1], true))
		default:
			utils.PrintError(fmt.Errorf("unknown command aliases %s", args[0]))
			cmd.Help()
			os.Exit(1)
		}

	},
}

func init() {
	RootCmd.AddCommand(aliasesCmd)
}
