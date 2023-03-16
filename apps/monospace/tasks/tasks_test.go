package tasks

import (
	"reflect"
	"testing"

	"github.com/software-t-rex/monospace/app"
)

func init() {
	// stub external functions for tests
	exit = func(msg string) {
		panic(msg)
	}
	configGet = func() (*app.MonospaceConfig, error) {
		return &app.MonospaceConfig{
			Projects: map[string]string{
				"apps/internalapp": "internal",
				"apps/localapp":    "local",
			},
			Aliases: map[string]string{
				"int":     "apps/internalapp",
				"invalid": "apps/unknown",
			},
			Pipeline: map[string]app.MonospaceConfigPipeline{
				"task": {},
				"build": {
					DependsOn: []string{"test", "task"},
				},
				"test": {
					DependsOn: []string{"int#task", "apps/localapp#test"},
				},
				"apps/localapp#test": {},
				"int#task": {
					DependsOn: []string{"apps/localapp#test"},
				},
			},
		}, nil
	}
}

func Test_parseTaskName(t *testing.T) {

	type args struct {
		name string
	}
	tests := []struct {
		name       string
		args       args
		want       *TaskName
		shouldExit bool
	}{
		{"should set project to * if unspecified", args{name: "task"}, &TaskName{Task: "task", Project: "*"}, false},
		{"should set project to * if blank", args{name: "#task"}, &TaskName{Task: "task", Project: "*"}, false},
		{"should exit on invalid task name", args{name: "inc#"}, &TaskName{}, true},
		{"should correctly separate project and name", args{name: "apps/internalapp#task"}, &TaskName{Task: "task", Project: "apps/internalapp"}, false},
		{"should replace  alias with project name", args{name: "int#task"}, &TaskName{Task: "task", Project: "apps/internalapp"}, false},
		{"should exit on invalid project", args{name: "unknown#task"}, &TaskName{}, true},
		{"should exit on invalid alias", args{name: "invalid#task"}, &TaskName{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.shouldExit {
					t.Fatalf("Unattended called to exit: " + r.(string))
				} else if r == nil && tt.shouldExit {
					t.Fatalf("Should have call exit")
				}
			}()
			if got := parseTaskName(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTaskName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getStandardizedPipeline(t *testing.T) {
	tests := []struct {
		name string
		want Pipeline
	}{
		{"should return standardized pipeline", Pipeline{
			"*#task": {},
			"*#build": {
				DependsOn: []string{"*#test", "*#task"},
			},
			"*#test": {
				DependsOn: []string{"apps/internalapp#task", "apps/localapp#test"},
			},
			"apps/localapp#test": {},
			"apps/internalapp#task": {
				DependsOn: []string{"apps/localapp#test"},
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStandardizedPipeline(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getStandardizedPipeline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipeline_TaskLookup(t *testing.T) {
	pipeline := getStandardizedPipeline()
	type args struct {
		name    string
		project string
	}
	tests := []struct {
		name       string
		args       args
		want       *Task
		shouldExit bool
	}{
		{"should return nil with unknown task name", args{name: "unknown", project: "apps/internalapp"}, nil, false},
		{"should return exit with project name", args{name: "task", project: "apps/unknown"}, nil, true},
		{"should return top level task with empty project name", args{name: "task", project: ""}, &Task{
			Name: TaskName{Task: "task", Project: "*"}, TaskDef: pipeline["*#task"],
		}, false},
		{"should return top level task with known task and * project name", args{name: "task", project: "*"}, &Task{
			Name: TaskName{Task: "task", Project: "*"}, TaskDef: pipeline["*#task"],
		}, false},
		{"should return task with known task and valid project name", args{name: "task", project: "apps/localapp"}, &Task{
			Name: TaskName{Task: "task", Project: "apps/localapp"}, TaskDef: pipeline["*#task"],
		}, false},
		{"should return task with known task and valid project name", args{name: "test", project: "apps/localapp"}, &Task{
			Name: TaskName{Task: "test", Project: "apps/localapp"}, TaskDef: pipeline["apps/localapp#test"],
		}, false},
		{"should return specific task with known task and valid project alias", args{name: "task", project: "int"}, &Task{
			Name: TaskName{Task: "task", Project: "apps/internalapp"}, TaskDef: pipeline["apps/internalapp#task"],
		}, false},
		{"should return top level task with known task and valid project alias", args{name: "build", project: "int"}, &Task{
			Name: TaskName{Task: "build", Project: "apps/internalapp"}, TaskDef: pipeline["*#build"],
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.shouldExit {
					t.Fatalf("Unattended called to exit: " + r.(string))
				} else if r == nil && tt.shouldExit {
					t.Fatalf("Should have call exit")
				}
			}()
			if got := pipeline.TaskLookup(tt.args.name, tt.args.project); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pipeline.TaskLookup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskList_ResolveDeps(t *testing.T) {
	pipeline := getStandardizedPipeline()
	taskTLTask := pipeline.TaskLookup("task", "*")
	// taskTLBuild := pipeline.TaskLookup("build", "*")
	taskInernalBuild := pipeline.TaskLookup("build", "apps/internalapp")
	taskTLTest := pipeline.TaskLookup("test", "*")
	taskLocalTest := pipeline.TaskLookup("test", "apps/localapp")
	taskInternalTask := pipeline.TaskLookup("task", "int")
	type args struct {
		task    string
		project string
	}
	tests := []struct {
		name string
		args args
		want TaskList
	}{
		{"should resolve dependencies", args{task: "task", project: "apps/internalapp"}, pipeline.NewTaskList().
			AddTask(taskInternalTask, false).AddTask(taskLocalTest, false),
		},
		{"should resolve dependencies", args{task: "build", project: "apps/internalapp"}, pipeline.NewTaskList().
			AddTask(taskInernalBuild, false).AddTask(taskTLTask, false).AddTask(taskTLTest, false).
			AddTask(taskInternalTask, false).AddTask(taskLocalTest, false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskList := pipeline.NewTaskList()
			task := pipeline.TaskLookup(tt.args.task, tt.args.project)
			if task == nil {
				t.Fatalf("can't find task %s#%s", tt.args.task, tt.args.project)
			}
			taskList.AddTask(task, true)
			if taskList.Len() != tt.want.Len() {
				t.Errorf("task dependency resolution mismatch got %v, want %v", taskList.Len(), tt.want.Len())
			}
			for key, wantTask := range tt.want.List {
				if gotTask, ok := taskList.List[key]; !ok {
					t.Errorf("task dependency resolution mismatch want %s %v, got %v", key, wantTask, gotTask)
				}
			}
		})
	}
}
