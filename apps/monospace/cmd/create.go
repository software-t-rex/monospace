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
	"os"

	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/software-t-rex/monospace/mono"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create internal|local projectName",
	Short: "Create a new internal or local project",
	Long: `Create a 'local' or 'internal' project:

` + ui.ApplyStyle("First argument:", ui.Underline) + ui.ApplyStyle(" MUST", ui.Bold) + ` be 'internal' or 'local':
- internal will embed the project in the monospace repository.
- local will init a new git repository within the monospace repository
  You will have to set a remote later to make it an external project instead

` + ui.ApplyStyle("Second argument:", ui.Underline) + ` 'ProjectName' is the relative path to the project from the root
of the monospace repository. Example for a new application: apps/my-new-app.
It should only contain letters, numbers, underscores, slashes and hyphens
and each path part must start with a letter or an underscore.

If you want to ` + ui.ApplyStyle(`import`, ui.Bold) + ` an existing ` + ui.ApplyStyle("'external'", ui.Bold) + ` git repository into the monospace
you should look at the ` + ui.ApplyStyle("monospace import", ui.Italic) + ` command instead`,
	Example: `  monospace create local apps/my-new-app
  monospace create local apps/my-new-js-app --type=js
  monospace create local gomodules/my-go-module -t go`,

	Args: func(cmd *cobra.Command, args []string) error {
		// Optionally run one of the validators provided by cobra
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		if args[0] != "internal" && args[0] != "local" {
			return errors.New("you must specify either " + theme.Error("'internal'") + " or " + theme.Error("'local'") + " as first argument")
		}
		if !mono.ProjectIsValidName(args[1]) {
			return fmt.Errorf(theme.Error("'%s'")+" is not a valid project name", args[1])
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
		CheckConfigFound(true)
		pType := GetFlagProjectType(cmd)
		noInteractive := FlagGetNoInteractive(cmd)
		if noInteractive {
			utils.CheckErr(os.Setenv("MONOSPACE_NO_INTERACTIVE", "1"))
		}
		mono.ProjectCreate(args[1], args[0], pType)
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
	FlagAddProjectType(createCmd)
	FlagAddNoInteractive(createCmd)
}
