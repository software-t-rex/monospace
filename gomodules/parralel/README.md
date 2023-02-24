# Parallel

go module to assist in running jobs in multiple goroutines and print output

## Sample usage:
```go
import "monospace/parralel"

func main () {
	executor = parallel.NewExecutor().WithProgressOutput()
	executor.AddJobFns(
		func() (string, error) {
			// do stuff here
			return "Success", nil
		},
	)
	executor.AddJobCmd("ls", "-l")

	jobErrors := executor.Execute()
}

```

