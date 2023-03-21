/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/gomodules/colors"
	"github.com/software-t-rex/monospace/utils"
)

// flags for the application
var flagRootDisableColorOutput bool
var flagRemoveRmDir bool
var flagCreatePType string

var bold = utils.Bold
var underline = utils.Underline
var italic = utils.Italic

// command that require the config must call this method before continuing execution
func CheckConfigFound(exitOnError bool) bool {
	if !app.ConfigIsLoaded() {
		if exitOnError {
			utils.CheckErr(errors.New("not inside a monospace"))
		}
		return false
	}
	return true
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Version: app.Version,
	Use:     "monospace",
	Short:   "monospace is not monorepo",
	Long: utils.BrightBlue(`
   _    __     ___   _  _    ___   ____ _ _     __ _   ___  ___
`) + utils.Green(`  | '_ ' _ \  / _ \ | '_ \  / _ \ / __|| '_ \  / _' | / __|/ _ \
`) + utils.Yellow(`  | | | | | || (_) || | | || (_) |\__ \| |_) || (_| || (__|  __/
`) + utils.Red(`  |_| |_| |_| \___/ |_| |_| \___/ |___/| .__/  \__,_| \___|\___|
                                       | |
                                       |_| v`+app.Version+`
`) + `
Monospace try to bring you best of monorepo and polyrepo paradigms
You'll enjoy work in a monorepo fashion while keeping advantages of polyrepo.

If not already done start by initializing a new monospace with:
monospace init

Want to discover more about monospace? Try the help command:
monospace help [command]

Or visit https://github.com/software-t-rex/monospace for more information.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	initConfig()
	RootCmd.PersistentFlags().BoolVarP(&flagRootDisableColorOutput, "no-color", "C", false, "Disable color output mode (you can also use env var NO_COLOR)")
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.monospace)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	colors.Toggle(!flagRootDisableColorOutput)
	app.ConfigInit(utils.MonospaceGetConfigPath())
}
