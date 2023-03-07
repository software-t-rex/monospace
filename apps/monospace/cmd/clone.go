/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package cmd

import (
	"github.com/software-t-rex/monospace/utils"

	"github.com/spf13/cobra"
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone destDirectory repoUrl",
	Short: "Clone an entire monospace",
	Long: `Clone is like git clone but for a whole monospace repo.
It will clone the monospace git repo and then checkout all 'external' projects
into it.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		destDirectory := args[0]
		repoUrl := args[1]
		utils.MonospaceClone(destDirectory, repoUrl)
	},
}

func init() {
	RootCmd.AddCommand(cloneCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cloneCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cloneCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
