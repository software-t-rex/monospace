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
	"path/filepath"

	"github.com/software-t-rex/monospace/scaffolders"
	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new monospace",
	Long: `Initialize a new monospace
It will perform the following steps:
- create some files if they are not present in the directory
	(.monospace.yml, .npmrc, .gitignore, go.work if go installed detected)
- init a git repository if not already initialized
each of these steps won't overwrite existing files if any
` + utils.Underline("usage:") + `
monospace init
monospace init path/to/new-monospace
`,
	Args: func(cmd *cobra.Command, args []string) error {
		// no more than one argument which shoult be path to the new monospace
		if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		noInteractive := utils.CheckErrOrReturn(cmd.Flags().GetBool("no-interactive"))
		if noInteractive {
			utils.CheckErr(os.Setenv("MONOSPACE_NO_INTERACTIVE", "1"))
		}
		var parentMonospace string
		if len(args) == 1 {
			wd, err := filepath.Abs(args[0])
			utils.CheckErr(err)
			parentMonospace = utils.MonospaceGetRootForPath(wd)
		} else {
			parentMonospace = utils.MonospaceGetRootNoCache()
		}

		// check path is not inside a monospace directory
		if parentMonospace != "" {
			utils.Exit(fmt.Sprintf("'%s' is already a monospace directory", parentMonospace))
		}

		if len(args) == 1 {
			utils.CheckErr(utils.MakeDir(args[0]))
			utils.CheckErr(os.Chdir(args[0]))
		}

		// scaffold monospace
		utils.CheckErr(scaffolders.Monospace())
		utils.CheckErr(utils.GitInit("./", true))
		fmt.Println(utils.Success("Monospace successfully initialized."))
		if len(args) == 1 {
			fmt.Printf("%s is ready for work\n", args[0])
		}
	},
}

func init() {
	// @todo add prompt for prefered js package manager and go.mod default prefix
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("no-interactive", "y", false, "Prevent any interactive prompts")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
