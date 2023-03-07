package utils

import (
	"github.com/spf13/cobra"
)

func CompleteAddFlagProjectFilter(cmd *cobra.Command) {
	cmd.Flags().StringSliceP("project-filter", "p", []string{}, "Filter projects by name")
}
func CompleteProjectFilter(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	suggestions := append(ProjectsGetAllNameOnly(), ProjectsGetAliasesNameOnly()...)
	return suggestions, cobra.ShellCompDirectiveDefault
}
