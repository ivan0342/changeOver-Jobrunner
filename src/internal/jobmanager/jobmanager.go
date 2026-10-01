package jobmanager

import (
	"bytes"
	"changeover/src/internal/protocol"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

type JobManager struct {
	jobs    chan protocol.Job
	storage map[string]*protocol.Job
	procs   map[string]*os.Process
	mu      sync.RWMutex
	id      int
}

func NewManager() *JobManager {
	var numWorkers int
	numWorkers = 3
	jb := &JobManager{
		jobs:    make(chan protocol.Job, 100),
		storage: make(map[string]*protocol.Job),
		procs:   make(map[string]*os.Process),
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

		actual, existe := j.storage[job.ID]
		if existe && actual.Estado == protocol.StateCanceled {
			j.mu.Unlock()
			continue // no lo ejecutamos, ya fue cancelado en cola
		}
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
			fmt.Printf("[%s] Error al ejecutar: %v\n", job.ID, err)
			fmt.Printf("Detalle del error: %s\n", stderr.String())
			continue
		}

		j.mu.Lock()
		j.procs[job.ID] = cmd.Process
		j.mu.Unlock()

		err = cmd.Wait()

		j.mu.Lock()
		defer_ := j.storage[job.ID] // solo para claridad, uso directo abajo

		if defer_.Estado == protocol.StateCanceled {
			// Cancel() ya lo marcó mientras corría; solo completamos su salida.
			defer_.Stdout = stdout.String()
			defer_.Stderr = stderr.String()
			delete(j.procs, job.ID)
			j.mu.Unlock()
			continue
		}
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
		defer_.Estado = estado
		defer_.ExitCode = exitCode
		defer_.ErrorMsg = errorMsg
		defer_.Stdout = stdout.String()
		defer_.Stderr = stderr.String()
		delete(j.procs, job.ID)
		j.mu.Unlock()

		fmt.Printf("[Worker %d] Completó %s con estado %s (Exit Code: %d), y el output es %s\n", workerId, job.ID, estado, exitCode, job.Stdout)

	}
}

func (j *JobManager) Submit(comando string, argumentos []string) string {

	j.mu.Lock()
	j.id++
	jobID := fmt.Sprintf("Job-%d", j.id)

	nuevoJob := protocol.Job{
		ID:         jobID,
		Comando:    comando,
		Argumentos: argumentos,
		Estado:     protocol.StateQueued,
	}
	j.storage[jobID] = &nuevoJob
	j.mu.Unlock()

	j.jobs <- nuevoJob

	return jobID
}

func (j *JobManager) Status(jobId string) (protocol.Job, error) {

	j.mu.RLock()
	job, ok := j.storage[jobId]
	j.mu.RUnlock()

	if !ok {
		return protocol.Job{}, fmt.Errorf("no existe un trabajo con ID %q", jobId)
	}

	// Copiamos los campos a un valor nuevo (no el puntero interno), para
	// que quien lo reciba no comparta memoria con el mapa protegido.
	return protocol.Job{
		ID:       job.ID,
		Comando:  job.Comando,
		Estado:   job.Estado,
		ExitCode: job.ExitCode,
		Stdout:   job.Stdout,
		Stderr:   job.Stderr,
		ErrorMsg: job.ErrorMsg,
	}, nil
}

func (j *JobManager) List() []protocol.Job {
	j.mu.RLock()
	defer j.mu.RUnlock()

	jobs := make([]protocol.Job, 0, len(j.storage))
	for _, job := range j.storage {
		jobs = append(jobs, protocol.Job{
			ID:       job.ID,
			Comando:  job.Comando,
			Estado:   job.Estado,
			ExitCode: job.ExitCode,
			Stdout:   job.Stdout,
			Stderr:   job.Stderr,
			ErrorMsg: job.ErrorMsg,
		})
	}
	return jobs
}

func (j *JobManager) Cancel(jobId string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	job, ok := j.storage[jobId]
	if !ok {
		return fmt.Errorf("no existe un trabajo con ID %q", jobId)
	}

	switch job.Estado {
	case protocol.StateQueued:
		// Aún no arranca: solo lo marcamos. El worker lo va a ver cancelado
		// cuando le toque su turno y no lo va a ejecutar (ver paso 3).
		job.Estado = protocol.StateCanceled
		return nil

	case protocol.StateRunning:
		proc, existeProc := j.procs[jobId]
		if !existeProc || proc == nil {
			return fmt.Errorf("no se encontró el proceso en ejecución para %q", jobId)
		}
		if err := proc.Kill(); err != nil {
			return fmt.Errorf("no se pudo cancelar el trabajo: %w", err)
		}
		job.Estado = protocol.StateCanceled
		return nil

	default:
		return fmt.Errorf("el trabajo %q ya terminó (estado %s), no se puede cancelar", jobId, job.Estado)
	}
}

func (j *JobManager) DefFunc(tipo string, trabajo protocol.Job, jobId string) protocol.Response {
	switch tipo {
	case "submit":
		jobID := j.Submit(trabajo.Comando, trabajo.Argumentos)
		return protocol.Response{Ok: true, JobID: jobID}

	case "status":
		job, err := j.Status(jobId)
		if err != nil {
			return protocol.Response{Ok: false, Error: err.Error()}
		}
		return protocol.Response{
			Ok:       true,
			JobID:    job.ID,
			Status:   job.Estado,
			ExitCode: job.ExitCode,
			Stdout:   job.Stdout,
			Stderr:   job.Stderr,
		}

	case "list":
		return protocol.Response{Ok: true, Jobs: j.List()}

	case "cancel":
		if err := j.Cancel(jobId); err != nil {
			return protocol.Response{Ok: false, Error: err.Error()}
		}
		return protocol.Response{Ok: true}
	default:
		return protocol.Response{Ok: false, Error: "tipo desconocido: " + tipo}
	}
}
