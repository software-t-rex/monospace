
{{/* helper for job status */}}
{{define "jobStateIndicator"}}{{
	if .IsState 1
}}🏃 running {{
	else if .IsState 4
}}👍 Success {{
	else if .IsState 8
}}💥 Error   {{
	else
}}⏳ pending {{end}}{{end}}

{{/* return a single job status line*/}}
{{define "jobStatusLine"}}
{{- template "jobStateIndicator" .}} {{.Name}}
{{end}}

{{/* return a single job status + outputs */}}
{{define "jobStatusFull"}}
{{- template "jobStateIndicator" .}} {{.Name}}:{{if .Err}} {{.Err}}{{end}}
{{if .Res}}{{.Res| trim | indent 2}}{{end}}
{{end}}

{{define "jobsStart"}}
{{- .Name}} added to queue
{{end}}

{{define "jobStart"}}starting {{- template "command"}}{{end}}

{{/* render summary for JobList */}}
{{define "startSummary"}}Starting {{len .}} job{{if gt (len .) 1}}s{{end}}:
{{range .}}  - {{.Name}}
{{end}}{{end}}

{{/* render Start for JobList */}}
{{define "startProgressReport"}}Starting {{len .}} job{{if gt (len .) 1}}s{{end}}:
{{range .}}{{template "jobStatusLine" .}}{{end -}}
{{end}}

{{/* render ordered Status for JobList */}}
{{define "progressReport"}}
{{- range .}}{{template "jobStatusLine" .}}{{end -}}
{{end}}

{{/* render ordered Status for JobList */}}
{{define "doneReport"}}{{len .}} job{{if gt (len .) 1}}s{{end}} terminated:
{{range .}}{{template "jobStatusFull" . }}{{end -}}
{{end}}