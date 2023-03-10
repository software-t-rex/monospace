/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/
package cmd

import (
	"fmt"

	"github.com/software-t-rex/monospace/app"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "return the monospace cli version",
	Long:  `return the monospace cli version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("monospace version %s\n", app.Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
