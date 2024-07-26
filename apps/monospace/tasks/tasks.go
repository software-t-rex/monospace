/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package tasks

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/software-t-rex/go-jobExecutor/v2"
	jspm "github.com/software-t-rex/js-packagemanager"
	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/software-t-rex/monospace/mono"
	"github.com/software-t-rex/packageJson"
)

// use local function to mock them in tests
var exit func(msg string)

var (
	ErrInvalidProjectName = errors.New("invalid project name")
	ErrNoAvailableOption  = errors.New("no available option")
)

func init() {
	exit = utils.Exit
}

type TaskName struct {
	Project    string // the project name the task belongs to
	Task       string // the name of the task (without project)
	ConfigName string // string used in the config file
}

// returns the standardized task name string
func (t TaskName) String() string {
	return t.Project + "#" + t.Task
}

type Task struct {
	Name    TaskName
	TaskDef app.MonospaceConfigTask
}

var taskNameRegex = regexp.MustCompile("^(?:([^#]+)#)?([^#]+)$")

// ParseTaskName will parse a task name string and return a TaskName struct
//
// If provided config is used to check if the project name is valid and to replace aliases with standard names
// Will exit on failure to retrieve a valid project name
func ParseTaskName(name string, config *app.MonospaceConfig) TaskName {
	parsedName := taskNameRegex.FindStringSubmatch(strings.TrimPrefix(name, "#"))
	var projectName string
	if parsedName == nil {
		exit("can't parse task name: " + name)
	}

	if parsedName[1] == "" || parsedName[1] == "*" {
		projectName = "*"
	} else {
		var err error
		projectName, err = getStandardProjectName(parsedName[1], config) // exit on invalid projtect name
		if err != nil {
			exit(fmt.Errorf("parsing taskname %s: %w", name, err).Error())
			return TaskName{ConfigName: name} // will never happen
		}
	}
	// var err error
	// projectName, err = getStandardProjectName(parsedName[1], config)
	// if err != nil {
	// 	exit(fmt.Errorf("parsing taskname %s: %w", name, err).Error())
	// 	return TaskName{ConfigName: name} // will never happen
	// }
	return TaskName{Project: projectName, Task: parsedName[2], ConfigName: name}
}

func StandardizedTaskName(name string, config *app.MonospaceConfig) string {
	return ParseTaskName(name, config).String()
}

type Pipeline map[string]Task
type TaskList struct {
	List     map[string]*Task
	Pipeline Pipeline
	config   *app.MonospaceConfig
}

// returns a clean pipeline
// Will return an error if pipeline contains invalid tasks or dependencies references,
// or if tasks depends on persistent task
func GetStandardizedPipeline(config *app.MonospaceConfig, failEmpty bool) (Pipeline, error) {
	if config.Pipeline == nil || len(config.Pipeline) == 0 {
		if failEmpty {
			return Pipeline{}, fmt.Errorf("no readable pipeline in monospace.yml")
		} else {
			return Pipeline{}, nil
		}
	}
	app.PopulateEnv(nil)
	res := make(Pipeline)
	for k, v := range config.Pipeline {
		taskName := ParseTaskName(k, config)
		taskDef := v
		if len(taskDef.DependsOn) > 0 {
			taskDef.DependsOn = append([]string{}, v.DependsOn...)
			for i, depName := range taskDef.DependsOn {
				// todo handle ^ prefix (or not)
				if !strings.Contains(depName, "#") {
					taskDef.DependsOn[i] = taskName.Project + "#" + depName
				} else {
					depName := ParseTaskName(depName, config)
					taskDef.DependsOn[i] = depName.String()
				}
			}
		}
		res[taskName.String()] = Task{taskName, taskDef}
	}
	// check dependencies are valid (tasks exists and are not persistent tasks)
	for _, task := range res {
		for _, depName := range task.TaskDef.DependsOn {
			dep, ok := res[depName]
			if !ok {
				return Pipeline{}, fmt.Errorf("%s depends on unknown task %s", task.String(), depName)
			} else if dep.Persists() {
				return Pipeline{}, fmt.Errorf("%s can't depend on persistent task %s", task.String(), depName)
			}
		}
	}
	return res, nil
}

// if config is nil, it will not perform extra check on project name validity
func getStandardProjectName(name string, config *app.MonospaceConfig) (string, error) {
	if name == "" {
		name = "*"
	}
	if name == "*" || name == "root" {
		return name, nil
	}
	if config == nil { // no extra check without a provided config
		return name, nil
	}
	// check if name is a valid name or an alias
	if _, ok := config.Projects[name]; ok {
		return name, nil
	} else if aliased, ok := config.Aliases[name]; ok {
		if _, ok := config.Projects[aliased]; !ok {
			return "", fmt.Errorf("%w: alias %s point to unknwon project %s", ErrInvalidProjectName, name, aliased)
		}
		return aliased, nil
	}
	return "", fmt.Errorf("%w: %s is neither a project name or an alias", ErrInvalidProjectName, name)
}

//######################### Pipeline methods #########################//

func (p Pipeline) TaskLookup(taskName, project string, config *app.MonospaceConfig) *Task {
	stdProjectName, err := getStandardProjectName(project, config)
	if err != nil {
		exit(fmt.Errorf("looking for task %s: %w", taskName, err).Error())
	}
	taskFullName := stdProjectName + "#" + taskName
	if taskDef, ok := p[taskFullName]; ok {
		return NewTask(taskFullName, taskDef.TaskDef)
	} else if taskDef, ok := p["*#"+taskName]; ok {
		return NewTask(taskFullName, taskDef.TaskDef)
	}
	return nil
}

func (p Pipeline) NewTaskList(config *app.MonospaceConfig) TaskList {
	return TaskList{List: make(map[string]*Task), Pipeline: p, config: config}
}

func (p Pipeline) IsAcyclic(exitOnError bool) bool {
	length := len(p)
	dependentCount := make(map[string]int, length)
	for _, task := range p {
		for _, to := range task.TaskDef.DependsOn {
			dependentCount[to]++
		}
	}

	var queue []string
	for taskName := range p {
		if dependentCount[taskName] == 0 {
			queue = append(queue, taskName)
		}
	}
	resolved := 0
	for len(queue) > 0 {
		at := queue[0]
		queue = queue[1:]
		resolved++
		for _, to := range p[at].TaskDef.DependsOn {
			dependentCount[to]--
			if dependentCount[to] == 0 {
				queue = append(queue, to)
			}
		}
	}
	if exitOnError && resolved != length {
		exit("Pipeline circular dependencies detected")
	}
	return resolved == length
}

// Returns a new Pipeline with given task removed from pipeline and other tasks dependencies
// the original pipeline is kept unchanged
// config is used to parse task names and replace project names with aliases when possible
func (p Pipeline) RemoveTask(name string, config *app.MonospaceConfig) Pipeline {
	taskName := StandardizedTaskName(name, config)
	res := make(Pipeline)
	for k, v := range p {
		if k == taskName {
			continue
		}
		task := v.TaskDef
		task.DependsOn = utils.SliceFilter(task.DependsOn, func(s string) bool {
			return StandardizedTaskName(s, config) != taskName // @todo check we need to parse task name as it should be standardized
		})
		res[k] = Task{v.Name, task}
	}
	return res
}

// it returns a map[string]MonospaceConfigTask from the pipeline and replace project names with aliases when possible
func (p Pipeline) ToConfig(config *app.MonospaceConfig) map[string]app.MonospaceConfigTask {
	res := make(map[string]app.MonospaceConfigTask)
	projectAliases := config.GetProjectsAliases()

	for k, v := range p {
		taskName := ParseTaskName(k, config)
		var key string
		if taskName.Project == "*" {
			key = taskName.Task
		} else if alias, ok := projectAliases[taskName.Project]; ok {
			key = alias + "#" + taskName.Task
		} else {
			key = taskName.String()
		}
		taskDef := v.TaskDef
		if taskDef.DependsOn != nil && len(taskDef.DependsOn) > 0 {
			// make a copy of the dependencies to avoid modifying the original task
			taskDef.DependsOn = make([]string, len(v.TaskDef.DependsOn))
			copy(taskDef.DependsOn, v.TaskDef.DependsOn)
			// replace project names in dependencies with alias if possible
			for i, dep := range taskDef.DependsOn {
				depName := ParseTaskName(dep, config) // @todo check we need to parse task name as it should be standardized
				if depName.Project == "" {
					depName.Project = taskName.Project
				}
				if alias, ok := projectAliases[depName.Project]; ok {
					taskDef.DependsOn[i] = alias + "#" + depName.Task
				} else {
					taskDef.DependsOn[i] = depName.Project + "#" + depName.Task
				}
			}
		}
		res[key] = taskDef
	}
	return res
}

func (p Pipeline) GetDependableTasks(excludedTasks []string, config *app.MonospaceConfig) []string {
	if len(p) == 0 {
		return []string{}
	}
	// standardize excluded tasks names
	if excludedTasks == nil {
		excludedTasks = []string{}
	} else {
		excludedTasks = utils.SliceMap(excludedTasks, func(task string) string {
			return StandardizedTaskName(task, config)
		})
	}
	// get all tasks that are not persistent and not excluded
	filteredPipeline := utils.MapFilter(p, func(task Task) bool {
		if utils.SliceContains(excludedTasks, task.Name.String()) {
			return false
		}
		return !task.TaskDef.Persistent
	})
	dependables := utils.MapGetKeys(filteredPipeline)
	return dependables
}

//######################### Task methods #########################//

func NewTask(fullName string, taskDef app.MonospaceConfigTask) *Task {
	return &Task{
		Name:    ParseTaskName(fullName, nil),
		TaskDef: taskDef,
	}
}

func (t *Task) Persists() bool {
	return t.TaskDef.Persistent
}
func (t *Task) DependsOn(taskName TaskName) bool {
	for _, dep := range t.TaskDef.DependsOn {
		if dep == taskName.String() {
			return true
		}
	}
	return false
}
func (t *Task) String() string {
	return t.Name.String()
}

func (t *Task) preparedCmd(cmdAndArgs ...string) *exec.Cmd {
	args := utils.SliceMap(cmdAndArgs, os.ExpandEnv)
	if _, err := exec.LookPath(args[0]); err != nil {
		// lookup for command in .monospace/bin
		binPath := filepath.Join(mono.SpaceGetRoot(), ".monospace", "bin", args[0])
		if utils.FileExistsNoErr(filepath.Join(mono.SpaceGetRoot(), ".monospace", "bin", args[0])) {
			args[0] = binPath
		}
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Path = os.ExpandEnv(cmd.Path)
	cmd.Args = utils.SliceMap(cmd.Args, os.ExpandEnv)
	cmd.Dir = mono.ProjectGetPath(t.Name.Project)
	return cmd
}
func (t *Task) preparedJSPMRunCmd(pmCmd string, args []string) *exec.Cmd {
	return t.preparedCmd(append(append([]string{pmCmd}, "run", t.Name.Task), args...)...)
}
func (t *Task) getJSPMCmdFromJSPMConfig(PMconfig string, args []string, printWarning bool) *exec.Cmd {
	pm, err := jspm.GetPackageManagerFromString(PMconfig)
	if err == nil {
		return t.preparedJSPMRunCmd(pm.Command, args)
	} else if printWarning {
		utils.PrintWarning("Can't find suitable package manager ("+PMconfig+") to execute task "+t.Name.Task+" in project "+t.Name.Project+" => skip", err.Error())
	}
	return nil
}

func (t *Task) GetJobRunner(additionalArgs []string, defaultJSPM string) *exec.Cmd {
	projectPath := mono.ProjectGetPath(t.Name.Project)
	if len(t.TaskDef.Cmd) > 0 {
		return t.preparedCmd(append(t.TaskDef.Cmd, additionalArgs...)...)
	}
	// check for package json script
	pjsonPath := filepath.Join(projectPath, "package.json")
	if utils.FileExistsNoErr(pjsonPath) {
		pjson, err := packageJson.Read(pjsonPath)
		// ignore error
		if err == nil && pjson.HasTask(t.Name.Task) { // we need to get the packageManager in use
			pjsonPM := pjson.PackageManager
			var pm *jspm.PackageManager
			if pjsonPM == "" && defaultJSPM == "" { // no pm defined in pjson or config try detection
				pm, err = jspm.GetPackageManager(projectPath, pjson)
				if err == nil {
					return t.preparedJSPMRunCmd(pm.Command, additionalArgs)
					// t.preparedCmd(pm.Command, "run", t.Name.Task)
				}
				// no pm found, ignore pjson task
				utils.PrintWarning("Can't find a package manager to execute task "+t.Name.Task+" in project "+t.Name.Project+" => skip", err.Error())
			} else if pjsonPM == defaultJSPM { // both config set same pm
				return t.getJSPMCmdFromJSPMConfig(defaultJSPM, additionalArgs, true)
			} else if pjsonPM != "" && defaultJSPM != "" { // both config and pjson set a pm compare them for compatibility
				projectPM, projectErr := jspm.GetPackageManagerFromString(pjsonPM)
				configPM, configErr := jspm.GetPackageManagerFromString(defaultJSPM)
				if projectErr == nil && configErr == nil {
					if projectPM == configPM {
						return t.preparedJSPMRunCmd(configPM.Command, additionalArgs)
						// t.preparedCmd(configPM.Command, "run", t.Name.Task)
					} else {
						utils.PrintWarning("Package manager in package.json (" + pjsonPM + ") and monospace config (" + defaultJSPM + ") are not compatible => skip until manual resolution")
						return nil
					}
				} else if configErr == nil {
					return t.preparedJSPMRunCmd(configPM.Command, additionalArgs)
					// t.preparedCmd(configPM.Command, "run", t.Name.Task)
				} else if projectErr == nil {
					return t.preparedJSPMRunCmd(projectPM.Command, additionalArgs)
					//t.preparedCmd(projectPM.Command, "run", t.Name.Task)
				} else { // both pm are invalid
					utils.PrintWarning("Can't find a package manager to execute task "+t.Name.Task+" in project "+t.Name.Project+" => skip\n", projectErr.Error(), "\n", configErr.Error())
					return nil
				}
			} else if defaultJSPM != "" { // use PM from config
				return t.getJSPMCmdFromJSPMConfig(defaultJSPM, additionalArgs, true)
			} else { // use PM from package.json
				return t.getJSPMCmdFromJSPMConfig(pjsonPM, additionalArgs, true)
			}
		}
	}
	return nil
}

//######################### TaskList methods #########################//

// add a task and resolve its dependencies
func (t TaskList) AddTask(task *Task, resolveDeps bool) TaskList {
	t.List[task.Name.String()] = task
	if resolveDeps {
		t.ResolveDeps(task)
	}
	return t
}

func (t TaskList) ResolveDeps(task *Task) {
	if len(task.TaskDef.DependsOn) < 1 {
		return
	}
	for _, depName := range task.TaskDef.DependsOn {
		if _, ok := t.List[depName]; ok { // dependency is already in the list
			continue
		}
		depTaskName := ParseTaskName(depName, t.config)
		depTask := t.Pipeline.TaskLookup(depTaskName.Task, depTaskName.Project, t.config)
		if depTask != nil {
			t.List[depName] = depTask
			t.ResolveDeps(depTask)
		}
	}
}
func (t TaskList) Len() int {
	return len(t.List)
}

func (t TaskList) GetExecutor(additionalArgs []string, outputMode string) *jobExecutor.JobExecutor {
	e := NewExecutor(outputMode)
	projectAliases := t.config.GetProjectsAliases()
	taskIds := make(map[string]int, t.Len())

	jobs := make(map[int]jobExecutor.Job, t.Len())
	for taskId, task := range t.List {
		taskRunner := task.GetJobRunner(additionalArgs, t.config.JSPM)
		taskName := task.Name.String()
		if outputMode == "interleaved" {
			// replace task name with alias if any when using interleaved output
			alias, hasAlias := projectAliases[task.Name.Project]
			if hasAlias {
				taskName = alias + "#" + task.Name.Task
			}
		}
		if taskRunner != nil {
			job := e.AddJob(jobExecutor.NamedJob{Name: taskName, Job: taskRunner})
			taskIds[taskId] = job.Id()
			jobs[job.Id()] = job
		} else if task.TaskDef.DependsOn != nil && len(task.TaskDef.DependsOn) > 0 {
			fmt.Printf(ui.GetTheme().Info("%s#%s is a dummy task, will only executes its dependencies.\n"), task.Name.Project, task.Name.Task)
			job := e.AddJob(jobExecutor.NamedJob{Name: taskName, Job: func() (string, error) { return "", nil }})
			taskIds[taskId] = job.Id()
			jobs[job.Id()] = job
		} else {
			// no cmd and no dependencies
			exit(taskName + " task has no cmd, no package.json script or dependencies, provide at least one of those or remove the task from pipeline.")
		}
	}
	// add dependencies
	for taskId, task := range t.List {
		for _, depTask := range task.TaskDef.DependsOn {
			if _, ok := taskIds[depTask]; !ok {
				exit(taskId + ": missing dependency task " + depTask)
			}
			e.AddJobDependency(jobs[taskIds[taskId]], jobs[taskIds[depTask]])
		}
	}
	return e
}

// This function will prepare a task list from a list of task names and a list of projects to search tasks for
// It will exit on failure
func PrepareTaskList(tasks []string, projects []mono.Project, config *app.MonospaceConfig) TaskList {
	pipeline, err := GetStandardizedPipeline(config, true)
	if err != nil {
		exit(err.Error())
	}

	taskList := pipeline.NewTaskList(config)

	var filteredProjects []string
	for _, project := range projects {
		filteredProjects = append(filteredProjects, project.Name)
	}

	// add matching task to the list for each projects
	for _, project := range filteredProjects {
		// first check for specific task for the project
		for _, taskName := range tasks {
			task := pipeline.TaskLookup(taskName, project, config)
			if task != nil {
				taskList.AddTask(task, true)
			}
		}
	}
	return taskList
}
func OpenGraphviz(taskList TaskList) {
	dot := taskList.GetDot()
	// print the dot graph
	fmt.Println(dot)
	utils.Open("https://dreampuf.github.io/GraphvizOnline/#" + url.PathEscape(dot))
}

func Run(taskList TaskList, additionalArgs []string, outputMode string) {
	if taskList.Len() == 0 {
		exit("no tasks found")
	}
	executor := taskList.GetExecutor(additionalArgs, outputMode)
	err := executor.DagExecute()
	if err.Len() > 0 {
		if errors.Is(err[0], jobExecutor.ErrCyclicDependencyDetected) {
			exit(err[0].Error())
		}
		os.Exit(1)
	}
}

func OpenGraphvizFull(config *app.MonospaceConfig) {
	pipeline, err := GetStandardizedPipeline(config, true)
	if err != nil {
		exit(err.Error())
	}
	taskList := pipeline.NewTaskList(config)
	taskNames := []string{}
	for taskName := range pipeline {
		if !utils.SliceContains(taskNames, taskName) {
			taskNames = append(taskNames, ParseTaskName(taskName, config).Task)
		}
	}
	for project := range config.Projects {
		// first check for specific task for the project
		for _, taskName := range taskNames {
			task := pipeline.TaskLookup(taskName, project, config)
			if task != nil {
				taskList.AddTask(task, true)
			}
		}
	}
	dot := taskList.GetDot()
	fmt.Println(dot)
	utils.Open("https://dreampuf.github.io/GraphvizOnline/#" + url.PathEscape(dot))
}
