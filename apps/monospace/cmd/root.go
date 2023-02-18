/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"monospace/monospace/cmd/colors"
	"monospace/monospace/cmd/utils"
)

var ColorOutput bool
var DisableColorOutput bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: "0.0.1",
	Use:     "monospace",
	Short:   "monospace is not monorepo",
	Long: colors.Style(colors.BrightBlue)(`
   _    __     ___   _  _    ___   ____ _ _     __ _   ___  ___
`) + colors.Style(colors.Green)(`  | '_ ' _ \  / _ \ | '_ \  / _ \ / __|| '_ \  / _' | / __|/ _ \
`) + colors.Style(colors.Yellow)(`  | | | | | || (_) || | | || (_) |\__ \| |_) || (_| || (__|  __/
`) + colors.Style(colors.Red)(`  |_| |_| |_| \___/ |_| |_| \___/ |___/| .__/  \__,_| \___|\___|
                                       | |
                                       |_|
`) + `
Monospace try to bring you best of monorepo and poly-repo paradigms
You'll enjoy work in a monorepo fashion while keeping advantages of polyrepo.

If not already done start by initializing a new monospace with:
monospace init

Want to discover more about monospace? Try the help command:
monospace help [command]

Or visit https://github.com/malko/monospace for more information.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVarP(&DisableColorOutput, "no-color", "C", false, "Disable color output mode (you can also use env var NO_COLOR)")
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
	viper.SetConfigFile(utils.MonospaceGetConfigPath())
	viper.SetDefault("js_package_manager", "^pnpm@7.27.0")
	// viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	utils.CheckErr(viper.ReadInConfig())
	colors.Toggle(!DisableColorOutput)
}
