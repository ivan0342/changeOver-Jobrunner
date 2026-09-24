package jobmanager

import (
	"changeover/src/internal/protocol"
	"fmt"
	"sync"
	"time"
)

type JobManager struct {
	jobs    chan protocol.Job
	storage map[string]*protocol.Job
	mu      sync.RWMutex
	id      int
}

func NewManager() *JobManager {
	var numWorkers int
	numWorkers = 4
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
		j.mu.Unlock()

		// Simulamos la ejecución del comando
		time.Sleep(3 * time.Second)

		fmt.Printf("[Worker %d] Completó %s\n", workerId, job.ID)
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

func (j *JobManager) DefFunc(tipo string, trabajo protocol.Job) {
	switch tipo {
	case "submit":
		{
			fmt.Println("seleccionaste el tipo submit")
			j.submit(trabajo.Comando, trabajo.Argumentos)
		}
	}
}
