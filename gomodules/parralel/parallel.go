/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/
package parallel

import (
	"os"
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

func parrallelExecutor(cps []*childProcess, onJobDone func(cp *childProcess), onJobsDone func(cps []*childProcess)) {
	nbChilds := len(cps)
	var wg sync.WaitGroup
	wg.Add(nbChilds)
	for _, child := range cps {
		cp := child
		cp.onDone = func() {
			<-limiterChan
			wg.Done()
			onJobDone(cp)
		}
		limiterChan <- struct{}{}
		go cp.run()
	}
	wg.Wait()
	close(limiterChan)
	onJobsDone(cps)
}

// execute given list of commands in parallel
// print combined outputs on stdout as they arrive
// You can change the output format by using the SetJobDoneTemplate function
func Exec(cmdsAndArgs ...[]string) []error {
	cps := cmdListtoChildProcess(cmdsAndArgs)
	errs := make([]error, len(cps))
	// keep values of template setted at launch time
	tpl := jobDoneTemplate
	onJobDone := func(cp *childProcess) {
		cp.execTemplate(tpl, os.Stdout)
	}
	onJobsDone := func(cps []*childProcess) {
		for i, cp := range cps {
			errs[i] = cp.err
		}
	}
	parrallelExecutor(cps, onJobDone, onJobsDone)
	return errs
}

// execute given list of commands in parallel
// returns
// - ordered list of combined outputs from child process
// - orderd list of errors from child process
func ExecRes(cmdsAndArgs ...[]string) ([]string, []error) {
	nbCommands := len(cmdsAndArgs)
	cps := cmdListtoChildProcess(cmdsAndArgs)
	res := make([]string, nbCommands)
	err := make([]error, nbCommands)
	onJobsDone := func(cps []*childProcess) {
		for i, cp := range cps {
			res[i] = string(cp.res)
			err[i] = cp.err
		}
	}
	parrallelExecutor(cps, func(cp *childProcess) {}, onJobsDone)
	return res, err
}
