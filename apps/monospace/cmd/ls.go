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
	"strings"

	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/software-t-rex/monospace/mono"

	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Aliases: []string{"list"},
	Use:     "ls",
	Short:   "list known workspaces in this monospace",
	Long:    `It will list workspaces in this monospace.`,
	Example: `  monospace ls -l
  monospace ls path/to/a/monospace`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			utils.CheckErr(os.Chdir(args[0]))
			mono.SpaceGetRootNoCache()
			initConfig() // force reload of monospace config
		}
		CheckConfigFound(true)
		projects := mono.ProjectsGetAll()
		if len(projects) < 1 {
			fmt.Println("No projects found start by adding one to your monospace.")
		} else {
			isLong, _ := cmd.Flags().GetBool("long")
			out := utils.SliceMap(projects, func(p mono.Project) string {
				if isLong {
					if p.RepoUrl == "" {
						return fmt.Sprintf("%s (%s)", p.StyledString(), p.Kind.String())
					}
					return fmt.Sprintf("%s (%s)", p.StyledString(), p.RepoUrl)
				} else {
					return p.StyledString()
				}
			})
			fmt.Println(strings.Join(out, "\n"))
		}
	},
}

func init() {
	RootCmd.AddCommand(lsCmd)
	lsCmd.Flags().BoolP("long", "l", false, "add information about projects repositories")
}
