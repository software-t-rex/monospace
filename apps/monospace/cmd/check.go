/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	jspm "github.com/software-t-rex/js-packagemanager"
	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/git"
	"github.com/software-t-rex/monospace/mono"
	"github.com/software-t-rex/monospace/tasks"
	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Performs various checks on the monospace config file and projects",
	Long: `Performs various checks on the monospace config file and projects.

## Check projects
Check projects remote origin match the monospace config file

While using monospace you will import some new repositories to the monospace,
create local projects which soon will become either internal or external projects.
This command help you maintain projects remote origins and projects settings
in the monospace.yml config file consistent.

Here's the reported anomalies and the action taken when --fix flag is used:
- for local projects:
	- check it's still a git repository: fixed by setting project as internal
	- check project remote origin is still not set: fixed by updating config
		with repo remote origin or setting project internal
- for external projects:
	- check project files exists: fixed by cloning the project
	- check project dir is a git repository: fixed by setting project internal
	- check project repo origin match the one in config: fixed by updating the
		config with repo remote origin
- for internal projects:
	- check project dir exists: fixed by removing project from config
	- check project is not a git repository: fixed by updating config with
		repo remote origin (error if remote origin is not set)

More choices may be available when --interactive flag is used

## Check pipeline (skipped if --project-filter is used)
- Check tasks are associated with existing projects.
- Check tasks depends on existing non persistent tasks.
- Check for circular task dependencies
There's no fix available on pipeline errors

## Check githooks path (warning only, skipped if --project-filter is used)
- if a .monospace/githooks dir exists check git core.hooksPath is set to it
There's no fix available on githooks path errors only a warning message, it
won't change the exit status of the command.

## Check js_package_manager (skipped if --project-filter is used)
- check package manager version match the one installed: fixed by updating config
won't change the exit status of the command.
`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		utils.CheckErr(mono.SpaceChdir())
		monoRoot := mono.SpaceGetRoot()
		monoIgnore := filepath.Join(monoRoot, ".gitignore")
		hasFilter := cmd.Flags().Changed("project-filter")
		filteredProjects := FlagGetFilteredProjects(cmd)
		interactive := FlagGetBool(cmd, "interactive")
		successIndicator := utils.Green("✔")
		failureIndicator := utils.Red("✘")
		unknwonIndicator := utils.Yellow("�")
		fix := FlagGetBool(cmd, "fix")
		exitStatus := 0

		printCheckHeader := func(p mono.Project, warningMsg ...string) {
			if len(warningMsg) < 1 {
				fmt.Printf(utils.Bold("%s %s\n"), successIndicator, p.StyledString())
				return
			}
			exitStatus = 1
			fmt.Printf(utils.Bold("%s %s\n"), failureIndicator, p.StyledString())
			utils.PrintWarning(warningMsg...)
		}

		setInternal := func(p mono.Project) {
			fmt.Println("setting project as internal...")
			utils.CheckErr(app.ConfigAddOrUpdateProject(p.Name, "internal", true))
			utils.CheckErr(utils.FileRemoveLine(monoIgnore, p.Name))
		}

		// check package manager version
		checkJSPMVersion := func() error {
			fmt.Println(utils.Bold(utils.Underline("checking js_package_manager config:")))
			var pm *jspm.PackageManager
			var err error
			var version string
			var updateVersion bool
			config := utils.CheckErrOrReturn(app.ConfigGet())
			pm, err = jspm.GetPackageManagerFromString(config.JSPM)
			if err != nil {
				fmt.Println(failureIndicator + " " + utils.Warning(err.Error()))
				return err
			}
			version, err = pm.GetVersion()
			if err != nil {
				fmt.Println(failureIndicator + " " + utils.Warning(err.Error()))
				return err
			}
			if version == "" {
				fmt.Println(unknwonIndicator + " " + utils.Warning("unable to get installed package manager version"))
				return nil
			}
			if strings.HasSuffix(config.JSPM, version) {
				fmt.Println(successIndicator + " " + utils.Success(pm.Slug+" at version "+version))
				return nil
			}
			versionConfig := "^" + pm.Slug + "@" + version
			if fix {
				updateVersion = true
			} else if interactive && utils.Confirm(fmt.Sprintf("Do you want to update js_package_manager config from %s to %s ?", config.JSPM, versionConfig), true) {
				updateVersion = true
			}
			if !updateVersion {
				fmt.Printf(failureIndicator+" "+utils.Warning("js_package_manager config (%s) is not up to date with installed version (%s)\n"), config.JSPM, version)
			} else {
				config.JSPM = versionConfig
				utils.CheckErr(config.Save())
				fmt.Printf("%s updated js_package_manager config to %s ...\n", successIndicator, versionConfig)
			}
			return nil
		}

		fmt.Println(utils.Bold(utils.Underline("checking projects repositories:")))
		if len(filteredProjects) < 1 {
			fmt.Println("no project to check")
		}
	Loop:
		for _, p := range filteredProjects {
			switch p.Kind {
			case mono.Local:
				if isdir, _ := utils.IsDir(p.Path()); !isdir { // unexisting directory
					continue Loop // local project can exists only on another dev machine
				}
				if !git.IsRepoRootDir(p.Path()) { // just normal directory
					printCheckHeader(p, fmt.Sprintf("%s is not a git repository", p.StyledString()))
					if fix {
						setInternal(p)
					} else if interactive {
						if utils.Confirm(fmt.Sprintf("Do you want to set %s as internal ?", p.StyledString()), true) {
							setInternal(p)
						} else if utils.Confirm("Do you want to init a git repo ?", true) {
							fmt.Println("initializing a git repo...")
							git.Init(p.Path(), true)
							utils.CheckErr(app.ConfigAddOrUpdateProject(p.Name, p.RepoUrl, true))
						}
					}
					continue Loop
				}
				// check origin is not set
				var origin string
				if utils.CheckErrOrReturn(git.HasOrigin(p.Path())) {
					origin = utils.CheckErrOrReturn(git.OriginGet(p.Path()))
				}
				if origin != "" {
					printCheckHeader(p, fmt.Sprintf("origin is set to %s for local project %s", origin, p.StyledString()))
					if fix || (interactive && utils.Confirm(fmt.Sprintf("Do you want to set %s as external(%s)?", p.StyledString(), origin), true)) {
						fmt.Println("setting project as external...")
						utils.CheckErr(app.ConfigAddOrUpdateProject(p.Name, origin, true))
					} else if interactive && utils.Confirm(fmt.Sprintf("Do you want to remove %s origin?", p.StyledString()), false) {
						fmt.Println("removing origin...")
						utils.CheckErr(git.OriginRemove(p.Path()))
					}
					continue Loop
				}
			case mono.External:
				if isdir, _ := utils.IsDir(p.Path()); !isdir { // unexisting directory
					printCheckHeader(p, fmt.Sprintf("project %s does not exist", p.StyledString()))
					if fix || (interactive && utils.Confirm(fmt.Sprintf("Do you want to clone %s to %s ?", p.RepoUrl, p.StyledString()), true)) {
						fmt.Println("cloning...")
						utils.CheckErr(git.Clone(p.RepoUrl, p.Path()))
					}
					continue Loop
				}
				if !git.IsRepoRootDir(p.Path()) { // just normal directory
					printCheckHeader(p, fmt.Sprintf("%s is not a git repository", p.StyledString()))
					if fix || (interactive && utils.Confirm(fmt.Sprintf("Do you want to set %s as internal ?", p.StyledString()), true)) {
						utils.CheckErr(app.ConfigAddOrUpdateProject(p.Name, "internal", true))
						utils.FileRemoveLine(monoIgnore, p.Name)
					}
					continue Loop
				}
				origin := utils.CheckErrOrReturn(git.OriginGet(p.Path()))
				// @todo if origin empty make it a local project
				if origin != p.RepoUrl {
					printCheckHeader(p, fmt.Sprintf("origin mismatch for external project %s", p.StyledString()))
					if fix {
						fmt.Println("updating config...")
						utils.CheckErr(app.ConfigAddOrUpdateProject(p.Name, origin, true))
					} else if interactive {
						if utils.Confirm(fmt.Sprintf("Do you want to update %s config from %s to %s ?", p.StyledString(), p.RepoUrl, origin), true) {
							fmt.Println("updating config...")
							utils.CheckErr(app.ConfigAddOrUpdateProject(p.Name, origin, true))
						} else if utils.Confirm(fmt.Sprintf("Do you want to set %s remote origin to %s ?", p.StyledString(), p.RepoUrl), true) {
							fmt.Println("setting remote origin...")
							utils.CheckErr(git.OriginSet(p.Path(), p.RepoUrl))
						}
					}
					continue Loop
				}
			case mono.Internal:
				if isdir, _ := utils.IsDir(p.Path()); !isdir { // unexisting directory fix: remove
					printCheckHeader(p, fmt.Sprintf("project %s does not exist", p.StyledString()))
					if fix || (interactive && utils.Confirm(fmt.Sprintf("Do you want to remove %s project ?", p.StyledString()), true)) {
						fmt.Println("removing project...")
						utils.CheckErr(app.ConfigRemoveProject(p.Name, true))
					}
					continue Loop
				}
				if git.IsRepoRootDir(p.Path()) {
					printCheckHeader(p, fmt.Sprintf("internal project %s is a git repository", p.StyledString()))
					origin := utils.CheckErrOrReturn(git.OriginGet(p.Path()))
					if fix || (interactive && utils.Confirm(fmt.Sprintf("Do you want to set %s as external(%s)?", p.StyledString(), origin), true)) {
						fmt.Println("setting project as external...")
						utils.CheckErr(app.ConfigAddOrUpdateProject(p.Name, origin, true))
					} else if interactive && utils.Confirm(fmt.Sprintf("Do you want to set %s as internal ?\n(this is a destructive action)", p.StyledString()), false) {
						fmt.Println("setting project as internal...")
						utils.CheckErr(app.ConfigAddOrUpdateProject(p.Name, "internal", true))
						utils.CheckErr(utils.FileRemoveLine(monoIgnore, p.Name))
						utils.CheckErr(os.RemoveAll(filepath.Join(p.Path(), ".git")))
					}
					continue Loop
				}
			}
			printCheckHeader(p)
		}

		// following checks are not performed when filter is in use
		if !hasFilter {
			// check Pipeline config is correct
			fmt.Println(utils.Bold(utils.Underline("checking pipeline:")))
			tasks.GetStandardizedPipeline(false).IsAcyclic(true)
			fmt.Println(successIndicator + " pipeline ok")

			// check githooks path is correctly set
			if utils.FileExistsNoErr(filepath.Join(monoRoot, app.DfltHooksDir)) {
				fmt.Printf(utils.Bold(utils.Underline("found %s checking git core.hookspath:\n")), app.DfltHooksDir)
				hookspath, err := git.HooksPathGet(monoRoot)
				if err == nil {
					if hookspath == app.DfltHooksDir {
						fmt.Println(successIndicator + " git core.hookspath set to " + app.DfltHooksDir)
					} else {
						fmt.Println(failureIndicator + " " + utils.Warning("git core.hookspath is not set to "+app.DfltHooksDir))
						fmt.Println("You can either remove this directory or set your git config to use it:")
						fmt.Printf("git -C %s config core.hookspath %s\n", monoRoot, app.DfltHooksDir)
					}
				}
			}

			// check packagemanager version is correct
			checkJSPMVersion()
		}

		if exitStatus != 0 && !fix && !interactive {
			os.Exit(exitStatus)
		}
		fmt.Println(utils.Success("All good!"))
	},
}

func init() {
	RootCmd.AddCommand(checkCmd)
	FlagAddProjectFilter(checkCmd, false)
	checkCmd.Flags().Bool("fix", false, "Try to fix reported anomalies, disable interactive mode")
	checkCmd.Flags().BoolP("interactive", "i", false, "Prompt for action to take on each reported anomaly")
}
