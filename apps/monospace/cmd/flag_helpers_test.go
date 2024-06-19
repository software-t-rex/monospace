package cmd

import (
	"reflect"
	"testing"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/mono"
)

func Test_getFilteredProjects(t *testing.T) {
	testConfig := &app.MonospaceConfig{
		Projects: map[string]string{
			"test1": "internal",
			"test2": "internal",
			"test3": "internal",
		},
		Aliases: map[string]string{
			"first": "test1",
		},
	}
	testProjects := mono.ProjectsAsStructs(testConfig.Projects)
	type args struct {
		config      *app.MonospaceConfig
		filters     []string
		includeRoot bool
	}
	p1 := testProjects[0]
	p2 := testProjects[1]
	p3 := testProjects[2]
	tests := []struct {
		name string
		args args
		want []mono.Project
	}{
		{
			"should return all projects without filters",
			args{testConfig, []string{}, false},
			testProjects,
		},
		{
			"should return all projects + root without filters when includeRoot",
			args{testConfig, []string{}, true},
			[]mono.Project{mono.RootProject, p1, p2, p3},
		},
		{
			"should return only root with root filter",
			args{testConfig, []string{"root"}, false},
			[]mono.Project{mono.RootProject},
		},
		{
			"should return only root with root filter and includeRoot",
			args{testConfig, []string{"root"}, true},
			[]mono.Project{mono.RootProject},
		},
		{
			"should return only filtered projects with project filter",
			args{testConfig, []string{"test2", "test3"}, false},
			[]mono.Project{p2, p3},
		},
		{
			"should return only filtered projects with alias project filter",
			args{testConfig, []string{"first", "test3"}, false},
			[]mono.Project{p1, p3},
		},
		{
			"should return only filtered projects (not root) with project filter and includeRoot",
			args{testConfig, []string{"test2", "test3"}, true},
			[]mono.Project{p2, p3},
		},
		{
			"should return all but excluded filtered projects with project exclude filter only",
			args{testConfig, []string{"!test2"}, false},
			[]mono.Project{p1, p3},
		},
		{
			"should return all but excluded filtered projects + root with project exclude filter and includeRoot",
			args{testConfig, []string{"!test2"}, true},
			[]mono.Project{mono.RootProject, p1, p3},
		},
		{
			"should return only fitlered project that are not excluded",
			args{testConfig, []string{"test1", "test2", "!test2"}, false},
			[]mono.Project{p1},
		},
		{
			"should return only fitlered project that are not excluded by alias",
			args{testConfig, []string{"!first"}, false},
			[]mono.Project{p2, p3},
		},
		{
			"should return only fitlered project that are not excluded even with includeRoot",
			args{testConfig, []string{"test1", "test2", "!test2"}, true},
			[]mono.Project{p1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFilteredProjects(testConfig, tt.args.filters, tt.args.includeRoot); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFilteredProjects() = %#v, want %#v, equal %t", got, tt.want, reflect.DeepEqual(got, tt.want))
			}
		})
	}
}
