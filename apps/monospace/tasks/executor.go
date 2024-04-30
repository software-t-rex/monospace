package tasks

import (
	"fmt"
	"os"
	"strings"
	"time"

	exctr "github.com/software-t-rex/go-jobExecutor/v2"
	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
)

func setInterleavedOutputDisplayNames(jobs exctr.JobList) {
	for i, job := range jobs {
		nameParts := strings.Split(job.Name(), "/")
		if !ui.EnhancedEnabled() {
			job.SetDisplayName(nameParts[len(nameParts)-1])
		} else {
			colorId := fmt.Sprintf("%d", (i+1)%7)
			if colorId == "0" {
				job.SetDisplayName(nameParts[len(nameParts)-1])
			} else {
				job.SetDisplayName(ui.ApplyStyle(nameParts[len(nameParts)-1], ui.Color(colorId).Foreground()))
			}
		}
	}
}

func NewExecutor(outputMode string) *exctr.JobExecutor {
	e := exctr.NewExecutor()
	startTime := time.Now()
	theme := ui.GetTheme()
	successIndicator := theme.Success("✔")
	failureIndicator := theme.Error("✘")
	e.OnJobsStart(func(jobs exctr.JobList) {
		fmt.Printf(theme.Bold("Starting %d tasks...\n"), len(jobs))
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
							pw.Write([]byte(theme.Error(err.Error())))
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
			statusLine := fmt.Sprintf("%s %s %s in %v\n", indicator, theme.Bold(job.Name()), verb, job.Duration)
			err := ""
			res := ""
			if job.Err != nil {
				err = utils.Indent(theme.Error(job.Err.Error()), "  ")
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
		sb := strings.Builder{}
		if ui.EnhancedEnabled() {
			sb.WriteString(ui.SGREscapeSequence(ui.Bold))
		}
		sb.WriteString("Tasks: ")
		sb.WriteString(allGreenIndicator)
		sb.WriteString(" ")
		if succeed > 0 {
			sb.WriteString(theme.Success(fmt.Sprintf("%d succeed", succeed)))
			sb.WriteString(" / ")
		}
		if failed > 0 {
			sb.WriteString(theme.Error(fmt.Sprintf("%d failed", failed)))
			sb.WriteString(" / ")
		}
		sb.WriteString(fmt.Sprintf("%d total", len(jobs)))
		if ui.EnhancedEnabled() {
			sb.WriteString(ui.SGRResetSequence())
		}
		sb.WriteString("\n")
		sb.WriteString(theme.Bold(fmt.Sprintf("total time: %v\n", elapsed)))
		fmt.Print(sb.String())
	})
	return e
}
