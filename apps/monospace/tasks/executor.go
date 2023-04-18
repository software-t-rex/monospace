package tasks

import (
	"fmt"
	"os"
	"strings"
	"time"

	exctr "github.com/software-t-rex/go-jobExecutor/v2"
	"github.com/software-t-rex/monospace/gomodules/colors"
	"github.com/software-t-rex/monospace/utils"
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

func NewExecutor(outputMode string) *exctr.JobExecutor {
	e := exctr.NewExecutor()
	startTime := time.Now()
	var bold, success, failure, reset string
	successIndicator := utils.Green("✔")
	failureIndicator := utils.Red("✘")
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
	case "errors-only":
		fallthrough
	case "interleaved":
		withStdout := true
		if outputMode == "errors-only" {
			withStdout = false
		}
		e.OnJobsStart(func(jobs exctr.JobList) {
			setInterleavedOutputDisplayNames(jobs)
			for _, job := range jobs {
				pw := exctr.NewPrefixedWriter(os.Stdout, job.Name()+": ")
				if job.Cmd != nil {
					if withStdout {
						job.Cmd.Stdout = pw
					}
					job.Cmd.Stderr = pw
				} else if job.Fn != nil {
					fn := job.Fn
					job.Fn = func() (string, error) {
						res, err := fn()
						if withStdout && res != "" {
							pw.Write([]byte(res))
						}
						if err != nil {
							pw.Write([]byte(utils.ErrorStyle(err.Error())))
						}
						return res, err
					}
				}
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
			statusLine := fmt.Sprintf("%s %s %s in %v\n", indicator, bold+job.Name()+reset, verb, job.Duration)
			err := ""
			res := ""
			if job.Err != nil {
				err = utils.Indent(utils.ErrorStyle(job.Err.Error()), "  ")
			}
			if job.Res != "" {
				res = utils.Indent(job.Res, "  ")
			}
			fmt.Print(statusLine, err, res)
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
