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
	"sort"
	"strings"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/software-t-rex/monospace/mono"
	"github.com/software-t-rex/monospace/tasks"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage the monospace task cache",
	Long: `Manage the task cache used to skip unchanged tasks during pipeline execution.

Cache entries are stored in .monospace/.cache/ inside the monospace root.
A task uses the cache when its pipeline entry sets cache: "skip" or cache: "restore".`,
}

var cacheStatusCmd = &cobra.Command{
	Use:   "status [task...]",
	Short: "Show cache status for tasks",
	Long: `Display currently cached tasks with their hash and timestamp.

You can optionally specify task names (in project#task form) to filter the output.
Use --project-filter to restrict results to specific projects.`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		monospaceRoot := mono.SpaceGetRoot()

		// build filter list: merge positional args and project filter flag
		config := utils.CheckErrOrReturn(app.ConfigGet())
		filteredProjectNames := FlagGetFilteredProjectsNames(cmd, config)
		filters := append([]string{}, args...)

		// add project-name filters when --project-filter was set
		if len(filteredProjectNames) < len(config.Projects)+1 { // +1 for "*"
			filters = append(filters, filteredProjectNames...)
		}

		entries, err := tasks.GetCacheStatus(monospaceRoot, filters)
		if err != nil {
			utils.CheckErr(err)
		}

		if len(entries) == 0 {
			fmt.Println(theme.Info("No cache entries found."))
			return
		}

		// sort by project then task
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Project == entries[j].Project {
				return entries[i].Task < entries[j].Task
			}
			return entries[i].Project < entries[j].Project
		})

		// header
		fmt.Printf("%-40s  %-10s  %s\n", theme.Bold("task"), theme.Bold("hash"), theme.Bold("cached at"))
		fmt.Println(strings.Repeat("─", 72))
		for _, e := range entries {
			shortHash := e.Hash
			if len(shortHash) > 10 {
				shortHash = shortHash[:10]
			}
			taskKey := e.Project + "#" + e.Task
			fmt.Printf("%-40s  %-10s  %s\n", taskKey, shortHash, e.CachedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("\n%d cache %s\n", len(entries), utils.If(len(entries) == 1, "entry", "entries"))
	},
}

var cacheClearCmd = &cobra.Command{
	Use:     "clear [task...]",
	Aliases: []string{"rm"},
	Short:   "Clear task cache",
	Long: `Remove cache entries for the given tasks.

Without arguments, clears the entire cache (will prompt for confirmation unless
--force is set). With arguments (in project#task form), only those tasks are cleared.`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		monospaceRoot := mono.SpaceGetRoot()
		force, _ := cmd.Flags().GetBool("force")

		if len(args) == 0 {
			// clear entire cache
			if !force {
				if !ui.GetTerminal().IsTerminal() {
					utils.Exit("Clearing all cache requires --force when not in an interactive terminal")
				}
				confirmed, err := ui.NewConfirm("Clear the entire task cache?", false).Run()
				if err != nil || !confirmed {
					if errors.Is(err, ui.ErrSelectEscaped) || !confirmed {
						fmt.Println(theme.Info("Cache clear canceled"))
						return
					}
					utils.CheckErr(err)
				}
			}
			utils.CheckErr(tasks.ClearCache(monospaceRoot, "", ""))
			fmt.Printf("%s All cache entries cleared\n", theme.SuccessIndicator())
			return
		}

		// clear specific tasks
		config := utils.CheckErrOrReturn(app.ConfigGet())
		for _, taskArg := range args {
			stdName := tasks.StandardizedTaskName(taskArg, config)
			parts := strings.SplitN(stdName, "#", 2)
			if len(parts) != 2 {
				fmt.Printf(theme.Warning("Invalid task name %q, expected project#task\n"), taskArg)
				continue
			}
			if err := tasks.ClearCache(monospaceRoot, parts[0], parts[1]); err != nil {
				utils.PrintWarning(fmt.Sprintf("Failed to clear cache for %s: %s", stdName, err))
				continue
			}
			fmt.Printf("%s Cache cleared for %s\n", theme.SuccessIndicator(), stdName)
		}
	},
}

var cachePruneCmd = &cobra.Command{
	Use:   "prune [task...]",
	Short: "Remove cache entries beyond the configured limit",
	Long: `Remove old cache entries that exceed the configured maximum for each task.

Without arguments, prunes all cacheable tasks defined in the pipeline.
With arguments (project#task form), only those tasks are pruned.

The maximum number of entries per task is resolved as follows:
  1. cache_max_entries defined on the task itself
  2. cache_max_entries defined at the root of monospace.yml
  3. Built-in default (` + fmt.Sprintf("%d", app.DefaultCacheMaxEntries) + ` entries)`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		monospaceRoot := mono.SpaceGetRoot()
		config := utils.CheckErrOrReturn(app.ConfigGet())

		pipeline, err := tasks.GetStandardizedPipeline(config, true)
		utils.CheckErr(err)

		globalMax := config.CacheMaxEntries
		if globalMax == 0 {
			globalMax = app.DefaultCacheMaxEntries
		}

		// normalise filter args to standardized task keys
		var filterKeys []string
		for _, arg := range args {
			filterKeys = append(filterKeys, tasks.StandardizedTaskName(arg, config))
		}

		pruned := 0
		for taskKey, task := range pipeline {
			if task.TaskDef.Cache == "" || task.TaskDef.Cache == "false" {
				continue
			}
			if len(filterKeys) > 0 {
				found := false
				for _, k := range filterKeys {
					if k == taskKey {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			maxEntries := task.TaskDef.CacheMaxEntries
			if maxEntries == 0 {
				maxEntries = globalMax
			}
			parts := strings.SplitN(taskKey, "#", 2)
			if len(parts) != 2 {
				continue
			}
			if err := tasks.PruneTaskCache(monospaceRoot, parts[0], parts[1], maxEntries); err != nil {
				utils.PrintWarning(fmt.Sprintf("Failed to prune cache for %s: %s", taskKey, err))
				continue
			}
			pruned++
		}
		fmt.Printf("%s Pruned %d task cache %s\n", theme.SuccessIndicator(), pruned, utils.If(pruned == 1, "entry", "entries"))
	},
}

func init() {
	cacheCmd.AddCommand(cacheStatusCmd)
	FlagAddProjectFilter(cacheStatusCmd, false)

	cacheCmd.AddCommand(cacheClearCmd)
	cacheClearCmd.Flags().BoolP("force", "f", false, "Skip confirmation when clearing all cache")

	cacheCmd.AddCommand(cachePruneCmd)

	RootCmd.AddCommand(cacheCmd)
}
