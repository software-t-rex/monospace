package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra"
)

// constants for flag enumValue, first item in the list is default value
const projectTypes = ",go,js"
const outputModes = "grouped,interleaved,none"

func exitAndHelp(cmd *cobra.Command, err error) {
	utils.PrintError(err)
	cmd.Help()
	os.Exit(1)
}

func AddFlagProjectFilter(cmd *cobra.Command) {
	cmd.Flags().StringSliceP("project-filter", "p", []string{}, "Filter projects by name")
	utils.CheckErr(cmd.RegisterFlagCompletionFunc("project-filter", completeProjectFilter))
}
func completeProjectFilter(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	suggestions := append(utils.ProjectsGetAllNameOnly(), utils.ProjectsGetAliasesNameOnly()...)
	return suggestions, cobra.ShellCompDirectiveDefault
}

// you should call ValidateFlagOutputMode in the Run of the associated command
func AddFlagOutputMode(cmd *cobra.Command) {
	cmd.Flags().StringP("output-mode", "O", "grouped", "output mode for multiple commands, grouped, interleaved, none")
	utils.CheckErr(cmd.RegisterFlagCompletionFunc("output-mode", CompleteOutputMode))
}
func CompleteOutputMode(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"none", "grouped", "interleaved"}, cobra.ShellCompDirectiveDefault
}

func ValidateFlagOutputMode(cmd *cobra.Command) string {
	outputMode := utils.CheckErrOrReturn(cmd.Flags().GetString("output-mode"))
	modes := strings.Split(outputModes, ",")
	if outputMode == "" {
		return modes[0]
	} else if utils.SliceContains(modes, outputMode) {
		return outputMode
	}
	exitAndHelp(cmd, fmt.Errorf("invalid output-mode %s", outputMode))
	return "" // will never get called
}

func AddFlagNoInteractive(cmd *cobra.Command) {
	cmd.Flags().BoolP("no-interactive", "y", false, "Prevent any interactive prompts")
}

// should use ValidateFlagProjectType in Run
func AddFlagProjectType(cmd *cobra.Command) {
	cmd.Flags().StringP("type", "t", "", "type of project to create for now only 'go' and 'js' are supported")
	utils.CheckErr(cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return strings.Split(projectTypes, ",")[1:], cobra.ShellCompDirectiveNoFileComp
	}))
}
func ValidateFlagProjectType(cmd *cobra.Command) string {
	pType := utils.CheckErrOrReturn(cmd.Flags().GetString("type"))
	if !utils.SliceContains(strings.Split(projectTypes, ","), pType) {
		exitAndHelp(cmd, fmt.Errorf("invalid project type '%s'", pType))
	}
	return pType
}
