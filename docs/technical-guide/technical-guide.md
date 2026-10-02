# Guía técnica — changeOver

## Arquitectura

changeOver sigue una arquitectura cliente-servicio:

- **`changeoverd`** (servicio): proceso persistente que escucha conexiones
  TCP en `localhost:8080`, mantiene el estado de todos los trabajos en
  memoria, y ejecuta cada trabajo como un proceso hijo separado del sistema
  operativo.
- **`changeoverclt`** (cliente): programa interactivo de línea de comandos
  que se conecta al servicio, envía solicitudes y muestra las respuestas.

```
┌─────────────────┐         TCP (localhost:8080)       ┌──────────────────┐
│  changeoverclt   │ ───── JSON (Request/Response) ───▶ │   changeoverd     │
│  (interactivo)   │ ◀──────────────────────────────── │  (servicio)        │
└─────────────────┘                                     │                    │
                                                          │  ┌──────────────┐  │
                                                          │  │  JobManager   │  │
                                                          │  │  (worker pool)│  │
                                                          │  └──────┬───────┘  │
                                                          │         │          │
                                                          │   os/exec.Command  │
                                                          │         │          │
                                                          │   proceso hijo     │
                                                          │   (echo, sleep,    │
                                                          │    etc.)           │
                                                          └──────────────────┘
```

## Modelo de procesos e IPC

La comunicación cliente-servicio usa un socket **TCP** (decisión documentada
en ADR-01). Los mensajes se serializan como **JSON**, uno por conexión activa,
usando `json.Encoder`/`json.Decoder` de la librería estándar de Go — este
mecanismo resuelve el framing de mensajes automáticamente (RF-21): el decoder
sabe reconocer dónde termina un objeto JSON completo dentro del stream de
bytes, sin necesitar un delimitador manual.

La ejecución de trabajos usa un **worker pool**: al arrancar, el servicio
lanza un número fijo de goroutines ("workers") que consumen trabajos de un
channel compartido (`jobs chan protocol.Job`, con buffer de 100). Cada worker,
al tomar un trabajo, lo ejecuta usando `os/exec.Command(...).Start()` +
`.Wait()` — el equivalente de Go a `fork()` + `execve()` + `waitpid()` en C.

## Modelo de estados

```
   QUEUED ──────────────▶ RUNNING ──────────────▶ SUCCEEDED
      │                      │                    (exit code 0)
      │                      │
      │                      └──────────────────▶ FAILED
      │                                           (exit code ≠ 0,
      │                                            o el comando
      │                                            no pudo arrancar)
      │
      └──────────────────────────────────────────▶ CANCELED
         (desde QUEUED: el worker lo detecta       (desde RUNNING:
          antes de ejecutar y lo salta)              Kill() al proceso)
```

- **QUEUED**: el trabajo fue aceptado, tiene ID, pero ningún worker lo ha
  tomado todavía del channel.
- **RUNNING**: un worker lo tomó y el proceso del sistema operativo ya
  arrancó (`cmd.Start()` tuvo éxito).
- **SUCCEEDED**: el proceso terminó con código de salida 0.
- **FAILED**: el proceso terminó con código de salida distinto de 0, o nunca
  pudo arrancar (ej. comando inexistente).
- **CANCELED**: se solicitó cancelación. Si el trabajo estaba en `QUEUED`, el
  worker lo detecta y nunca lo ejecuta. Si estaba en `RUNNING`, se le manda
  `Kill()` (SIGKILL) al proceso del sistema operativo.

## Manejo de concurrencia

Todo el estado compartido (`map[string]*protocol.Job`, el mapa de procesos
activos) está protegido por un `sync.RWMutex`:

- Operaciones de **lectura** (`Status`, `List`) usan `RLock()`/`RUnlock()`,
  permitiendo múltiples lecturas simultáneas.
- Operaciones de **escritura** (el worker actualizando el estado de un
  trabajo, `Cancel`) usan `Lock()`/`Unlock()` exclusivo.

**Regla de diseño aplicada:** el mutex nunca envuelve una operación de
duración impredecible (como `cmd.Wait()`, que puede tardar segundos o
minutos) — solo protege el acceso directo al mapa, liberándose de inmediato
después. Esto se corrigió explícitamente tras detectar un deadlock real
durante pruebas manuales (ver `docs/ai-usage/`).

## Manejo de errores y señales

- Un comando que no existe o no se puede ejecutar (`cmd.Start()` falla) se
  marca como `FAILED` con el mensaje de error guardado, **sin** detener el
  servicio ni afectar otros trabajos (RNF-08, RNF-09).
- La cancelación de un trabajo en ejecución usa `Process.Kill()`, que envía
  `SIGKILL` al proceso hijo — terminación inmediata y no capturable por el
  proceso (ver ADR-04 para la justificación de esta decisión).

## Verificación de concurrencia

El paquete `jobmanager` se prueba con `go test -race`, el detector de
condiciones de carrera integrado de Go. Esto permitió encontrar y corregir un
bug real: `Status()` liberaba el lock de lectura (`RUnlock()`) antes de leer
los campos del trabajo, dejando esa lectura sin protección frente a
escrituras concurrentes del worker. El fix consistió en extender el alcance
del lock (`defer j.mu.RUnlock()`) para cubrir toda la copia de campos.

## Limitaciones conocidas (Avance 1)

- Sin persistencia a disco (ver ADR-02).
- Número de workers fijo en código (3), no configurable todavía (RF-16,
  pendiente a Hito 2).
- Cancelación sin escalamiento gradual (ver ADR-04).
- Sin operación remota (el servicio solo acepta conexiones en `localhost`;
  RF-18 a RF-22 son alcance de Hito 3).
