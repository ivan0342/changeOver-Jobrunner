# ADR-02 — Mecanismo de persistencia de metadatos

**Estado:** Aceptado
**Fecha:** 2026-10-01
**Responsable:** Iván Lozano Sánchez — Revisor: pendiente (Luis Francisco Rosas Vega, ausente en este periodo)

## Contexto y problema

RF-12 y RF-13 exigen que el sistema conserve metadatos y resultados después de
reiniciar el servicio, y que recupere coherentemente el historial al arrancar.
Sin embargo, la funcionalidad mínima exigida para Avance 1 (núcleo local) no
incluye persistencia ni recuperación tras reinicio — solo envío, ejecución,
consulta, listado, cancelación y manejo de errores en una sola sesión del
servicio. Es necesario decidir qué hacer mientras tanto, para no bloquear el
desarrollo del núcleo local con un problema que no corresponde a esta fase.

## Alternativas consideradas

### Opción A — Archivos planos con escritura atómica (write-then-rename)
- Cada trabajo (o el conjunto completo) se serializa a JSON y se escribe a un
  archivo temporal, que luego se renombra sobre el archivo final (`rename` es
  atómico en Linux), evitando dejar un archivo a medio escribir si el proceso
  se interrumpe (RNF-28).
- Simple de implementar, sin dependencias externas.
- Contra: consultas por ID requieren cargar y buscar en memoria tras leer el
  archivo completo; con cientos de trabajos (RNF-06: 500 registros) empieza a
  ser menos eficiente que una base de datos indexada.

### Opción B — SQLite
- Motor de base de datos embebido, sin proceso servidor aparte, con soporte de
  transacciones (ayuda directo con RNF-28: atomicidad).
- Consultas e índices por ID nativos, más eficiente a mayor volumen.
- Contra: agrega una dependencia (driver de SQLite para Go) y una capa de SQL
  que no es estrictamente necesaria para el volumen esperado del proyecto.

### Opción C — Sin persistencia (todo en memoria), pospuesto a Hito 2
- El mapa `storage` del `jobmanager` vive únicamente en RAM durante la
  ejecución del servicio.
- Permite enfocar el esfuerzo de Avance 1 en concurrencia, procesos y
  protocolo, que son los requisitos realmente exigidos en esta etapa.
- Contra: si el servicio se reinicia (caída, actualización, cierre manual),
  todo el historial de trabajos se pierde — RF-12/RF-13 quedan **sin cumplir
  hasta Hito 2**, de forma consciente y documentada.

## Decisión

Se elige la **Opción C** para Avance 1: **sin persistencia en disco**. El
`jobmanager` mantiene todo el estado únicamente en memoria (`map[string]*Job`
protegido por `sync.RWMutex`). La elección entre Opción A (archivos planos) y
Opción B (SQLite) se definirá formalmente en Hito 2, cuando RF-12/RF-13 sean
exigibles, considerando en ese momento el volumen real de trabajos y la
facilidad de depuración manual (los archivos planos son más fáciles de
inspeccionar a mano durante pruebas).

## Consecuencias

**Positivas:**
- Cero complejidad adicional en Avance 1; todo el esfuerzo se concentra en
  procesos, concurrencia y protocolo, que son los requisitos de esta fase.
- Sin dependencias externas nuevas que instalar o versionar.

**Negativas / riesgos:**
- Si el servicio se cae o se reinicia durante una demo o prueba, se pierde
  todo el historial — hay que tenerlo presente al preparar la demostración de
  la Technical Review 1 (no reiniciar el servicio a medio de una secuencia de
  comandos que se esté mostrando).
- RF-12 y RF-13 quedan explícitamente incumplidos hasta Hito 2; esto debe
  quedar reflejado como "Pendiente" en la matriz de trazabilidad, no como
  "FAIL" (porque no corresponde a esta entrega).

## Requisitos afectados

RF-12 (pospuesto), RF-13 (pospuesto), RNF-06, RNF-11, RNF-28 (a resolver en
Hito 2 según la opción final elegida).

## Evidencia / prototipo

Ninguno aún — la decisión de Hito 2 (Opción A vs. B) se validará con un
prototipo pequeño de escritura/lectura antes de integrarlo al jobmanager.
