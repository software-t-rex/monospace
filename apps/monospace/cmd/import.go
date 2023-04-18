/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package cmd

import (
	"fmt"

	"github.com/software-t-rex/monospace/gomodules/colors"
	"github.com/software-t-rex/monospace/mono"

	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import projectName repoUrl",
	Short: "Import an external project repository",
	Long: `Import an 'external' project repository:

Import behave like the create command but instead of creating a new project,
it will clone a remote 'external' repository into the current monospace.`,
	Example: `  monospace import packages/fancylib git@github.com:username/fancylib.git`,
	Args: func(cmd *cobra.Command, args []string) error {
		// Optionally run one of the validators provided by cobra
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		if !mono.ProjectIsValidName(args[0]) {
			return fmt.Errorf(colors.Error("'%s'")+" is not a valid project name", args[0])
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return mono.ProjectsGetAllNameOnly(), cobra.ShellCompDirectiveDefault
	},
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		mono.ProjectCreate(args[0], args[1], "")
	},
}

func init() {
	RootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
