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
	"slices"
	"sort"
	"strings"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/software-t-rex/monospace/mono"
	"github.com/software-t-rex/monospace/tasks"
	"github.com/software-t-rex/packageJson"
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List tasks defined in monospace.yml pipeline.",
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		preferFullNames, _ := cmd.Flags().GetBool("full-project-names")
		details, _ := cmd.Flags().GetBool("details")
		config := utils.CheckErrOrReturn(app.ConfigGet())
		filteredProjects := FlagGetFilteredProjectsNames(cmd, config)
		pipeline := utils.CheckErrOrReturn(tasks.GetStandardizedPipeline(config, true))

		// get project aliases
		projectAliases := config.GetProjectsAliases()

		// sort tasks by projects then by task names
		keys := make([]string, 0, len(pipeline))
		for key := range pipeline {
			keys = append(keys, key)
		}
		sort.SliceStable(keys, func(i, j int) bool {
			keyI := keys[i]
			keyJ := keys[j]
			if pipeline[keyI].Name.Project == pipeline[keyJ].Name.Project {
				return pipeline[keyI].Name.Task < pipeline[keyJ].Name.Task
			}
			return pipeline[keyI].Name.Project < pipeline[keyJ].Name.Project
		})

		// print tasks
		sb := strings.Builder{}
		for key := range keys {
			task := pipeline[keys[key]]
			name := task.Name.String()
			if !utils.SliceContains(filteredProjects, task.Name.Project) {
				continue
			}
			if !preferFullNames && task.Name.Project != "*" {
				alias := projectAliases[task.Name.Project]
				if alias != "" {
					name = fmt.Sprintf("%s#%s", alias, task.Name.Task)
				}
			}
			if task.TaskDef.Persistent {
				name = theme.Warning(name + " (persistent)")
			}
			if !details {
				sb.WriteString(fmt.Sprintf("%s\n", name))
			} else {
				sb.WriteString(fmt.Sprintf("%s:\n", theme.Underline(name)))
				if task.TaskDef.Description != "" {
					sb.WriteString(fmt.Sprintf("  %s: %s\n", theme.Italic("description"), task.TaskDef.Description))
				}
				sb.WriteString(fmt.Sprintf("  %s: %s\n", theme.Italic("command"), task.GetJobRunner(nil, config.JSPM).String()))
				if task.TaskDef.DependsOn != nil {
					sb.WriteString(fmt.Sprintf("  %s: %s\n", theme.Italic("dependends on"), strings.Join(task.TaskDef.DependsOn, ", ")))
				}
				if task.TaskDef.OutputMode != "" && task.TaskDef.OutputMode != config.PreferredOutputMode {
					sb.WriteString(fmt.Sprintf("  %s: %s\n", theme.Italic("output mode"), task.TaskDef.OutputMode))
				}
			}
		}
		fmt.Print(sb.String())
	},
}

var tasksImportCmd = &cobra.Command{
	Use:   "import [projectName#scriptName]...",
	Short: "Import scripts entries from projects package.json.",
	Long: `Import scripts entries from projects package.json file as monospace pipeline task.

You can either pass a list of scripts to import using the syntax
'projectName#scriptName', or you can let the command prompt you for scripts to
import. You can narrow choices presented to you by using the --project-filter
flag. 

If running in interactive mode you will be able to edit the task before adding it
to the pipeline.

Pipeline will be checked for cyclic dependencies before saving any changes.
`,

	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		noInteractive := FlagGetNoInteractive(cmd)
		if !ui.GetTerminal().IsTerminal() && !noInteractive { // must be in non-interactive mode when not in a terminal
			utils.Exit("This command requires an interactive terminal unless --no-interactive flag is set")
		} else if noInteractive && len(args) == 0 {
			utils.Exit("You must provide at least one script descriptor to import in non-interactive mode")
		}

		// load config
		config := utils.CheckErrOrReturn(app.ConfigGet())
		// get requested pkgjsonProjects and filter only the ones that have a package.json file
		pkgjsonProjects := FlagGetFilteredProjects(cmd, config)
		pkgjsonProjects = utils.SliceFilter(pkgjsonProjects, func(p mono.Project) bool {
			return p.HasPackageJson()
		})
		if len(pkgjsonProjects) == 0 {
			utils.PrintInfo("No matching project with package.json file found")
			return
		}
		// cache for packageJsons files indexed by project names
		packageJsons := map[string]*packageJson.PackageJSON{}
		// retrieve tasks already defined in the pipeline and filter only the ones with no command defined
		pipeline := utils.CheckErrOrReturn(tasks.GetStandardizedPipeline(config, false))

		// check pipeline is acyclic before adding new tasks
		if !pipeline.IsAcyclic(false) {
			utils.Exit("Pipeline has cyclic dependencies, please fix it before adding new tasks")
		}

		// get a map of reversed project aliases to easily find alias from project names
		var scriptsToImport []string
		// will hold new tasks to add to the pipeline
		newTasks := []tasks.Task{}
		if len(args) > 0 {
			scriptsToImport = args
		} else { // no args given we will search for importable scripts to presents to user
			// retrieve all scripts defined in the package.json files that are not already defined in the pipeline
			scriptOptions := []string{}
			for _, project := range pkgjsonProjects {
				packageJson, err := project.GetPackageJson()
				if err != nil {
					utils.PrintWarning(fmt.Sprintf("Failed to read package.json file for project %s: %s", project.Name, err))
					continue
				}
				packageJsons[project.Name] = packageJson
				if packageJson.Scripts != nil {
					for scriptName := range packageJson.Scripts {
						// skip task that already exists in the pipeline
						if _, exists := pipeline[fmt.Sprintf("%s#%s", project.Name, scriptName)]; exists {
							continue
						}
						scriptOptions = append(scriptOptions, fmt.Sprintf("%s#%s", project.Name, scriptName))
					}
				}
			}
			if len(scriptOptions) == 0 {
				utils.PrintInfo("No undeclared scripts found in package.json files")
				return
			}
			slices.Sort(scriptOptions)
			scriptSelection, errSelScripts := ui.NewMultiSelectStrings("Select scripts you want to import in your pipeline?", scriptOptions).
				AllowEmptySelection().
				MaxVisibleOptions(10).
				WithCleanup(true).
				WithSelectAll("a").
				Escapable(true).
				Run()
			if errSelScripts != nil {
				if errors.Is(errSelScripts, ui.ErrSelectEscaped) {
					fmt.Println(theme.Info("Task import canceled"))
					return
				}
				utils.Exit(fmt.Sprintf("Error while selecting tasks: %s", errSelScripts))
			}
			scriptsToImport = scriptSelection
		}

		if len(scriptsToImport) == 0 {
			fmt.Println(theme.Info("No tasks to import"))
			return
		}

		// now proceded with importing the selected scripts
		for _, scriptName := range scriptsToImport {
			if !strings.Contains(scriptName, "#") {
				utils.PrintWarning(fmt.Sprintf("Invalid script name '%s', it must be in the form 'projectName#scriptName'", scriptName))
				continue
			}
			taskName := tasks.ParseTaskName(scriptName, config)
			if _, exists := pipeline[taskName.String()]; exists {
				utils.PrintWarning(fmt.Sprintf("Task %s already defined in the pipeline", taskName))
				continue
			}
			// load packagejson file if not already loaded
			packageJson, packageJsonCached := packageJsons[taskName.Project]
			if !packageJsonCached {
				// try to load the package.json file
				project, errProject := mono.ProjectGetByName(taskName.Project)
				if errProject != nil {
					utils.PrintWarning(fmt.Sprintf("Failed to load project %s: %s", taskName.Project, errProject))
				}
				pkgJson, errPkg := project.GetPackageJson()
				if errPkg != nil {
					utils.PrintWarning(fmt.Sprintf("Failed to read package.json file for project %s: %s", project.Name, errPkg))
					continue
				}
				packageJson = pkgJson
				packageJsons[taskName.Project] = packageJson
			}
			// check script exists in package.json
			if _, scriptExists := packageJson.Scripts[taskName.Task]; !scriptExists {
				utils.PrintWarning(fmt.Sprintf("can't find matching package.json script for %s", scriptName))
			}
			// now we can add the task to newTasksMap with default config
			newTasks = append(newTasks, tasks.Task{
				Name: taskName,
				TaskDef: app.MonospaceConfigTask{
					Description: fmt.Sprintf("Run %s script from %s/package.json", taskName.Task, taskName.Project),
				},
			})
		}

		if len(newTasks) == 0 {
			utils.PrintInfo("No task added to the pipeline")
			return
		}

		// add new tasks to the pipeline
		for _, task := range newTasks {
			for {
				if !noInteractive { // allow editing the task in interactive mode
					editedTask, errEdit := tasks.NewTaskEditor(config, task).Run()
					if errEdit != nil {
						utils.Exit(fmt.Sprintf("Error while editing task %s: %s", task.Name, errEdit))
					}
					task = editedTask
				}
				// add the task to the pipeline
				pipeline[task.Name.String()] = task
				if !pipeline.IsAcyclic(false) {
					utils.PrintWarning("Pipeline has cyclic dependencies, please fix it before saving the changes")
					delete(pipeline, task.Name.String())
					continue
				}
				break
			}
			config.Pipeline = pipeline.ToConfig(config)
			fmt.Printf("%s Task %s added to the pipeline\n", theme.SuccessIndicator(), task.Name)
		}
		utils.CheckErr(app.ConfigSave())
		fmt.Print(theme.Success("monospace.yml updated\n"))
	},
}

var tasksRmCmd = &cobra.Command{
	Use:     "remove [taskName]...",
	Aliases: []string{"rm"},
	Short:   "Remove tasks defined in monospace.yml pipeline.",
	Long: `Remove tasks defined in monospace.yml pipeline.
You can pass a list of taskNames to remove using the exact same syntax used in 
the monospace.yml file. If no task is provided you will be prompted to select a
task to remove from the pipeline.

It will also remove the task from other tasks dependencies. And perform a check
on cyclic dependencies before saving the changes.`,
	ValidArgsFunction: completeTaskNameArgs,
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		config := utils.CheckErrOrReturn(app.ConfigGet())
		theme := ui.GetTheme()
		pipeline := utils.CheckErrOrReturn(tasks.GetStandardizedPipeline(config, true))
		filteredProjects := FlagGetFilteredProjectsNames(cmd, config)
		filteredPipeline := utils.MapFilter(pipeline, func(task tasks.Task) bool {
			return utils.SliceContains(filteredProjects, task.Name.Project)
		})
		if len(args) == 0 {
			if !ui.GetTerminal().IsTerminal() {
				utils.Exit("This command requires an interactive terminal, unless you provide tasks to remove as arguments")
			}
			selected, errSel := tasks.TaskSingleSelector(filteredPipeline, "Select task to remove", "Exit")
			if errSel != nil {
				if errors.Is(errSel, tasks.ErrNoAvailableOption) {
					fmt.Println(theme.Info("No task to remove start by adding some to the pipeline\n"))
					return
				} else if errors.Is(errSel, ui.ErrSelectEscaped) {
					fmt.Println(theme.Info("No task selected\n"))
					return
				}
				utils.Exit(fmt.Sprintf("Error while selecting task: %s", errSel))
				return // unreachable
			} else if selected == "" {
				return
			}
			args = []string{selected}
		}
		// case where we have tasks given for removal
		if len(args) > 0 {
			// each args should be a task to remove
			needSave := false
			for _, taskName := range args {
				stdName := tasks.StandardizedTaskName(taskName, config)
				if _, exists := pipeline[stdName]; !exists {
					fmt.Printf(theme.Warning("Task %s not found in pipeline\n"), stdName)
					continue
				}
				pipeline = pipeline.RemoveTask(taskName, config)
				fmt.Printf("%s %s removed from pipeline\n", theme.SuccessIndicator(), taskName)
				needSave = true
			}
			// replace the pipeline in the config
			if needSave {
				config.Pipeline = pipeline.ToConfig(config)
				if !pipeline.IsAcyclic(false) {
					utils.Exit("Pipeline has cyclic dependencies, changes not saved")
				}
				utils.CheckErr(app.ConfigSave())
				fmt.Print(theme.Success("monospace.yml updated\n"))
			}
			return
		}
	},
}

var tasksEditCmd = &cobra.Command{
	Use:               "edit [taskName]...",
	Short:             "Edit tasks defined in monospace.yml pipeline.",
	ValidArgsFunction: completeTaskNameArgs,
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		config := utils.CheckErrOrReturn(app.ConfigGet())
		if !ui.GetTerminal().IsTerminal() {
			utils.Exit("This command requires an interactive terminal")
		}
		pipeline := utils.CheckErrOrReturn(tasks.GetStandardizedPipeline(config, true))

		if len(args) == 0 {
			selected, errSel := tasks.TaskSingleSelector(pipeline, "Select task to edit", "Cancel")
			if errSel != nil {
				if errors.Is(errSel, tasks.ErrNoAvailableOption) {
					fmt.Println(theme.Info("No task to edit start by adding some to the pipeline\n"))
					return
				} else if errors.Is(errSel, ui.ErrSelectEscaped) {
					fmt.Println(theme.Info("No task selected"))
					return // user selected cancel or pressed escape key
				}
				utils.Exit(fmt.Sprintf("Error while selecting task: %s", errSel))
			} else if selected == "" {
				return
			}
			args = []string{selected}
		}
		needSave := false
		for _, taskStringName := range args {
			taskName := tasks.StandardizedTaskName(taskStringName, config)
			task, taskExists := pipeline[taskName]
			if !taskExists {
				fmt.Printf(theme.Warning("Task %s not found in pipeline\n"), taskStringName)
				continue
			}
			for {
				editedTask, errEdit := tasks.NewTaskEditor(config, task).Run()
				if errEdit != nil {
					if errors.Is(errEdit, tasks.ErrEditorCanceled) {
						fmt.Println(theme.Info("Task editing canceled"))
						break
					}
					utils.Exit(fmt.Sprintf("Error while editing task %s: %s", task.Name, errEdit))
				}
				task = editedTask

				// override task in pipeline
				pipeline[taskName] = task
				if !pipeline.IsAcyclic(false) {
					utils.PrintWarning("Pipeline has cyclic dependencies, please fix it before saving the changes")
					continue
				}
				needSave = true
				config.Pipeline = pipeline.ToConfig(config)
				fmt.Printf("%s Task %s updated in pipeline\n", theme.SuccessIndicator(), task.Name)
				break
			}
		}
		if !needSave {
			fmt.Println(theme.Info("No task updated"))
			return
		}
		utils.CheckErr(app.ConfigSave())
		fmt.Print(theme.Success("monospace.yml updated\n"))
	},
}

func init() {
	tasksCmd.AddCommand(tasksImportCmd)
	FlagAddProjectFilter(tasksImportCmd, true)
	FlagAddNoInteractive(tasksImportCmd)

	tasksCmd.AddCommand(tasksRmCmd)
	FlagAddProjectFilter(tasksRmCmd, true)

	tasksCmd.AddCommand(tasksEditCmd)

	RootCmd.AddCommand(tasksCmd)
	FlagAddProjectFilter(tasksCmd, true)
	tasksCmd.Flags().BoolP("full-project-names", "f", false, "Prefer full project names rather than aliases")
	tasksCmd.Flags().BoolP("details", "d", false, "Print additional task details")
}
