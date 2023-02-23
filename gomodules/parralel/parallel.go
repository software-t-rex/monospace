/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/
package parallel

import (
	"runtime"
	"sync"
)

var limiterChan chan struct{}

// set the default number of concurrent jobs to run default to GOMAXPROCS
func SetMaxConcurrentJobs(n int) {
	if n < 1 {
		n = runtime.GOMAXPROCS(0)
	}
	limiterChan = make(chan struct{}, n)
}

func init() {
	SetMaxConcurrentJobs(runtime.GOMAXPROCS(0))
}

type ExecuteOptions struct {
	onJobsStart func(cps JobList)
	onJobStart  func(cps JobList, jobIndex int)
	onJobDone   func(cps JobList, jobIndex int)
	onJobsDone  func(cps JobList)
}

// effectively launch the child process, call on jobDone
// you should prepare child process before by calling either
// PrepareCmds, PrepareFns
// returns the number of errors encountered
// @todo add cancelation support
func execute(jobs JobList, opts ExecuteOptions) {
	if opts.onJobsStart != nil {
		opts.onJobsStart(jobs)
	}
	var wg sync.WaitGroup
	wg.Add(len(jobs))
	for i, child := range jobs {
		jobIndex := i
		limiterChan <- struct{}{}
		if opts.onJobStart != nil {
			opts.onJobStart(jobs, jobIndex)
		}
		go child.run(func() {
			defer func() { <-limiterChan }()
			defer wg.Done()
			if opts.onJobDone != nil {
				opts.onJobDone(jobs, jobIndex)
			}
		})
	}
	wg.Wait()
	close(limiterChan)
	if opts.onJobsDone != nil {
		opts.onJobsDone(jobs)
	}
}
