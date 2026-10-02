# changeOver — Implementación de JobRunner

> Nombre confirmado por el equipo: **changeOver**.

## Propósito

changeOver es una plataforma ligera para registrar, ejecutar, supervisar y controlar
trabajos del sistema operativo en equipos Linux, desarrollada como parte del
proyecto de Programación de Sistemas Avanzados (2026B). Permite enviar comandos
como trabajos, ejecutarlos en segundo plano con concurrencia controlada,
consultar su estado, recuperar su salida y solicitar su cancelación, con
soporte para operación local y acceso remoto restringido a red privada/VPN
(remoto: pendiente a Hito 3).

## Integrantes

| Nombre | Correo institucional | Rol |
|---|---|---|
| Iván Lozano Sánchez | ivan.lsanchez03@alumnos.udg.mx | Encargado del proyecto |
| Luis Francisco Rosas Vega | francisco.rosas3783@alumnos.udg.mx | Integrante |

## Lenguaje y stack

- **Lenguaje:** Go 1.22 (ver `docs/decisions/ADR-01-arquitectura-base.md`)
- **Transporte:** TCP sobre `localhost:8080` (ver ADR-01)
- **Protocolo:** JSON estructurado (`Request`/`Response`) vía `encoding/json`
- **Persistencia:** en memoria por ahora, pospuesta a Hito 2 (ver `docs/decisions/ADR-02-persistencia.md`)
- **Cancelación:** terminación directa (SIGKILL), escalamiento pospuesto a Hito 2 (ver `docs/decisions/ADR-04-cancelacion.md`)

## Construcción y ejecución

Requiere Go 1.22 o superior instalado (`go version` para verificar).

```bash
git clone <url-del-repo>
cd changeOver-Jobrunner
go build -o changeoverd ./src/cmd/changeoverd
go build -o changeoverclt ./src/cmd/changeoverclt
```

**Terminal 1 — iniciar el servicio:**
```bash
./changeoverd
```

**Terminal 2 — usar el cliente:**
```bash
./changeoverclt
```

```
> submit echo hola mundo
 [RESPUESTA SERVIDOR] Status: OK | JobID: Job-1
> status Job-1
 JobID: Job-1 | Estado: SUCCEEDED | ExitCode: 0
--- stdout ---
hola mundo
> list
> cancel Job-1
```

Ver `docs/user-guide/user-guide.md` para la guía completa con todos los
comandos y ejemplos.

## Correr las pruebas automatizadas

```bash
go test -v ./src/internal/jobmanager/...
```

Para verificación adicional de condiciones de carrera (recomendado):
```bash
go test -race -v ./src/internal/jobmanager/...
```

## Estructura mínima de carpetas

Estructura oficial según "JobRunner — Estándares de Repositorio e Ingeniería":

```
changeOver-Jobrunner/
├── README.md
├── go.mod
├── src/
│   ├── cmd/
│   │   ├── changeoverd/               # servicio (daemon)
│   │   └── changeoverclt/             # cliente CLI
│   └── internal/
│       ├── jobmanager/                # ciclo de vida de trabajos
│       └── protocol/                  # formato de mensajes
├── docs/
│   ├── user-guide/                    # instalación y operación
│   ├── technical-guide/               # arquitectura, protocolo, persistencia, decisiones
│   ├── decisions/                     # ADRs (mínimo 5, ver sección de ADR)
│   ├── ai-usage/                      # registro de uso de IA (aceptado/modificado/rechazado)
│   ├── change-requests/               # análisis de impacto de cambios del cliente
│   └── incidents/                     # síntomas, hipótesis, causa, corrección, regresión
├── verif/
│   ├── verification-plan/             # plan y matriz de trazabilidad (traceability-matrix.md)
│   ├── test-cases/                    # casos TC-XXX
│   ├── scripts/                       # automatización de pruebas
│   ├── test-data/                     # datos de entrada controlados
│   └── results/                       # evidencia por ejecución
├── project-management/                # planificación, roles, minutas, acuerdos, evidencia
└── .github/
    └── ISSUE_TEMPLATE/                # plantillas de trabajo y defectos
```



## Estado del proyecto

**Fase actual:** Avance 1 — Núcleo local ejecutable (entrega 2 de octubre).

Funcionalidad completa: envío de trabajos, ID único, ejecución como proceso
separado, consulta de estado, listado, cancelación, captura de stdout/stderr,
código de salida, y manejo de comandos inválidos sin afectar el servicio.
Verificado con 6 casos de prueba automatizados (TC-001 a TC-006) y el
detector de condiciones de carrera de Go (`-race`).

Pendiente para Hito 2: persistencia en disco, límite de concurrencia
configurable, escalamiento de cancelación, rechazo explícito por saturación
de cola.

## Documentación relacionada

- `docs/decisions/` — Architecture Decision Records (mínimo 5 requeridos; 3 cerrados hasta ahora: ADR-01, ADR-02, ADR-04).
- `docs/user-guide/` — cómo instalar y usar el sistema.
- `docs/technical-guide/` — arquitectura, modelo de estados, manejo de procesos.
- `docs/ai-usage/` — registro de uso de herramientas de IA (aceptado, modificado, rechazado).
- `verif/verification-plan/traceability-matrix.md` — matriz de trazabilidad RF/RNF.
- `project-management/planificacion.md` — planificación, roles, cronograma, minutas y evidencia.
