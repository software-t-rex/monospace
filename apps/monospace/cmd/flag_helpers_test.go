package cmd

import (
	"reflect"
	"testing"

	"github.com/software-t-rex/monospace/utils"
)

func Test_getFilteredProjects(t *testing.T) {
	// @todo should test for aliases replacement too
	type args struct {
		projects    []utils.Project
		filters     []string
		includeRoot bool
	}
	tests := []struct {
		name string
		args args
		want []utils.Project
	}{
		{
			"should return all projects without filters",
			args{[]utils.Project{{Name: "test1"}, {Name: "test2"}}, []string{}, false},
			[]utils.Project{{Name: "test1"}, {Name: "test2"}},
		},
		{
			"should return all projects + root without filters when includeRoot",
			args{[]utils.Project{{Name: "test1"}, {Name: "test2"}}, []string{}, true},
			[]utils.Project{utils.RootProject, {Name: "test1"}, {Name: "test2"}},
		},
		{
			"should return only root with root filter",
			args{[]utils.Project{{Name: "test1"}, {Name: "test2"}}, []string{"root"}, false},
			[]utils.Project{utils.RootProject},
		},
		{
			"should return only root with root filter and includeRoot",
			args{[]utils.Project{{Name: "test1"}, {Name: "test2"}}, []string{"root"}, true},
			[]utils.Project{utils.RootProject},
		},
		{
			"should return only filtered projects with project filter",
			args{[]utils.Project{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}}, []string{"test2", "test3"}, false},
			[]utils.Project{{Name: "test2"}, {Name: "test3"}},
		},
		{
			"should return only filtered projects (not root) with project filter and includeRoot",
			args{[]utils.Project{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}}, []string{"test2", "test3"}, true},
			[]utils.Project{{Name: "test2"}, {Name: "test3"}},
		},
		{
			"should return all but excluded filtered projects with project exclude filter only",
			args{[]utils.Project{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}}, []string{"!test2"}, false},
			[]utils.Project{{Name: "test1"}, {Name: "test3"}},
		},
		{
			"should return all but excluded filtered projects + root with project exclude filter and includeRoot",
			args{[]utils.Project{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}}, []string{"!test2"}, true},
			[]utils.Project{utils.RootProject, {Name: "test1"}, {Name: "test3"}},
		},
		{
			"should return only fitlered project that are not excluded",
			args{[]utils.Project{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}}, []string{"test1", "test2", "!test2"}, false},
			[]utils.Project{{Name: "test1"}},
		},
		{
			"should return only fitlered project that are not excluded even with includeRoot",
			args{[]utils.Project{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}}, []string{"test1", "test2", "!test2"}, true},
			[]utils.Project{{Name: "test1"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFilteredProjects(tt.args.projects, tt.args.filters, tt.args.includeRoot); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFilteredProjects() = %q, want %q", got, tt.want)
			}
		})
	}
}
