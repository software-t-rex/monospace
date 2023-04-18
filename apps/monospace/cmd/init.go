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

	"github.com/software-t-rex/monospace/git"
	"github.com/software-t-rex/monospace/gomodules/scaffolders"
	"github.com/software-t-rex/monospace/mono"
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
  (.monospace/monospace.yml, .npmrc, .gitignore, go.work if go installed detected)
- init a git repository if not already initialized

each of these steps won't overwrite existing files if any`,
	Example: `  monospace init
  monospace init path/to/new-monospace`,
	Args: func(cmd *cobra.Command, args []string) error {
		// no more than one argument which shoult be path to the new monospace
		if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		noInteractive := FlagGetNoInteractive(cmd)
		if noInteractive {
			utils.CheckErr(os.Setenv("MONOSPACE_NO_INTERACTIVE", "1"))
		}
		var parentMonospace string
		if len(args) == 1 {
			wd, err := filepath.Abs(args[0])
			utils.CheckErr(err)
			parentMonospace = mono.SpaceGetRootForPath(wd)
		} else {
			parentMonospace = mono.SpaceGetRootNoCache()
		}

		// check path is not inside a monospace directory
		if parentMonospace != "" {
			utils.Exit(fmt.Sprintf("'%s' is already a monospace directory", parentMonospace))
		}

		// if path is given create the directory and cd into it
		if len(args) == 1 {
			utils.CheckErr(utils.MakeDir(args[0]))
			utils.CheckErr(os.Chdir(args[0]))
		}

		// scaffold monospace
		fmt.Println("initialize git repository")
		utils.CheckErr(git.GitInit("./", true))

		fmt.Println("initialize monospace")
		utils.CheckErr(scaffolders.Monospace())

		fmt.Println(utils.Success("Monospace successfully initialized."))
		if len(args) == 1 {
			fmt.Printf("%s is ready for work\n", args[0])
		}
	},
}

func init() {
	// @todo add prompt for preferred js package manager and go.mod default prefix
	RootCmd.AddCommand(initCmd)
	FlagAddNoInteractive(initCmd)
}
