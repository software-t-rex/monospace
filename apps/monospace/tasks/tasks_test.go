package tasks

import (
	"reflect"
	"slices"
	"testing"

	"github.com/software-t-rex/monospace/app"
)

var testConfig = &app.MonospaceConfig{
	Projects: map[string]string{
		"apps/internalapp": "internal",
		"apps/localapp":    "local",
	},
	Aliases: map[string]string{
		"local":   "apps/localapp",
		"int":     "apps/internalapp",
		"invalid": "apps/unknown",
	},
	Pipeline: map[string]app.MonospaceConfigTask{
		"task": {},
		"build": {
			DependsOn: []string{"test", "task"},
		},
		"test": {
			DependsOn: []string{"int#task", "apps/localapp#test"},
		},
		"apps/localapp#test": {},
		"local#indep":        {},
		"unprefix_indep": {
			Persistent: true,
		},
		"int#task": {
			DependsOn: []string{"apps/localapp#test"},
		},
	},
}

func init() {
	// stub external functions for tests
	exit = func(msg string) {
		panic(msg)
	}
}

func Test_parseTaskName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name       string
		args       args
		want       TaskName
		shouldExit bool
	}{
		{"should set project to * if unspecified", args{name: "task"}, TaskName{Task: "task", Project: "*", ConfigName: "task"}, false},
		{"should set project to * if blank", args{name: "#task"}, TaskName{Task: "task", Project: "*", ConfigName: "#task"}, false},
		{"should exit on invalid task name", args{name: "inc#"}, TaskName{ConfigName: "inc#"}, true},
		{"should correctly separate project and name", args{name: "apps/internalapp#task"}, TaskName{Task: "task", Project: "apps/internalapp", ConfigName: "apps/internalapp#task"}, false},
		{"should replace  alias with project name", args{name: "int#task"}, TaskName{Task: "task", Project: "apps/internalapp", ConfigName: "int#task"}, false},
		{"should exit on invalid project", args{name: "unknown#task"}, TaskName{ConfigName: "unknown#task"}, true},
		{"should exit on invalid alias", args{name: "invalid#task"}, TaskName{ConfigName: "invalid#task"}, true},
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
			if got := ParseTaskName(tt.args.name, testConfig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTaskName() = %#v, want %#v", got, tt.want)
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
			"*#task": Task{
				Name: TaskName{Task: "task", Project: "*", ConfigName: "task"},
			},
			"*#build": Task{
				Name: TaskName{Task: "build", Project: "*", ConfigName: "build"},
				TaskDef: app.MonospaceConfigTask{
					DependsOn: []string{"*#test", "*#task"},
				},
			},
			"*#test": Task{
				Name: TaskName{Task: "test", Project: "*", ConfigName: "test"},
				TaskDef: app.MonospaceConfigTask{
					DependsOn: []string{"apps/internalapp#task", "apps/localapp#test"},
				},
			},
			"apps/localapp#test": Task{
				Name: TaskName{Task: "test", Project: "apps/localapp", ConfigName: "apps/localapp#test"},
			},
			"apps/localapp#indep": Task{
				Name: TaskName{Task: "indep", Project: "apps/localapp", ConfigName: "local#indep"},
			},
			"*#unprefix_indep": Task{
				Name:    TaskName{Task: "unprefix_indep", Project: "*", ConfigName: "unprefix_indep"},
				TaskDef: app.MonospaceConfigTask{Persistent: true},
			},
			"apps/internalapp#task": Task{
				Name: TaskName{Task: "task", Project: "apps/internalapp", ConfigName: "int#task"},
				TaskDef: app.MonospaceConfigTask{
					DependsOn: []string{"apps/localapp#test"},
				},
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := GetStandardizedPipeline(testConfig, true)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getStandardizedPipeline() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestPipeline_TaskLookup(t *testing.T) {
	pipeline, _ := GetStandardizedPipeline(testConfig, true)
	topTask := pipeline["*#task"]
	topTest := pipeline["*#test"]
	topBuild := pipeline["*#build"]
	localTest := pipeline["apps/localapp#test"]
	internalTask := pipeline["apps/internalapp#task"]
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
		{"should return top level task with empty project name", args{name: "task", project: ""}, &topTask, false},
		{"should return top level task with known task and * project name", args{name: "task", project: "*"}, &topTask, false},
		{"should return task with known task and valid project name", args{name: "task", project: "apps/localapp"}, &topTask, false},
		{"should return task with known task and valid project name", args{name: "test", project: "apps/localapp"}, &localTest, false},
		{"should return nil task with known task and valid project alias", args{name: "task", project: "int"}, &internalTask, false},
		{"should return top level task with known task and valid project alias", args{name: "build", project: "int"}, &topBuild, false},
		{"should return top level task with known *#task with valid project", args{name: "test", project: "int"}, &topTest, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.shouldExit {
					t.Fatalf("Unattended called to exit: %s", r.(string))
				} else if r == nil && tt.shouldExit {
					t.Fatalf("Should have call exit")
				}
			}()
			got := pipeline.TaskLookup(tt.args.name, tt.args.project, testConfig)
			if tt.want == nil && got != nil {
				t.Errorf("Pipeline.TaskLookup() = %v, want %v", got.String(), tt.want.String())
			} else if tt.want != nil && !reflect.DeepEqual(got.TaskDef, tt.want.TaskDef) {
				t.Errorf("Pipeline.TaskLookup() = %v, want %v, %#v", got.String(), tt.want.String(), got)
			}
		})
	}
}

func TestPipeline_RemoveTask(t *testing.T) {
	pipeline, _ := GetStandardizedPipeline(testConfig, true)
	makeTestPipeline := func(configPipeline map[string]app.MonospaceConfigTask) Pipeline {
		cfg := *testConfig
		cfg.Pipeline = configPipeline
		p, _ := GetStandardizedPipeline(&cfg, true)
		return p
	}

	tests := []struct {
		name     string
		taskName string
		want     Pipeline
	}{
		{"should remove independent unprefixed task", "unprefix_indep", makeTestPipeline(map[string]app.MonospaceConfigTask{
			"task":               {},
			"build":              {DependsOn: []string{"test", "task"}},
			"test":               {DependsOn: []string{"int#task", "apps/localapp#test"}},
			"apps/localapp#test": {},
			"local#indep":        {},
			// "unprefix_indep":     {},
			"int#task": {DependsOn: []string{"apps/localapp#test"}},
		})},
		{"should remove independent prefixed task", "apps/localapp#indep", makeTestPipeline(map[string]app.MonospaceConfigTask{
			"task":               {},
			"build":              {DependsOn: []string{"test", "task"}},
			"test":               {DependsOn: []string{"int#task", "apps/localapp#test"}},
			"apps/localapp#test": {},
			// "local#indep":        {},
			"unprefix_indep": {Persistent: true},
			"int#task":       {DependsOn: []string{"apps/localapp#test"}},
		})},
		{"should remove dependency prefixed task", "apps/internalapp#task", makeTestPipeline(map[string]app.MonospaceConfigTask{
			"task":               {},
			"build":              {DependsOn: []string{"test", "task"}},
			"test":               {DependsOn: []string{"apps/localapp#test"}},
			"apps/localapp#test": {},
			"local#indep":        {},
			"unprefix_indep":     {Persistent: true},
			// "int#task":           {DependsOn: []string{"apps/localapp#test"}},
		})},
		{"should remove dependency unprefixed task", "test", makeTestPipeline(map[string]app.MonospaceConfigTask{
			"task":  {},
			"build": {DependsOn: []string{"task"}},
			// "test":               {DependsOn: []string{"int#task", "apps/localapp#test"}},
			"apps/localapp#test": {},
			"local#indep":        {},
			"unprefix_indep":     {Persistent: true},
			"int#task":           {DependsOn: []string{"apps/localapp#test"}},
		})},
		{"should remove task from alias", "apps/localapp#test", makeTestPipeline(map[string]app.MonospaceConfigTask{
			"task":  {},
			"build": {DependsOn: []string{"test", "task"}},
			"test":  {DependsOn: []string{"int#task"}},
			// "apps/localapp#test": {},
			"local#indep":    {},
			"unprefix_indep": {Persistent: true},
			"int#task":       {},
		})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newPipeline := pipeline.RemoveTask(tt.taskName, testConfig)
			// test newPipeline and pipeline are different
			if &pipeline == &newPipeline {
				t.Errorf("Pipeline.RemoveTask() should return a new pipeline")
			}
			if !reflect.DeepEqual(newPipeline, tt.want) {
				t.Errorf("Pipeline.RemoveTask() = %#v, want %#v", pipeline, tt.want)
			}
		})
	}
}

func TestPipeline_GetDependableTasks(t *testing.T) {
	pipeline, _ := GetStandardizedPipeline(testConfig, true)
	tests := []struct {
		name     string
		excluded []string
		want     []string
	}{
		{
			"should return all dependable tasks",
			nil,
			[]string{"*#build", "*#test", "apps/internalapp#task", "*#task", "apps/localapp#test", "apps/localapp#indep"},
		},
		{
			"should return all dependable tasks not excluded",
			[]string{"build", "int#task"},
			[]string{"*#test", "*#task", "apps/localapp#test", "apps/localapp#indep"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pipeline.GetDependableTasks(tt.excluded, testConfig)
			want := tt.want
			slices.Sort(got)
			slices.Sort(want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Pipeline.GetDependableTasks() = %v, want %v", got, want)
			}
		})
	}

}

func TestPipeline_ToConfig(t *testing.T) {
	pipeline, _ := GetStandardizedPipeline(testConfig, true)
	tests := []struct {
		name string
		want map[string]app.MonospaceConfigTask
	}{
		{"should convert pipeline to config", map[string]app.MonospaceConfigTask{
			"task":           {},
			"build":          {DependsOn: []string{"*#test", "*#task"}},
			"test":           {DependsOn: []string{"int#task", "local#test"}},
			"local#test":     {},
			"local#indep":    {},
			"unprefix_indep": {Persistent: true},
			"int#task":       {DependsOn: []string{"local#test"}},
		},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pipeline.ToConfig(testConfig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pipeline.ToConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPipeline_IsAcyclic(t *testing.T) {
	testPipeline, _ := GetStandardizedPipeline(testConfig, true)
	var cyclicPipeline Pipeline = map[string]Task{}
	cyclicPipeline["myproject#build"] = Task{
		Name:    TaskName{Task: "build", Project: "myproject"},
		TaskDef: app.MonospaceConfigTask{DependsOn: []string{"myproject#test"}},
	}
	cyclicPipeline["myproject#test"] = Task{
		Name:    TaskName{Task: "test", Project: "myproject"},
		TaskDef: app.MonospaceConfigTask{DependsOn: []string{"myproject#build"}},
	}
	tests := []struct {
		name     string
		pipeline Pipeline
		want     bool
	}{
		{"should return true for acyclic pipeline", testPipeline, true},
		{"should return false for acyclic pipeline", cyclicPipeline, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pipeline.IsAcyclic(false); got != tt.want {
				t.Errorf("Pipeline.IsAcyclic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskList_ResolveDeps(t *testing.T) {
	pipeline, _ := GetStandardizedPipeline(testConfig, true)
	taskTLTask := pipeline.TaskLookup("task", "*", testConfig)
	// taskTLBuild := pipeline.TaskLookup("build", "*")
	taskInernalBuild := pipeline.TaskLookup("build", "apps/internalapp", testConfig)
	taskTLTest := pipeline.TaskLookup("test", "*", testConfig)
	taskLocalTest := pipeline.TaskLookup("test", "apps/localapp", testConfig)
	taskInternalTask := pipeline.TaskLookup("task", "int", testConfig)
	type args struct {
		task    string
		project string
	}
	tests := []struct {
		name string
		args args
		want TaskList
	}{
		{"should resolve dependencies", args{task: "task", project: "apps/internalapp"}, pipeline.NewTaskList(testConfig).
			AddTask(taskInternalTask, false).AddTask(taskLocalTest, false),
		},
		{"should resolve dependencies", args{task: "build", project: "apps/internalapp"}, pipeline.NewTaskList(testConfig).
			AddTask(taskInernalBuild, false).AddTask(taskTLTask, false).AddTask(taskTLTest, false).
			AddTask(taskInternalTask, false).AddTask(taskLocalTest, false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskList := pipeline.NewTaskList(testConfig)
			task := pipeline.TaskLookup(tt.args.task, tt.args.project, testConfig)
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
