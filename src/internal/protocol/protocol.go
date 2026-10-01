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
	ExitCode int    // corregido: era string, ahora es el número real
	Stdout   string // nuevo: para regresar el stdout por separado
	Stderr   string // nuevo: para regresar el stderr por separado

	Jobs []Job
}

const (
	StateQueued    = "QUEUED"
	StateRunning   = "RUNNING"
	StateSucceeded = "SUCCEEDED"
	StateFailed    = "FAILED"
	StateCanceled  = "CANCELED"
)
