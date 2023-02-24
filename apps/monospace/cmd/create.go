/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/
package cmd

import (
	"errors"
	"fmt"

	"github.com/software-t-rex/monospace/colors"
	"github.com/software-t-rex/monospace/utils"

	"github.com/spf13/cobra"
)

var bold = colors.Style(colors.Bold)
var underline = colors.Style(colors.Underline)
var italic = colors.Style(colors.Italic)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create internal|local projectName",
	Short: "Create a new internal or local project",
	Long: `Create a 'local' or 'internal' project:

` + underline("First argument:") + bold(" MUST") + ` be 'internal' or 'local':
- internal will embbed the project in the monospace repository.
- local will init a new git repository within the monospace repository
  You will have to set a remote later to make it an external project instead

` + underline("Second argument:") + ` 'ProjectName' is the relative path to the project from the root
of the monospace repository. Example for a new application: apps/my-new-app

` + underline("Example:") + `
` + italic(" monospace create local apps/my-new-app") + `
` + italic(" monospace create local apps/my-new-js-app --type=js") + `
` + italic(" monospace create local gomodules/my-go-module -t go") + `

If you want to ` + bold(`import`) + ` an existing ` + bold("'external'") + ` git repository into the monospace
you should look at the ` + italic("monospace import") + ` command instead
`,

	Args: func(cmd *cobra.Command, args []string) error {
		// Optionally run one of the validators provided by cobra
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		if args[0] != "internal" && args[0] != "local" {
			return errors.New("you must specify either " + colors.Error("'internal'") + " or " + colors.Error("'local'") + " as first argument")
		}
		if !utils.ProjectIsValidName(args[1]) {
			return fmt.Errorf(colors.Error("'%s'")+" is not a valid project name", args[1])
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveFilterDirs
		}
		return []string{"local", "internal"}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound()
		if !utils.Contains([]string{"", "go", "js"}, flagCreatePType) {
			utils.Exit("Unknown project type '" + flagCreatePType + "'")
		}
		utils.ProjectCreate(args[1], args[0], flagCreatePType)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&flagCreatePType, "type", "t", "", "type of project to create for now only 'go' and 'js' are supported")
	createCmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"go", "js"}, cobra.ShellCompDirectiveNoFileComp
	})
	//flagCreateSkipPM
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
