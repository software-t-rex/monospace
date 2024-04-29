/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/
package cmd

import (
	"fmt"

	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/software-t-rex/monospace/mono"
	"github.com/spf13/cobra"
)

// stateCmd represents the state command
var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Allow to pin a full monospace state and restore it later",
	Long: `The state command allow to "pin" the state of all projects in the monospace and
to "restore" pinned states at a later time. This can be useful when you want to 
reproduce a particular state of your monospace and share it with co-workers.

Basically this store the current revision of all projects in the monospace.
This is not a full backup of the monospace as it does not store the content of 
the projects but only the revision of each project.

When using the "state restore" command, each projects will be checked out to the
given revision. This will leave your repositories in a detached head state.

Local projects will be ignored and left as is.

` + ui.ApplyStyle(">> This is highly experimental, any feedback will be greatly appreciated! <<", ui.YellowBright.Background(), ui.Black.Foreground()),
	Example: `  # pin the current state of the monospace
  monospace state pin myState
  # restore a previously pinned state
  monospace state restore myState
  # list all pinned states
  monospace state list
  # remove a pinned state
  monospace state unpin myState`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{"pin", "unpin", "restore", "list"}, cobra.ShellCompDirectiveDefault
		}
		if len(args) == 1 {
			if args[0] == "unpin" || args[0] == "restore" {
				return mono.StateList(), cobra.ShellCompDirectiveDefault
			} else {
				return nil, cobra.ShellCompDirectiveDefault
			}
		}
		return nil, cobra.ShellCompDirectiveDefault
	},
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		states := utils.CheckErrOrReturn(mono.StateLoad())
		if len(args) == 0 {
			args = append(args, "list")
		}
		if args[0] == "list" {
			if states.Len() == 0 {
				fmt.Println("no pinned state")
				return
			}
			fmt.Println("pinned states:")
			for name := range states.States {
				fmt.Println(" -", name)
			}
			return
		}
		if args[0] == "pin" || args[0] == "unpin" || args[0] == "restore" {
			if len(args) < 2 {
				utils.Exit("missing state name")
			}
		}
		switch args[0] {
		case "pin":
			states.Add(args[1])
			utils.CheckErr(states.Save())
			return
		case "unpin":
			states.Remove(args[1])
			utils.CheckErr(states.Save())
			return
		case "restore":
			states.Restore(args[1])
			return
		default:
			fmt.Println("unknown command")
		}
	},
}

func init() {
	RootCmd.AddCommand(stateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
