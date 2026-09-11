# changeOver — Implementación de JobRunner

> Nombre propuesto: **changeOver*.

## Propósito

changeOver es una plataforma ligera para registrar, ejecutar, supervisar y controlar
trabajos del sistema operativo en equipos Linux, desarrollada como parte del
proyecto de Programación de Sistemas Avanzados (2026B). Permite enviar comandos
como trabajos, ejecutarlos en segundo plano con concurrencia controlada,
consultar su estado, recuperar su salida y solicitar su cancelación, con
soporte para operación local y acceso remoto restringido a red privada/VPN.

## Integrantes

Nombre | Correo institucional | Rol |

Iván Lozano Sánchez - ivan.lsanchez03@alumnos.udg.mx - Encargado del proyecto
Luis Francisco Rosas Vega -francisco.rosas3783@alumnos.udg.mx - Integrante 

## Lenguaje y stack

- **Lenguaje:** Go *(propuesto — ver ADR-01)*
- **Persistencia:** por definir *(ver ADR-02)*
- **Protocolo de red:** por definir *(ver ADR-03)*

## Construcción provisional

```bash
# Requiere Go instalado (ver https://go.dev/dl/)
git clone <url-del-repo>
cd anden
go build ./src/...
```

*(Este bloque se actualizará conforme exista código funcional.)*

## Estructura mínima de carpetas

Estructura oficial según "JobRunner — Estándares de Repositorio e Ingeniería":

```
anden/
├── README.md
├── src/                              # código de producción
├── docs/
│   ├── user-guide/                   # instalación y operación
│   ├── technical-guide/              # arquitectura, protocolo, persistencia, decisiones
│   ├── decisions/                    # ADRs (mínimo 5, ver sección de ADR)
│   ├── ai-usage/                     # registro de uso de IA (aceptado/modificado/rechazado)
│   ├── change-requests/              # análisis de impacto de cambios del cliente
│   └── incidents/                    # síntomas, hipótesis, causa, corrección, regresión
├── verif/
│   ├── verification-plan/            # plan y matriz de trazabilidad (traceability-matrix.md)
│   ├── test-cases/                   # casos TC-XXX
│   ├── scripts/                      # automatización de pruebas
│   ├── test-data/                    # datos de entrada controlados
│   └── results/                      # evidencia por ejecución (verif/results/<run-id>/)
├── project-management/               # planificación, roles, minutas, acuerdos, evidencia
└── .github/
    └── ISSUE_TEMPLATE/                # plantillas de trabajo y defectos
```

## Estado del proyecto

**Fase actual:** Avance 0 — Organización del equipo y preparación del repositorio.
Sin código funcional aún. Ver Issues abiertos para el plan de arranque.

## Documentación relacionada

- `docs/decisions/` — Architecture Decision Records (mínimo 5 requeridos).
- `docs/user-guide/` / `docs/technical-guide/` — se completarán a partir del Hito 1.
- `docs/ai-usage/` — registro de uso de herramientas de IA (aceptado, modificado, rechazado).
- `verif/verification-plan/traceability-matrix.md` — matriz de trazabilidad RF/RNF.
- `project-management/` — planificación, roles, cronograma, minutas y evidencia (ver `planificacion.md`).
