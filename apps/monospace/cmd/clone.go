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
	"regexp"
	"strings"

	"github.com/software-t-rex/monospace/mono"
	"github.com/software-t-rex/monospace/utils"

	"github.com/spf13/cobra"
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone repoUrl [destDirectory]",
	Short: "Clone an entire monospace",
	Long: `Clone is like git clone but for a whole monospace repo.
It will clone the monospace git repo and then checkout all 'external' projects
into it.`,
	Args: cobra.MatchAll(cobra.MaximumNArgs(2), cobra.MinimumNArgs(1)),
	Run: func(cmd *cobra.Command, args []string) {
		var destDirectory string
		var repoUrl string
		if len(args) == 2 {
			destDirectory = args[1]
			repoUrl = args[0]
			if strings.HasSuffix(destDirectory, ".git") && !strings.HasSuffix(repoUrl, ".git") {
				utils.PrintWarning("Seems like you inverted the arguments,")
				fmt.Println("Should we use " + destDirectory + " as your repository url")
				if utils.Confirm("and "+repoUrl+" as the destination directory ?", true) {
					destDirectory = args[0]
					repoUrl = args[1]
				}
			}
		} else if len(args) == 1 && strings.HasSuffix(args[0], ".git") {
			// choose name from the git url
			repoUrl = args[0]
			destDirectory = regexp.MustCompile(`([^/]+)\.git$`).FindStringSubmatch(repoUrl)[1]
			if destDirectory == "" {
				fmt.Println(utils.ErrorStyle("can't detect destination directory"))
				cmd.Help()
				os.Exit(1)
			}
		}
		mono.SpaceClone(destDirectory, repoUrl)
	},
}

func init() {
	RootCmd.AddCommand(cloneCmd)
}
