package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/mono"
	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra"
)

// constants for flag enumValue, first item in the list is default value
const projectTypes = ",go,js"
const outputModes = "grouped,interleaved,status-only,errors-only,none"

func exitAndHelp(cmd *cobra.Command, err error) {
	utils.PrintError(err)
	cmd.Help()
	os.Exit(1)
}

func FlagGetBool(cmd *cobra.Command, name string) bool {
	return utils.CheckErrOrReturn(cmd.Flags().GetBool(name))
}
func FlagGetString(cmd *cobra.Command, name string) string {
	return utils.CheckErrOrReturn(cmd.Flags().GetString(name))
}
func FlagGetStringSlice(cmd *cobra.Command, name string) []string {
	return utils.CheckErrOrReturn(cmd.Flags().GetStringSlice(name))
}

func FlagAddProjectFilter(cmd *cobra.Command, includeRootAsDefault bool) {
	if !includeRootAsDefault {
		cmd.Flags().BoolP("include-root", "r", false, "Include 'root' monospace directory in the list of projects\n- Without any filter, 'root' is only appended to projects list\n- Used with --project-filter, 'root' is appended to filters list")
	}
	cmd.Flags().StringSliceP("project-filter", "p", []string{}, "Filter projects by name\nThis is like 'whitelisting' project in the list\nYou can use 'root' for monospace root directory")
	cmd.Flags().StringSliceP("project-filter-out", "P", []string{}, "Filter out by name\nExclude projects from the list (blacklisting)")
	utils.CheckErr(cmd.RegisterFlagCompletionFunc("project-filter", completeProjectFilter))
	utils.CheckErr(cmd.RegisterFlagCompletionFunc("project-filter-out", completeProjectFilter))
}

//	func FlagAddPersistentProjectFilter(cmd *cobra.Command) {
//		cmd.PersistentFlags().StringSliceP("project-filter", "p", []string{}, "Filter projects by name\nYou can use 'root' for monospace root directory\nUse '\\!' prefix to exclude a project")
//		utils.CheckErr(cmd.RegisterFlagCompletionFunc("project-filter", completeProjectFilter))
//	}
func completeProjectFilter(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	suggestions := append(append(mono.ProjectsGetAllNameOnly(), mono.ProjectsGetAliasesNameOnly()...), "root")
	return suggestions, cobra.ShellCompDirectiveDefault
}

func GetFilteredProjects(projects []mono.Project, filters []string, includeRoot bool) []mono.Project {
	config := utils.CheckErrOrReturn(app.ConfigGet())
	filterLen := len(filters)
	if !includeRoot && utils.SliceContains(filters, "root") {
		includeRoot = true
	}
	// prepend with root monospace
	if includeRoot {
		projects = append([]mono.Project{mono.RootProject}, projects...)
	}
	if filterLen < 1 { // no filter return all projects
		return projects
	}
	replaceAlias := func(name string) string { return name }
	if len(config.Aliases) > 0 {
		replaceAlias = func(name string) string {
			if alias := config.Aliases[name]; alias != "" {
				return alias
			}
			return name
		}
	}

	// split filters between white and black list
	var whiteList []string
	var blackList []string
	for _, f := range filters {
		if strings.HasPrefix(f, "!") {
			blackList = append(blackList, replaceAlias(strings.TrimPrefix(f, "!")))
		} else {
			whiteList = append(whiteList, replaceAlias(f))
		}
	}

	// apply black list
	if len(blackList) > 0 {
		projects = utils.SliceFilter(projects, func(p mono.Project) bool {
			return !utils.SliceContains(blackList, p.Name)
		})
	}
	// apply white list
	if len(whiteList) > 0 {
		projects = utils.SliceFilter(projects, func(p mono.Project) bool {
			return utils.SliceContains(whiteList, p.Name)
		})
	}
	return projects
}
func FlagGetFilteredProjects(cmd *cobra.Command) []mono.Project {
	projects := mono.ProjectsGetAll()
	filters := utils.CheckErrOrReturn(cmd.Flags().GetStringSlice("project-filter"))
	filtersOut := utils.CheckErrOrReturn(cmd.Flags().GetStringSlice("project-filter-out"))
	includeRoot := false
	if cmd.Flags().Lookup("include-root") == nil {
		includeRoot = true
	} else if FlagGetBool(cmd, "include-root") {
		if len(filters) > 0 {
			if !utils.SliceContains(filters, "root") {
				filters = append(filters, "root")
			}
		} else {
			includeRoot = true
		}
	}
	for _, f := range filtersOut {
		filters = append(filters, "!"+f)
	}
	return GetFilteredProjects(projects, filters, includeRoot)
}

// you should call GEtFlagOutputMode in the Run of the associated command
func FlagAddOutputMode(cmd *cobra.Command) {
	cmd.Flags().StringP("output-mode", "O", "", "output mode for multiple commands:\n- "+strings.Replace(outputModes, ",", "\n- ", -1)+"\n(default to monospace.yml settings or grouped if not set)")
	utils.CheckErr(cmd.RegisterFlagCompletionFunc("output-mode", completeOutputMode))
}

// you should call GetFlagOutputMode in the Run of the associated command
//
//	func FlagAddPersistentOutputMode(cmd *cobra.Command) {
//		cmd.PersistentFlags().StringP("output-mode", "O", "grouped", "output mode for multiple commands:\n- "+strings.Replace(outputModes, ",", "\n- ", -1)+"\n")
//		utils.CheckErr(cmd.RegisterFlagCompletionFunc("output-mode", completeOutputMode))
//	}
func completeOutputMode(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return strings.Split(outputModes, ","), cobra.ShellCompDirectiveDefault
}
func FlagGetOutputMode(cmd *cobra.Command, dflt string) string {
	outputMode := utils.CheckErrOrReturn(cmd.Flags().GetString("output-mode"))
	modes := strings.Split(outputModes, ",")
	if utils.SliceContains(modes, outputMode) {
		return outputMode
	}
	if outputMode == "" {
		if dflt != "" {
			return dflt
		} else {
			return modes[0]
		}
	}
	exitAndHelp(cmd, fmt.Errorf("invalid output-mode %s", outputMode))
	return "" // will never get called
}

func FlagAddNoInteractive(cmd *cobra.Command) {
	cmd.Flags().BoolP("no-interactive", "y", false, "Prevent any interactive prompts by choosing default values (not always yes)")
}
func FlagGetNoInteractive(cmd *cobra.Command) bool {
	return utils.CheckErrOrReturn(cmd.Flags().GetBool("no-interactive"))
}

// should use GetFlagProjectType in Run
func FlagAddProjectType(cmd *cobra.Command) {
	cmd.Flags().StringP("type", "t", "", "type of project to create for now only 'go' and 'js' are supported")
	utils.CheckErr(cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return strings.Split(projectTypes, ",")[1:], cobra.ShellCompDirectiveNoFileComp
	}))
}
func GetFlagProjectType(cmd *cobra.Command) string {
	pType := utils.CheckErrOrReturn(cmd.Flags().GetString("type"))
	if !utils.SliceContains(strings.Split(projectTypes, ","), pType) {
		exitAndHelp(cmd, fmt.Errorf("invalid project type '%s'", pType))
	}
	return pType
}
