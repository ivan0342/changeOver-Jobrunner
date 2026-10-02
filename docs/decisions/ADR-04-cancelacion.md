# ADR-04 — Política de cancelación y escalamiento

**Estado:** Aceptado
**Fecha:** 2026-10-01
**Responsable:** Iván Lozano Sánchez — Revisor: pendiente (Luis Francisco Rosas Vega, ausente en este periodo)

## Contexto y problema

RF-10 exige que el usuario pueda solicitar la cancelación de un trabajo en
cola o en ejecución. RF-30 (fuera del alcance obligatorio de Avance 1) pide
además una política de escalamiento si el trabajo no termina tras la
cancelación normal. Se necesita decidir el mecanismo concreto de cancelación
para el núcleo local, sabiendo que el escalamiento completo no se implementará
todavía.

## Alternativas consideradas

### Opción A — Terminación directa (SIGKILL vía `Process.Kill()`)
- Simple: una sola llamada (`proc.Kill()`), sin esperar ni verificar si el
  proceso respondió.
- Efecto inmediato y garantizado a nivel de sistema operativo — SIGKILL no
  puede ser ignorado ni capturado por el proceso hijo.
- Contra: no le da oportunidad al proceso de limpiar recursos propios (cerrar
  archivos, liberar locks internos, etc.) antes de morir.

### Opción B — Escalamiento gradual (SIGTERM → espera con timeout → SIGKILL)
- Más "educado": manda SIGTERM primero, para que el proceso termine por su
  cuenta si maneja esa señal; si no responde en N segundos, se manda SIGKILL.
- Cumple directamente el espíritu de RF-30.
- Contra: requiere lógica adicional (temporizador, verificación de si el
  proceso ya terminó) que no es indispensable para la funcionalidad mínima de
  Avance 1, y agrega una variable más (el timeout) que requeriría su propia
  justificación y pruebas.

## Decisión

Se elige la **Opción A** (terminación directa vía `Kill()`) para Avance 1. Es
suficiente para cumplir RF-10 con una interfaz clara (el usuario cancela, el
proceso muere, el estado pasa a `CANCELED`). El escalamiento gradual (Opción
B, RF-30 completo) se diseñará e implementará en Hito 2, una vez que el
núcleo local esté validado y haya tiempo para probar los casos de borde del
temporizador (qué pasa si el proceso ignora SIGTERM pero sí es afectado por
SIGKILL, cuánto tiempo esperar, etc.).

Para un trabajo que todavía está en cola (estado `QUEUED`, sin proceso del
sistema operativo asociado todavía), la cancelación es aún más simple: se
marca el estado como `CANCELED` directamente en el mapa, y el worker que
eventualmente lo saque de la cola verifica ese estado antes de ejecutar nada,
evitando así arrancar un proceso que ya fue cancelado.

## Consecuencias

**Positivas:**
- Implementación simple y rápida de verificar (un `Kill()` es determinístico:
  o el proceso muere, o `Kill()` regresa un error claro).
- Cubre el caso de uso principal de Avance 1 sin necesitar lógica de
  temporización todavía.

**Negativas / riesgos:**
- Un proceso cancelado no tiene oportunidad de hacer limpieza propia (por
  ejemplo, si escribiera un archivo temporal, podría quedar a medio escribir)
  — este riesgo es aceptable para Avance 1 porque los comandos de prueba
  usados (`sleep`, `echo`, etc.) no mantienen estado externo relevante.
- RF-30 (escalamiento) queda explícitamente pendiente para Hito 2; debe
  marcarse como "Pendiente", no "FAIL", en la matriz de trazabilidad.

## Requisitos afectados

RF-10 (cumplido), RF-26 (cumplido parcialmente: una sola cancelación a la vez
sobre el mismo ID da un resultado coherente gracias al mutex), RF-30
(pospuesto a Hito 2).

## Evidencia / prototipo

Prueba manual: `submit sleep 40` seguido de `cancel <id>` mientras está
`RUNNING` — el proceso termina de inmediato y el estado pasa a `CANCELED`
(ver TC-005 en los casos de prueba de Avance 1).
