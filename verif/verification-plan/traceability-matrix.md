# Matriz de Trazabilidad de Requisitos — changeOver

Formato según "JobRunner — Plan de Verificación, Validación y Aceptación".
Esta matriz cubre los requisitos exigibles en Avance 1 (núcleo local). El
resto de RF/RNF (operación remota, persistencia, escalamiento, etc.) queda
marcado como "Pendiente" — se completará en Hito 2 y Hito 3.

| Req ID | Descripción breve | Prioridad | Método | Caso(s) | Evidencia | Resultado | Defecto/Excepción |
|---|---|---|---|---|---|---|---|
| RF-01 | Enviar trabajo y recibir ID único | Alta | Prueba | TC-001 | `go test -v`, verif/results/test_results.txt | PASS | — |
| RF-02 | Validar entradas vacías/mal formadas | Alta | Prueba | TC-001 | cliente rechaza `submit` sin comando (ver user-guide) | PASS | — |
| RF-03 | Mantener cola cuando no hay capacidad inmediata | Alta | Prueba | TC-002 | channel con buffer (jobs chan, cap 100) | PASS (parcial, sin rechazo explícito aún) | Rechazo explícito al saturar cola es RF-25, pendiente a Hito 2 |
| RF-04 | Ejecutar en procesos separados sin bloquear | Alta | Prueba | TC-002 | `go test -race -v`, 10 trabajos concurrentes | PASS | — |
| RF-05 | Límite de concurrencia configurable | Alta | Prueba | — | worker pool fijo (3 workers) | Pendiente | Límite hoy está fijo en código, no es configurable (RF-16); a resolver en Hito 2 |
| RF-06 | Estados QUEUED/RUNNING/SUCCEEDED/FAILED/CANCELED | Alta | Prueba | TC-003, TC-005 | `go test -v` | PASS | — |
| RF-07 | Registrar tiempos y código de salida | Alta | Prueba | TC-006 | `go test -v` | PASS (código de salida); tiempos de inicio/fin no implementados aún | Falta registrar ReceivedAt/StartedAt/EndedAt explícitamente |
| RF-08 | Consultar estado por ID | Alta | Prueba | TC-004 | `go test -v` | PASS | — |
| RF-09 | Listar trabajos con filtros básicos | Media | Prueba | TC-004 | `go test -v` | PASS (parcial) | Listado sin filtro por estado todavía; filtros pendientes |
| RF-10 | Cancelación de trabajo en cola o ejecución | Alta | Prueba | TC-005 | `go test -v` | PASS | — |
| RF-11 | Capturar stdout/stderr sin mezclar | Alta | Prueba | TC-006 | `go test -v` | PASS | — |
| RF-12 | Persistencia tras reinicio | Media | — | — | — | Pendiente | Pospuesto a Hito 2, ver ADR-02 |
| RF-13 | Recuperación de historial al arrancar | Media | — | — | — | Pendiente | Pospuesto a Hito 2, ver ADR-02 |
| RF-15 | Inicio/cierre controlado del servicio | Media | — | — | — | Pendiente | No implementado aún (sin manejo de SIGTERM en el servicio) |
| RF-16 | Configuración de directorio, concurrencia, puerto | Media | — | — | — | Pendiente | Puerto fijo (8080) y workers fijos (3) en código, no configurables aún |
| RF-17 | Ayuda de uso y códigos de salida del cliente | Baja | Inspección | — | docs/user-guide/user-guide.md | PASS (parcial) | Cliente interactivo no implementa `--help`; mensajes de uso sí existen |
| RF-18 a RF-22 | Operación remota (protocolo, framing, desconexión) | Alta | — | — | — | Pendiente | Fuera de alcance de Avance 1 (Hito 3) |
| RF-26 | Cancelaciones concurrentes sobre mismo ID | Alta | Análisis | — | protegido por mutex en `Cancel()` | PASS (parcial) | No hay prueba automatizada específica del caso doble-cancel aún |
| RF-29 | Terminación inesperada de proceso hijo | Alta | — | — | — | Pendiente | No probado aún con un proceso que termine por señal externa |
| RNF-01 | Compila y corre en Linux declarado | Alta | Prueba | — | `go build ./...` en WSL2/Ubuntu | PASS | — |
| RNF-02 | Construcción reproducible desde clon limpio | Alta | Prueba | — | `go build ./...`, `go.mod` fijado a Go 1.22 | PASS | Corregido: go.mod pedía Go 1.27.1, causaba fallo de build sin red; bajado a 1.22 |
| RNF-04 | Al menos 3 trabajos simultáneos con límite=3 | Alta | Prueba | TC-002 | `go test -race -v`, 10 jobs con 3 workers | PASS | — |
| RNF-08 | Solicitud inválida no termina el servicio | Crítica | Prueba | TC-006 | `go test -v`, job con comando inexistente | PASS | — |
| RNF-09 | Terminación anormal de un trabajo no afecta a otros | Crítica | Prueba | TC-002, TC-006 | `go test -v` | PASS | — |
| RNF-19 | Compila sin advertencias nuevas | Media | Inspección | — | `go vet ./...` | PASS | — |
| RNF-20 | Pruebas automatizadas con un único comando | Media | Inspección | — | `go test ./...` | PASS | — |
| RNF-27 | Transiciones de estado consistentes ante concurrencia | Crítica | Prueba | TC-002 | `go test -race -v` | PASS (tras corrección) | Se detectó y corrigió un data race real en `Status()` (RUnlock antes de leer campos); ver docs/ai-usage/ |

## Pendientes explícitos para Hito 2 (no aplican a Avance 1)

RF-14, RF-23, RF-24, RF-25, RF-27, RF-28, RF-30, RNF-03, RNF-05, RNF-06,
RNF-07, RNF-10 a RNF-18, RNF-21 a RNF-26, RNF-28 a RNF-34 quedan como
"Pendiente" — corresponden a persistencia, administración, calidad de red y
recuperación avanzada, fuera del alcance mínimo de Avance 1.
