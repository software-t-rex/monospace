package utils

import (
	"fmt"
	"time"

	"github.com/software-t-rex/go-jobExecutor/v2"
	"github.com/software-t-rex/monospace/gomodules/colors"
)

func NewTaskExecutor() *jobExecutor.JobExecutor {
	e := jobExecutor.NewExecutor()
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
	e.OnJobsStart(func(jobs jobExecutor.JobList) {
		fmt.Printf(bold+"Starting %d tasks...\n"+reset, len(jobs))
	})
	e.OnJobDone(func(jobs jobExecutor.JobList, jobId int) {
		job := jobs[jobId]
		indicator := failureIndicator
		verb := "failed"
		if job.IsState(jobExecutor.JobStateSucceed) {
			verb = "succeed"
			indicator = successIndicator
		}
		fmt.Printf("%s %s %s in %v\n%s", indicator, job.Name(), verb, job.Duration, Indent(job.Res, "  "))
	})
	e.OnJobsDone(func(jobs jobExecutor.JobList) {
		var succeed int
		var failed int
		allGreenIndicator := failureIndicator
		elapsed := time.Since(startTime)
		for _, job := range jobs {
			if job.IsState(jobExecutor.JobStateSucceed) {
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
