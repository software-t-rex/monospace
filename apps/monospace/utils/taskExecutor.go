package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	exctr "github.com/software-t-rex/go-jobExecutor/v2"
	"github.com/software-t-rex/monospace/gomodules/colors"
)

func setInterleavedOutputDisplayNames(jobs exctr.JobList) {
	for i, job := range jobs {
		nameParts := strings.Split(job.Name(), "/")
		if !colors.ColorEnabled() {
			job.SetDisplayName(nameParts[len(nameParts)-1])
		} else {
			colorId := (i + 1) % 7
			color := "\033[39m"
			if colorId != 0 {
				color = fmt.Sprintf("\033[3%dm", colorId)
			}
			job.SetDisplayName(color + nameParts[len(nameParts)-1] + string(colors.Reset))
		}
	}
}

func NewTaskExecutor(outputMode string) *exctr.JobExecutor {
	e := exctr.NewExecutor()
	startTime := time.Now()
	var bold, success, failure, reset string
	successIndicator := Green("✔")
	failureIndicator := Red("✘")
	if colors.ColorEnabled() {
		bold = string(colors.Bold)
		success = string(colors.BrightGreen)
		failure = string(colors.BrightRed)
		reset = string(colors.Reset)
	}
	e.OnJobsStart(func(jobs exctr.JobList) {
		fmt.Printf(bold+"Starting %d tasks...\n"+reset, len(jobs))
	})
	switch outputMode { //grouped,interleaved,status-only,errors-only,none
	case "none": // do nothing
	case "interleaved":
		e.OnJobsStart(setInterleavedOutputDisplayNames)
		e.WithInterleavedOutput()
	case "errors-only":
		e.OnJobsStart(setInterleavedOutputDisplayNames)
		e.OnJobsStart(func(jobs exctr.JobList) {
			setInterleavedOutputDisplayNames(jobs)
			for _, job := range jobs {
				job.Cmd.Stderr = exctr.NewPrefixedWriter(os.Stdout, job.Name()+": ")
			}
		})
	case "status-only":
		statusFunc := func(jobs exctr.JobList) string {
			out := make([]string, len(jobs))
			for i, j := range jobs {
				status := "⏳"
				if j.IsState(exctr.JobStateRunning) {
					status = "🏃"
				} else if j.IsState(exctr.JobStateFailed) {
					status = failureIndicator
				} else if j.IsState(exctr.JobStateSucceed) {
					status = successIndicator
				}
				out[i] = fmt.Sprintf("%s %s", status, j.Name())
			}
			return strings.Join(out, "\n")
		}
		printSummary := func(jobs exctr.JobList) {
			fmt.Println(statusFunc(jobs))
		}
		printProgress := func(jobs exctr.JobList, jobId int) {
			esc := fmt.Sprintf("\033[%dA\033[J", len(jobs)) // clean sequence
			fmt.Println(esc + statusFunc(jobs))
		}
		e.OnJobsStart(printSummary)
		e.OnJobDone(printProgress)
		e.OnJobStart(printProgress)
	default: // grouped is the default
		e.OnJobDone(func(jobs exctr.JobList, jobId int) {
			job := jobs[jobId]
			indicator := failureIndicator
			verb := "failed"
			if job.IsState(exctr.JobStateSucceed) {
				verb = "succeed"
				indicator = successIndicator
			}
			fmt.Printf("%s %s %s in %v\n%s", indicator, bold+job.Name()+reset, verb, job.Duration, Indent(job.Res, "  "))
		})
	}
	e.OnJobsDone(func(jobs exctr.JobList) {
		var succeed int
		var failed int
		allGreenIndicator := failureIndicator
		elapsed := time.Since(startTime)
		for _, job := range jobs {
			if job.IsState(exctr.JobStateSucceed) {
				succeed++
			} else {
				failed++
			}
		}
		if succeed == len(jobs) {
			allGreenIndicator = successIndicator
		}
		fmt.Print(bold + "Tasks: " + allGreenIndicator + " " + bold)
		if succeed > 0 {
			fmt.Printf(success+"%d succeed"+reset+bold+" / ", succeed)
		}
		if failed > 0 {
			fmt.Printf(failure+"%d failed"+reset+bold+" / ", failed)
		}
		fmt.Printf("%d total"+reset+"\n", len(jobs))
		fmt.Printf(bold+"total time: %v\n"+reset, elapsed)
	})
	return e
}
