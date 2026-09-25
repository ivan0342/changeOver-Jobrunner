package jobmanager

import (
	"bytes"
	"changeover/src/internal/protocol"
	"fmt"
	"os/exec"
	"sync"
)

type JobManager struct {
	jobs    chan protocol.Job
	storage map[string]*protocol.Job
	mu      sync.RWMutex
	id      int
}

func NewManager() *JobManager {
	var numWorkers int
	numWorkers = 5
	jb := &JobManager{
		jobs:    make(chan protocol.Job, 100),
		storage: make(map[string]*protocol.Job),
	}
	var i int
	for i = 1; i <= numWorkers; i++ {
		go jb.worker(i)
	}

	return jb
}

func (j *JobManager) worker(workerId int) {

	for job := range j.jobs {
		fmt.Printf("[Worker %d] Tomó el trabajo %s de la cola\n", workerId, job.ID)

		// Actualizamos estado en el mapa
		j.mu.Lock()
		j.storage[job.ID] = &job
		j.storage[job.ID].Estado = protocol.StateRunning
		j.mu.Unlock()

		//Ejecución del comando
		cmd := exec.Command(job.Comando, job.Argumentos...)

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		fmt.Printf("[Worker %d] Completó %s\n", workerId, job.ID)

		err := cmd.Start()
		if err != nil {
			// Si el comando falló (ej: el comando no existe o retornó un código de error)
			j.storage[job.ID].ErrorMsg = err.Error()
			j.storage[job.ID].Estado = protocol.StateFailed
			j.storage[job.ID].ExitCode = -1
			fmt.Printf("❌ [%s] Error al ejecutar: %v\n", job.ID, err)
			fmt.Printf("📋 Detalle del error: %s\n", stderr.String())
			continue
		}

		err = cmd.Wait()

		var exitCode int = 0
		var estado string = "SUCCEEDED"
		var errorMsg string = ""

		if err != nil {
			estado = "FAILED"
			errorMsg = err.Error()

			// Investigando exec.ExitError para obtener el ExitCode real (ej: 1, 127, etc)
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				// Si el error no es de salida (ej: problemas del sistema), asignamos -1
				exitCode = -1
			}
		}

		// --- TODO 6: Actualizar el storage con los resultados finales ---
		j.mu.Lock()
		j.storage[job.ID].Estado = estado
		j.storage[job.ID].ExitCode = exitCode
		j.storage[job.ID].ErrorMsg = errorMsg
		j.storage[job.ID].Stdout = stdout.String() // Convierte el buffer de bytes a string
		j.storage[job.ID].Stderr = stderr.String()
		j.mu.Unlock()

		fmt.Printf("[Worker %d] Completó %s con estado %s (Exit Code: %d), y el output es %s\n", workerId, job.ID, estado, exitCode, job.Stdout)

	}
}

func (j *JobManager) submit(comando string, argumentos []string) string {

	j.mu.Lock()
	j.id++
	jobID := fmt.Sprintf("Job-%d", j.id)
	j.mu.Unlock()

	nuevoJob := protocol.Job{
		ID:         jobID,
		Comando:    comando,
		Argumentos: argumentos,
	}

	j.jobs <- nuevoJob

	return jobID
}

func (j *JobManager) status(JobId string) []string {

	var exitCode string
	var status = j.storage[JobId].Estado
	if j.storage[JobId].Stdout == "" {
		exitCode = j.storage[JobId].Stderr
	} else {
		exitCode = j.storage[JobId].Stdout
	}
	if status == protocol.StateRunning {
		return []string{JobId, status, ""}
	} else {
		return []string{JobId, status, exitCode}
	}
}

func (j *JobManager) DefFunc(tipo string, trabajo protocol.Job) protocol.Response {
	switch tipo {
	case "submit":
		{
			fmt.Println("seleccionaste el tipo submit")
			jobId := j.submit(trabajo.Comando, trabajo.Argumentos)
			return protocol.Response{Ok: true, JobID: jobId}
		}
	case "status":
		{
			fmt.Println("Sleccionaste el tipo estatus")
			status := j.status(trabajo.Comando)
			return protocol.Response{Ok: true, JobID: status[0], Status: status[1], ExitCode: status[2]}

		}
	case "list":
		{
			fmt.Println("Seleccionaste el tipo lista")
			return protocol.Response{Ok: true, Storage: j.storage}
		}

	default:
		return protocol.Response{Ok: false, Error: "tipo desconocido: " + tipo}
	}
}
