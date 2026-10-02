package jobmanager

import (
	"changeover/src/internal/protocol"
	"sync"
	"testing"
	"time"
)

// Helper para esperar a que un trabajo termine antes de validar
func waitForJobCompletion(jm *JobManager, jobID string, timeout time.Duration) (protocol.Job, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		job, err := jm.Status(jobID)
		if err == nil && (job.Estado == protocol.StateSucceeded || job.Estado == protocol.StateFailed || job.Estado == protocol.StateCanceled) {
			return job, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return jm.Status(jobID)
}

// TC-001: Envío válido
func TestTC001_SubmitValido(t *testing.T) {
	jm := NewManager()

	jobID := jm.Submit("echo", []string{"hello world"})
	if jobID == "" {
		t.Fatalf("TC-001: Se esperaba un JobID válido, pero se obtuvo un string vacío")
	}

	job, err := waitForJobCompletion(jm, jobID, 2*time.Second)
	if err != nil {
		t.Fatalf("TC-001: Error consultando el estado del trabajo: %v", err)
	}

	if job.Estado != protocol.StateSucceeded {
		t.Errorf("TC-001: Se esperaba estado SUCCEEDED, obtuve: %s", job.Estado)
	}
}

// TC-002: Concurrencia básica
func TestTC002_ConcurrenciaBasica(t *testing.T) {
	jm := NewManager()
	numJobs := 10
	var wg sync.WaitGroup
	jobIDs := make(chan string, numJobs)

	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := jm.Submit("sleep", []string{"0.1"})
			jobIDs <- id
		}()
	}

	wg.Wait()
	close(jobIDs)

	for id := range jobIDs {
		job, err := waitForJobCompletion(jm, id, 3*time.Second)
		if err != nil || job.Estado != protocol.StateSucceeded {
			t.Errorf("TC-002: El trabajo concurrente %s falló o no terminó. Estado: %s", id, job.Estado)
		}
	}
}

// TC-003: Estados y tiempos
func TestTC003_EstadosYTiempos(t *testing.T) {
	jm := NewManager()

	jobID := jm.Submit("sleep", []string{"0.3"})

	// Consultar inmediatamente para ver si pasó a QUEUED o RUNNING
	job, err := jm.Status(jobID)
	if err != nil {
		t.Fatalf("TC-003: Error al consultar estado inicial: %v", err)
	}

	if job.Estado != protocol.StateQueued && job.Estado != protocol.StateRunning {
		t.Errorf("TC-003: Estado inicial no esperado: %s", job.Estado)
	}

	// Esperar completación
	jobFinal, err := waitForJobCompletion(jm, jobID, 3*time.Second)
	if err != nil {
		t.Fatalf("TC-003: Error al esperar la finalización: %v", err)
	}

	if jobFinal.Estado != protocol.StateSucceeded {
		t.Errorf("TC-003: Se esperaba estado SUCCEEDED, obtuve: %s", jobFinal.Estado)
	}
}

// TC-004: Consultar y listar
func TestTC004_ConsultarYListar(t *testing.T) {
	jm := NewManager()

	id1 := jm.Submit("echo", []string{"job1"})
	id2 := jm.Submit("echo", []string{"job2"})

	waitForJobCompletion(jm, id1, 2*time.Second)
	waitForJobCompletion(jm, id2, 2*time.Second)

	// Validar List
	lista := jm.List()
	if len(lista) < 2 {
		t.Errorf("TC-004: Se esperaban al menos 2 trabajos en la lista, pero hay %d", len(lista))
	}

	// Validar Status de ID inexistente
	_, err := jm.Status("Job-Inexistente")
	if err == nil {
		t.Errorf("TC-004: Se esperaba error al consultar un Job ID que no existe")
	}
}

// TC-005: Cancelar
func TestTC005_Cancelar(t *testing.T) {
	jm := NewManager()

	// Proceso largo para dar tiempo a cancelarlo
	jobID := jm.Submit("sleep", []string{"5"})

	time.Sleep(100 * time.Millisecond) // Dejar que inicie ejecución

	err := jm.Cancel(jobID)
	if err != nil {
		t.Fatalf("TC-005: Error al cancelar el trabajo %s: %v", jobID, err)
	}

	job, _ := jm.Status(jobID)
	if job.Estado != protocol.StateCanceled {
		t.Errorf("TC-005: Se esperaba estado CANCELED, obtuve: %s", job.Estado)
	}
}

// TC-006: stdout/stderr y código de salida
func TestTC006_StdoutStderrAndExitCode(t *testing.T) {
	jm := NewManager()

	// 1. Probar captura de Stdout y ExitCode 0
	id1 := jm.Submit("echo", []string{"prueba_exitosa"})
	job1, _ := waitForJobCompletion(jm, id1, 2*time.Second)

	if job1.ExitCode != 0 {
		t.Errorf("TC-006: Se esperaba ExitCode 0, obtuve %d", job1.ExitCode)
	}
	if job1.Stdout == "" {
		t.Errorf("TC-006: Se esperaba contenido en Stdout, pero está vacío")
	}

	// 2. Probar error ( ExitCode != 0 y captura de Stderr/Error )
	id2 := jm.Submit("ls", []string{"/directorio_no_existente_xyz"})
	job2, _ := waitForJobCompletion(jm, id2, 2*time.Second)

	if job2.Estado != protocol.StateFailed {
		t.Errorf("TC-006: Se esperaba estado FAILED, obtuve %s", job2.Estado)
	}
	if job2.ExitCode == 0 {
		t.Errorf("TC-006: Se esperaba ExitCode distinto de 0 para comando fallido")
	}
}
