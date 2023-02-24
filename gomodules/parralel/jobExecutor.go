/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/
package parallel

import (
	_ "embed"
	"fmt"
	"html/template"
	"os/exec"
	"strings"
)

type jobEventHandler func(cps JobList, jobId int)
type jobsEventHandler func(cps JobList)
type JobExecutor struct {
	jobs JobList
	opts *ExecuteOptions
}
type JobError struct {
	Id            int
	OriginalError error
}

func NewJobError(jobId int, err error) JobError {
	jobErr := JobError{Id: jobId, OriginalError: err}
	return jobErr
}
func (e *JobError) Error() string {
	return e.OriginalError.Error()
}
func (e *JobError) String() string {
	return e.OriginalError.Error()
}

type JobsError map[int]JobError

func (e JobsError) Error() string {
	return e.String()
}
func (es JobsError) String() string {
	strs := make([]string, len(es))
	for i, err := range es {
		strs[i] = err.Error()
	}
	return strings.Join(strs, "\n")
}

var parallelTemplate *template.Template

// go:embed parallel.gtpl
var dfltTemplateString string

func init() {
	SetTemplateString(dfltTemplateString)
}

func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return pad + strings.Replace(v, "\n", "\n"+pad, -1)
}
func trim(v string) string {
	return strings.Trim(v, "\n")
}

// Template for all outputs related to jobs
// it must define the following templates:
// - startSummary: which will receive a JobList
// - jobStatus: which will receive a single job
// - progressReport: which will receive a jobList
// - doneReport: which will receive a jobList
func SetTemplateString(templateString string) {
	parallelTemplate = template.Must(template.New("parallel").
		Funcs(template.FuncMap{
			"indent": indent,
			"trim":   trim,
		}).
		Parse(templateString),
	)
}

func augmentJobHandler(fn jobEventHandler, decoratorFn jobEventHandler) jobEventHandler {
	if fn == nil {
		return decoratorFn
	}
	return func(cps JobList, jobId int) {
		fn(cps, jobId)
		decoratorFn(cps, jobId)
	}
}
func augmentJobsHandler(fn jobsEventHandler, decoratorFn jobsEventHandler) jobsEventHandler {
	if fn == nil {
		return decoratorFn
	}
	return func(cps JobList) {
		fn(cps)
		decoratorFn(cps)
	}
}

// instanciate a new JobExecutor
func NewExecutor() *JobExecutor {
	executor := &JobExecutor{
		opts: &ExecuteOptions{},
	}
	return executor
}

// return the total number of jobs in the pool
func (p *JobExecutor) Len() int {
	return len(p.jobs)
}

// Add miltiple job command to execute
func (p *JobExecutor) AddJobCmds(cmdsAndArgs ...[]string) *JobExecutor {
	for _, cmdAndArgs := range cmdsAndArgs {
		p.AddJobCmd(cmdAndArgs[0], cmdAndArgs[1:]...)
	}
	return p
}
func (p *JobExecutor) AddJobCmd(cmd string, args ...string) *JobExecutor {
	p.jobs = append(p.jobs, &job{Cmd: exec.Command(cmd, args...)})
	return p
}

// Add one or more job function to run (func() (string, error))
func (p *JobExecutor) AddJobFns(fns ...runnableFn) *JobExecutor {
	for _, fn := range fns {
		p.jobs = append(p.jobs, &job{Fn: fn})
	}
	return p
}

// Add an handler which will be call after a jobs is terminated
func (p *JobExecutor) OnJobDone(fn jobEventHandler) *JobExecutor {
	p.opts.onJobDone = augmentJobHandler(p.opts.onJobDone, fn)
	return p
}

// Add an handler which will be call after all jobs are terminated
func (p *JobExecutor) OnJobsDone(fn jobsEventHandler) *JobExecutor {
	p.opts.onJobsDone = augmentJobsHandler(p.opts.onJobsDone, fn)
	return p
}

// Add an handler which will be call before a jobs is started
func (p *JobExecutor) OnJobStart(fn jobEventHandler) *JobExecutor {
	p.opts.onJobStart = augmentJobHandler(p.opts.onJobStart, fn)
	return p
}

// Add an handler which will be call before any jobs is started
func (p *JobExecutor) OnJobsStart(fn jobsEventHandler) *JobExecutor {
	p.opts.onJobsStart = augmentJobsHandler(p.opts.onJobsStart, fn)
	return p
}

// Output a summary of the job that will be run
func (p *JobExecutor) WithStartSummary() *JobExecutor {
	p.opts.onJobsStart = func(cps JobList) {
		fmt.Print(cps.execTemplate("startSummary"))
	}
	return p
}

// Output a line to say a job is starting
func (p *JobExecutor) WithStartOutput() *JobExecutor {
	p.opts.onJobStart = func(cps JobList, jobId int) {
		fmt.Print("Starting " + cps[jobId].execTemplate("jobStatusLine"))
		// fmt.Print(cps[jobId].execTemplate("jobStatusLine"))
	}
	return p
}

// Display full jobStatus as they arrive
func (p *JobExecutor) WithFifoOutput() *JobExecutor {
	p.opts.onJobDone = func(cps JobList, jobId int) {
		fmt.Print(cps[jobId].execTemplate("jobStatusFull"))
	}
	return p
}

// display doneReport when all jobs are Done
func (p *JobExecutor) WithOrderedOutput() *JobExecutor {
	p.opts.onJobsDone = func(cps JobList) {
		fmt.Print(cps.execTemplate("doneReport"))
	}
	return p
}

// will override onJobStarts / onJobStart / onJobDone handlers previsously defined
// generally you should avoid using these method with other handlers bound to the
// JobExecutor instance
func (p *JobExecutor) WithProgressOutput() *JobExecutor {
	p.opts.onJobsStart = func(cps JobList) {
		fmt.Print(cps.execTemplate("startProgressReport"))
	}
	esc := fmt.Sprintf("\033[%dA", len(p.jobs)) // clean sequence
	printProgress := func(cps JobList, jobId int) { fmt.Print(esc + cps.execTemplate("progressReport")) }
	p.opts.onJobDone = printProgress
	p.opts.onJobStart = printProgress
	return p
}

// effectively execute jobs and return collected errors as JobsError
func (p *JobExecutor) Execute() JobsError {
	var errs JobsError
	p.OnJobDone(func(jobs JobList, jobId int) {
		err := jobs[jobId].Err
		if err != nil {
			errs[jobId] = NewJobError(jobId, err)
		}
	})
	execute(p.jobs, *p.opts)
	return errs
}
