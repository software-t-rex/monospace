/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package tasks

import (
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/software-t-rex/go-jobExecutor/v2"
	packagemanager "github.com/software-t-rex/js-packagemanager"
	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/utils"
	"github.com/software-t-rex/packageJson"
)

var exit func(msg string)
var configGet func() (*app.MonospaceConfig, error)

func init() {
	exit = utils.Exit
	configGet = app.ConfigGet
}

type TaskName struct {
	Project string
	Task    string
}

func (t *TaskName) String() string {
	return t.Project + "#" + t.Task
}

type Task struct {
	Name    TaskName
	TaskDef app.MonospaceConfigPipeline
}

func getConfig() *app.MonospaceConfig {
	return utils.CheckErrOrReturn(configGet())
}

var taskNameRegex = regexp.MustCompile("^(?:([^#]+)#)?([^#]+)$")

func parseTaskName(name string) *TaskName {
	parsedName := taskNameRegex.FindStringSubmatch(strings.TrimPrefix(name, "#"))
	var projectName string
	if parsedName == nil {
		exit("can't parse task name: " + name)
	}
	if parsedName[1] == "" || parsedName[1] == "*" {
		projectName = "*"
	} else {
		projectName = getStandardProjectName(parsedName[1]) // exit on invalid projtect name
	}
	return &TaskName{Project: projectName, Task: parsedName[2]}
	// check a task can be found
	// if _, ok := config.Pipeline[projectName+"#"+parsedName[2]]; !ok {
	// 	if _, ok := config.Pipeline[parsedName[2]]; !ok {
	// 		if _, ok := config.Pipeline["*#"+parsedName[2]]; !ok {
	// 			exit(fmt.Sprintf("can't find task %s for project %s in pipeline config", parsedName[2], projectName))
	// 		}
	// 	}
	// }
	// return &TaskName{Project: projectName, Name: parsedName[2]}
}

type Pipeline map[string]app.MonospaceConfigPipeline
type TaskList struct {
	List     map[string]*Task
	Pipeline Pipeline
}

func getStandardizedPipeline() Pipeline {
	res := make(Pipeline)
	config := getConfig()
	for k, v := range config.Pipeline {
		taskName := parseTaskName(k)
		taskDef := v
		if len(taskDef.DependsOn) > 0 {
			for i, depName := range taskDef.DependsOn {
				// todo handle ^ prefix
				if !strings.Contains(depName, "#") {
					taskDef.DependsOn[i] = taskName.Project + "#" + depName
				} else {
					depName := parseTaskName(depName)
					taskDef.DependsOn[i] = depName.String()
				}
			}
		}
		res[taskName.String()] = v
	}
	return res
}

func getStandardProjectName(name string) string {
	config, _ := configGet()
	if name == "*" || name == "" {
		return "*"
	}
	if _, ok := config.Projects[name]; ok {
		return name
	} else if aliased, ok := config.Aliases[name]; ok {
		if _, ok := config.Projects[aliased]; !ok {
			exit(fmt.Sprintf("alias %s point to unknwon project %s", name, aliased))
		}
		return aliased
	}
	exit(fmt.Sprintf("%s is neither a project name or an alias", name))
	return ""
}

// project can't be an alias there
func (p Pipeline) TaskLookup(taskName, project string) *Task {
	taskFullName := getStandardProjectName(project) + "#" + taskName
	if taskDef, ok := p[taskFullName]; ok {
		return NewTask(taskFullName, &taskDef)
	} else if taskDef, ok := p["*#"+taskName]; ok {
		return NewTask(taskFullName, &taskDef)
	}
	return nil
}

func (p Pipeline) NewTaskList() TaskList {
	return TaskList{List: make(map[string]*Task), Pipeline: p}
}

func NewTask(fullName string, taskDef *app.MonospaceConfigPipeline) *Task {
	return &Task{
		Name:    *parseTaskName(fullName),
		TaskDef: *taskDef,
	}
}

func getJSPMTaskCmd(pm *packagemanager.PackageManager, task *Task) *exec.Cmd {
	cmd := exec.Command(pm.Command, "run", task.Name.Task)
	cmd.Dir = utils.ProjectGetPath(task.Name.Project)
	return cmd
}

func (t *Task) GetJobRunner() interface{} {
	projectPath := utils.ProjectGetPath(t.Name.Project)
	if len(t.TaskDef.Cmd) > 0 {
		cmd := exec.Command(t.TaskDef.Cmd[0], t.TaskDef.Cmd[1:]...)
		cmd.Dir = projectPath
		return cmd
	}
	// check for package json script
	pjsonPath := filepath.Join(projectPath, "package.json")
	if utils.FileExistsNoErr(pjsonPath) {
		pjson, err := packageJson.Read(pjsonPath)
		// ignore error
		if err == nil && pjson.HasTask(t.Name.Task) { // we need to get the packageManager in use
			configJSPM := getConfig().JSPM
			pjsonPM := pjson.PackageManager
			var pm *packagemanager.PackageManager
			if pjsonPM == "" && configJSPM == "" { // no pm defined in pjson or config try detection
				pm, err = packagemanager.GetPackageManager(projectPath, pjson)
				if err == nil {
					return getJSPMTaskCmd(pm, t)
				}
				utils.PrintWarning("Can't found a package manager to execute task " + t.Name.Task + " in project " + t.Name.Project + " => skip")
				// no pm found, ignore pjson task
			} else if pjsonPM != "" && configJSPM != "" { // both config and pjson set a pm compare them for compatibility
				if pjsonPM == configJSPM {
					pm, err := packagemanager.GetPackageManagerFromString(configJSPM)
					if err == nil {
						return getJSPMTaskCmd(pm, t)
					}
					utils.PrintWarning("Can't found suitable package manager (" + configJSPM + ") to execute task " + t.Name.Task + " in project " + t.Name.Project + " => skip")
				} else {
					projectPM, projectErr := packagemanager.GetPackageManagerFromString(pjsonPM)
					configPM, configErr := packagemanager.GetPackageManagerFromString(configJSPM)
					if projectErr == nil && configErr == nil {
						if projectPM == configPM {
							return getJSPMTaskCmd(configPM, t)
						} else {
							utils.PrintWarning("Package manager in package.json (" + pjsonPM + ") and monospace config (" + configJSPM + ") are not compatible")
							return nil
						}
					} else if projectErr != nil && configErr == nil {
						return getJSPMTaskCmd(configPM, t)
					} else if projectErr == nil && configErr != nil {
						return getJSPMTaskCmd(projectPM, t)
					} else if projectErr != nil && configErr != nil {
						utils.PrintWarning("Can't found a package manager to execute task "+t.Name.Task+" in project "+t.Name.Project+" => skip\n", projectErr.Error(), "\n", configErr.Error())
						return nil
					}
				}
			}
		}
	}
	return nil
}

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
		depTaskName := parseTaskName(depName)
		depTask := t.Pipeline.TaskLookup(depTaskName.Task, depTaskName.Project)
		if depTask != nil {
			t.List[depName] = depTask
			t.ResolveDeps(depTask)
		}
	}
}
func (t TaskList) Len() int {
	return len(t.List)
}

func (t TaskList) GetExecutor() *jobExecutor.JobExecutor {
	e := utils.NewTaskExecutor()
	taskIds := make(map[string]int, t.Len())

	jobs := make(map[int]jobExecutor.Job, t.Len())
	for taskId, task := range t.List {
		taskRunner := task.GetJobRunner()
		if taskRunner != nil {
			job := e.AddJob(jobExecutor.NamedJob{Name: task.Name.String(), Job: taskRunner})
			taskIds[taskId] = job.Id()
			jobs[job.Id()] = job
		} else {
			utils.Exit("Don't know what to do for task " + task.Name.String())
			fmt.Printf("no task %s for project %s\n", task.Name.Task, task.Name.Project)
		}
	}
	// add dependencies
	for taskId, task := range t.List {
		for _, depTask := range task.TaskDef.DependsOn {
			if _, ok := taskIds[depTask]; !ok {
				utils.Exit(taskId + ": missing dependency task " + depTask)
			}
			e.AddJobDependency(jobs[taskIds[taskId]], jobs[taskIds[depTask]])
		}
	}
	return e
}

func prepareTaskList(tasks []string, filters []string) TaskList {
	config, err := configGet()
	if err != nil {
		exit("can't get config")
	}
	if config.Pipeline == nil || len(config.Pipeline) == 0 {
		exit("no readable pipeline in .monospace.yml")
	}
	projects := config.Projects
	aliases := config.Aliases
	pipeline := getStandardizedPipeline()
	taskList := pipeline.NewTaskList()

	var filteredProjects []string

	// projects we want to run tasks for
	if len(filters) == 0 {
		filteredProjects = utils.MapGetKeys(projects)
	} else {
		for _, filter := range filters {
			if p, ok := aliases[filter]; ok {
				filteredProjects = append(filteredProjects, p)
			} else if _, ok := projects[filter]; ok {
				filteredProjects = append(filteredProjects, filter)
			} else {
				exit("unknown project filter:" + filter)
			}
		}
	}
	// add matching task to the list for each projects
	for _, project := range filteredProjects {
		// first check for specific task for the project
		for _, taskName := range tasks {
			task := pipeline.TaskLookup(taskName, project)
			if task != nil {
				taskList.AddTask(task, true)
			}
		}
	}
	return taskList
}
func OpenGraphviz(tasks []string, filters []string) {
	taskList := prepareTaskList(tasks, filters)
	executor := taskList.GetExecutor()
	dot := executor.GetDot()
	// print the dot graph
	fmt.Println(dot)
	utils.Open("https://dreampuf.github.io/GraphvizOnline/#" + url.PathEscape(dot))
}

func Run(tasks []string, filters []string) {
	taskList := prepareTaskList(tasks, filters)
	if taskList.Len() == 0 {
		utils.Exit("no tasks found")
	}
	executor := taskList.GetExecutor()
	executor.DagExecute()
}
