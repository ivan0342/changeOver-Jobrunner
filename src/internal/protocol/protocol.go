package protocol

type Job struct {
	ID         string
	Comando    string
	Argumentos []string
	Estado     string
	Stdout     string
	Stderr     string
	ErrorMsg   string
	ExitCode   int
}

type Request struct {
	Type string
	Cmd  string
	Args []string
}

type Response struct {
	Ok    bool
	Error string
	JobID string

	Status   string
	ExitCode string
	Storage  map[string]*Job
}

const (
	StateQueued    = "QUEUED"
	StateRunning   = "RUNNING"
	StateSucceeded = "SUCCEEDED"
	StateFailed    = "FAILED"
)
