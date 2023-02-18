/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/
package cmd

import (
	"fmt"
	"monospace/monospace/cmd/utils"
	"strings"

	"github.com/spf13/cobra"
)

var longFormat bool

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list knwon workspaces in this monospace",
	Long: `It will list workspaces in this monospace.

usage:
monospace ls [options]
monospace ls [options] path/to/a/monospace
`,
	Run: func(cmd *cobra.Command, args []string) {
		projects := utils.ProjectsGetAll()
		if len(projects) < 1 {
			fmt.Println("No projects found start by adding one to your monospace.")
		} else {
			out := utils.Map(projects, func(p utils.Project) string {
				if longFormat {
					if p.RepoUrl == "" {
						return fmt.Sprintf("%s\t(%s)", p.StyledString(), p.Kind.String())
					}
					return fmt.Sprintf("%s\t(%s)", p.StyledString(), p.RepoUrl)
				} else {
					return p.StyledString()
				}
			})
			fmt.Println(strings.Join(out, "\n"))
		}
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	lsCmd.Flags().BoolVarP(&longFormat, "longFormat", "l", false, "add information about projects repositories")

}
