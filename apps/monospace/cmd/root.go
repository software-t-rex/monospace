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
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/software-t-rex/monospace/mono"
)

// flags for the application
var flagRootDisableColorOutput bool
var theme *ui.Theme

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
	Long: ui.ApplyStyle(`
   _    __     ___   _  _    ___   ____ _ _     __ _   ___  ___
`, ui.Blue.Foreground()) +
		ui.ApplyStyle(`  | '_ ' _ \  / _ \ | '_ \  / _ \ / __|| '_ \  / _' | / __|/ _ \
`, ui.Green.Foreground()) +
		ui.ApplyStyle(`  | | | | | || (_) || | | || (_) |\__ \| |_) || (_| || (__|  __/
`, ui.Yellow.Foreground()) +
		ui.ApplyStyle(`  |_| |_| |_| \___/ |_| |_| \___/ |___/| .__/  \__,_| \___|\___|
                                       | |
                                       |_| v`+app.Version+`
`, ui.Red.Foreground()) + `
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
	// this is a trick to disable color output when in completion mode
	completionMode := false
	for _, arg := range os.Args {
		if arg == "__complete" || arg == "completion" {
			completionMode = true
			break
		}
	}
	// cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().BoolVarP(&flagRootDisableColorOutput, "no-color", "C", false, "Disable color output mode (you can also use env var NO_COLOR)")
	ui.ToggleEnhanced(!completionMode && !flagRootDisableColorOutput)
	theme = ui.SetTheme(ui.ThemeMonoSpace)
	app.ConfigInit(mono.SpaceGetConfigPath())
}

// utility function to force reload of monospace config from a given directory
// exit on error
func forceReloadFromDir(dir string) {
	dir = utils.CheckErrOrReturn(filepath.Abs(dir))
	utils.CheckErr(os.Chdir(dir))
	// force reload of monospace config
	rootDir := mono.SpaceGetRootNoCache()
	if rootDir == "" {
		utils.Exit(fmt.Sprintf("%s is not part of a monospace", dir))
	}
	app.ConfigInitNoCheck(mono.SpaceGetConfigPath()) // force reload of monospace config
}
