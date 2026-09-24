package protocol

type Job struct {
	ID         string
	Comando    string
	Argumentos []string
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
}
