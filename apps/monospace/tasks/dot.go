/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package tasks

import (
	"fmt"
	"sort"
	"strings"

	"github.com/software-t-rex/monospace/utils"
)

type taskNameList []*TaskName

func (l taskNameList) Len() int { return len(l) }
func (l taskNameList) Less(i, j int) bool {
	if l[i].Task == l[j].Task {
		return l[i].Project < l[j].Project
	}
	return l[i].Task < l[j].Task
}
func (l taskNameList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (t *TaskList) GetDot() string {
	taskGroups := []string{}
	for _, t := range t.List {
		if !utils.SliceContains(taskGroups, t.Name.Task) {
			taskGroups = append(taskGroups, t.Name.Task)
		}
	}
	groupColor := make(map[string]string, len(taskGroups))
	for i, group := range taskGroups {
		lcGroup := strings.ToLower(group)
		if strings.Contains(lcGroup, "test") {
			groupColor[group] = "1"
		} else if strings.Contains(lcGroup, "build") {
			groupColor[group] = "4"
		} else if strings.Contains(lcGroup, "format") || strings.Contains(lcGroup, "lint") {
			groupColor[group] = "2"
		} else if strings.Contains(lcGroup, "doc") {
			groupColor[group] = "3"
		} else {
			groupColor[group] = fmt.Sprintf("%d", i%8+5)
		}
	}
	out := []string{`digraph G{
	graph [bgcolor="#121212" fontcolor="black" rankdir="RL"]
	node [colorscheme="set312" style="filled,rounded" shape="box"]
	edge [color="#f0f0f0"]`}
	// append in color order
	var orderedTasks taskNameList
	for _, t := range t.List {
		orderedTasks = append(orderedTasks, &t.Name)
	}
	sort.Sort(orderedTasks)
	for _, tName := range orderedTasks {
		out = append(out, fmt.Sprintf("\t\"%s\" [color=\"%s\"]", tName.String(), groupColor[tName.Task]))
	}
	// for _, t := range t.List {
	// 	out = append(out, fmt.Sprintf("\t\"%s\" [color=\"%s\"]", t.Name.String(), groupColor[t.Name.Task]))
	// }
	for _, t := range t.List {
		for _, dep := range t.TaskDef.DependsOn {
			out = append(out, fmt.Sprintf("\t\"%s\" -> \"%s\"", t.Name.String(), dep))
		}
	}
	// finally group all nodes without dependencies
	noDepNodes := []string{}
	for _, t := range t.List {
		if len(t.TaskDef.DependsOn) == 0 {
			noDepNodes = append(noDepNodes, fmt.Sprintf("\"%s\"", t.Name.String()))
		}
	}
	if len(noDepNodes) > 1 {
		out = append(out, fmt.Sprintf("\t{rank=same; %s}", strings.Join(noDepNodes, ";")))
	}

	return strings.Join(out, "\n") + "\n}"
}
