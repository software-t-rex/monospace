/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/
package parallel

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"text/template"
	"time"
)

type childProcess struct {
	cmd       *exec.Cmd
	onDone    func()
	res       []byte
	err       error
	startTime time.Time
	duration  time.Duration
}

type childProcessExport struct {
	Cmd       string
	Res       string
	Err       error
	StartTime time.Time
	Duration  time.Duration
}

type cmdList [][]string

var jobDoneTemplate *template.Template

func init() {
	t, err := template.New("jobDone").Parse(
		`{{.Cmd}} {{if .Err}}failed{{else}}succeed{{end}} in {{.Duration}}:
{{.Res}}`,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	SetJobDoneTemplate(t)
}

func SetJobDoneTemplate(t *template.Template) {
	jobDoneTemplate = t
}

func (cp *childProcess) run() {
	defer cp.onDone()
	cp.startTime = time.Now()
	cp.res, cp.err = cp.cmd.CombinedOutput()
	cp.duration = time.Since(cp.startTime)
}

func (cp *childProcess) execTemplate(t *template.Template, w io.Writer) error {
	readable := childProcessExport{fmt.Sprint(cp.cmd.Args), string(cp.res), cp.err, cp.startTime, cp.duration}
	return t.Execute(w, readable)
}

func cmdListtoChildProcess(cmdsAndArgs cmdList) []*childProcess {
	cps := make([]*childProcess, len(cmdsAndArgs))
	for i, cmdAndArgs := range cmdsAndArgs {
		cps[i] = &childProcess{
			cmd: exec.Command(cmdAndArgs[0], cmdAndArgs[1:]...),
		}
	}
	return cps
}
