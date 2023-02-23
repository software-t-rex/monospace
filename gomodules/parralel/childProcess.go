/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/
package parallel

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	JobStatePending = 0
	JobStateRunning = 1
	JobStateDone    = 2
	JobStateSucceed = 4
	JobStateFailed  = 8
)

type runnableFn func() (string, error)
type JobList []*job
type job struct {
	Cmd       *exec.Cmd
	Fn        runnableFn
	Res       string
	Err       error
	status    int
	StartTime time.Time
	Duration  time.Duration
	mutex     sync.RWMutex
}

func (cp *job) run(done func()) {
	defer func() {
		defer done()
		cp.mutex.Lock()
		cp.status = cp.status ^ JobStateRunning | JobStateDone
		cp.mutex.Unlock()
	}()
	cp.mutex.Lock()
	cp.StartTime = time.Now()
	cp.status = cp.status | JobStateRunning
	cp.mutex.Unlock()
	if cp.Cmd != nil {
		res, err := cp.Cmd.CombinedOutput()
		cp.mutex.Lock()
		cp.Res = string(res)
		cp.Err = err
	} else if cp.Fn != nil {
		res, err := cp.Fn()
		cp.mutex.Lock()
		cp.Res = res
		cp.Err = err
	}
	if cp.Err != nil {
		cp.status = cp.status | JobStateFailed
	} else {
		cp.status = cp.status | JobStateSucceed
	}
	cp.Duration = time.Since(cp.StartTime)
	cp.mutex.Unlock()
}

func (cp *job) Name() string {
	if cp != nil && cp.Cmd != nil {
		return strings.Join(cp.Cmd.Args, " ")
	} else if cp != nil && cp.Fn != nil {
		return runtime.FuncForPC(reflect.ValueOf(cp.Fn).Pointer()).Name()
	}
	return "a job"
}

func (cp *job) IsState(jobState int) bool {
	cp.mutex.RLock()
	res := cp.status&jobState != 0
	cp.mutex.RUnlock()
	return res
}

// helper method for jobs execTemplate
func tplExec(tplName string, subject interface{}) string {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, tplName, " is not defined, see parallel.setTemplate", r)
		}
	}()
	var out bytes.Buffer
	err := parallelTemplate.ExecuteTemplate(&out, tplName, subject)
	if err != nil {
		return err.Error()
	}
	return out.String()
}

func (cp *job) execTemplate(tplName string) string {
	return tplExec(tplName, cp)
}
func (cps *JobList) execTemplate(tplName string) string {
	return tplExec(tplName, cps)
}
