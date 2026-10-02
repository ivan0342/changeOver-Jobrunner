# ADR-01 — Lenguaje, modelo de concurrencia y arquitectura base

**Estado:** Aceptado
**Fecha:** 2026-09-18
**Responsable:** Iván Lozano Sánchez (Producto/Ingeniería) — Revisor: Luis Francisco Rosas Vega

## Contexto y problema

changeOver (implementación de JobRunner) debe registrar, ejecutar, supervisar y
controlar trabajos del sistema operativo en Linux, con concurrencia controlada,
manejo de procesos y señales, y (a partir del Hito 3) acceso remoto restringido
a LAN/VPN. Se necesita elegir el lenguaje de implementación y el mecanismo base
de comunicación entre el cliente CLI y el servicio, antes de poder iniciar
cualquier desarrollo del núcleo local (Avance 1).

## Alternativas consideradas

### Opción A — Go (goroutines/channels, `os/exec`, `net`)
- Concurrencia nativa vía goroutines y channels, sin necesidad de un pool de
  hilos manual.
- `os/exec` da control directo de procesos hijos, señales y captura separada
  de stdout/stderr.
- El paquete `net` cubre tanto sockets Unix (fase local) como TCP (fase remota,
  Hito 3) con la misma API, facilitando la migración futura.
- Compila a un binario estático — build reproducible desde clon limpio sin
  gestionar entorno de ejecución (RNF-02).
- Tipado estático: reduce errores de tipo en tiempo de compilación frente a un
  lenguaje dinámico.
- Contra: recolector de basura implica menos control fino de memoria que C/Rust.

### Opción B — C (hilos POSIX o `fork`/`exec`, sockets BSD)
- Control más fino sobre procesos y señales a nivel de sistema operativo.
- Sin recolector de basura: control total de memoria.
- Contra: mayor riesgo de errores de memoria y condiciones de carrera difíciles
  de depurar con un equipo de 2 personas y tiempo limitado (mes y medio).
  Requiere escribir manualmente el pool de concurrencia, el framing del
  protocolo y el manejo de buffers.

### Opción C — Python (`subprocess`, `asyncio`, sockets)
- Desarrollo más rápido inicialmente, sintaxis simple.
- `subprocess` maneja procesos hijos e IPC de forma razonable.
- Contra: tipado dinámico incrementa el riesgo de errores no detectados hasta
  ejecución, justamente en las áreas más sensibles del proyecto (transiciones
  de estado concurrentes, RNF-27). Requiere gestionar intérprete y dependencias
  en el entorno de destino (venv/requirements.txt), lo cual complica RNF-02
  frente a un binario único.

## Decisión

Se elige **Go** como lenguaje de implementación, con la siguiente arquitectura
base para el Avance 1 (núcleo local):

- **Transporte:** Unix domain socket (ej. `/tmp/changeover.sock`) para la
  comunicación cliente-servicio en esta fase local. Se elige sobre TCP local
  porque evita exponer un puerto de red antes de que exista la capa de
  seguridad de Hito 3, y porque la migración a TCP (Hito 3) es directa gracias
  a que el paquete `net` de Go comparte la misma interfaz para ambos tipos de
  socket.
- **Protocolo:** JSON delimitado por línea (un objeto JSON por línea,
  *newline-delimited JSON*) sobre el socket. Se elige sobre un protocolo
  binario propio por simplicidad de depuración en esta fase temprana; se
  revisará en ADR-03 si conviene mantenerlo o migrar a un framing binario para
  la fase remota.
- **Modelo de procesos:** una goroutine por trabajo en ejecución, sincronizada
  mediante channels para reportar cambios de estado al mapa central de
  trabajos.

## Consecuencias

**Positivas:**
- Permite empezar el desarrollo del núcleo local de inmediato con herramientas
  estándar de la librería de Go, sin dependencias externas.
- El mismo diseño de transporte (paquete `net`) se reutiliza al pasar de socket
  Unix a TCP en Hito 3, minimizando retrabajo.
- Build reproducible con un solo comando (`go build`), cumpliendo RNF-02 sin
  configuración adicional en la máquina de un tercero.

**Negativas / riesgos:**
- El equipo tiene que aprender o reforzar el modelo de concurrencia de Go
  (goroutines/channels/`context`) si no tiene experiencia previa; se mitiga
  dedicando los primeros 2 días del Avance 1 a un prototipo mínimo antes de
  construir el núcleo completo.
- El protocolo JSON por línea es menos eficiente que un formato binario; se
  acepta el costo por ahora porque el volumen de trabajos esperado (RNF-06:
  500 registros) no lo hace un cuello de botella real.

## Requisitos afectados

RF-01, RF-02, RF-04, RF-06, RF-07, RF-11, RF-18 (base para Hito 3), RNF-01,
RNF-02, RNF-03, RNF-04, RNF-07, RNF-08, RNF-09.

## Evidencia / prototipo

Prototipo mínimo: servicio Go que escucha en un Unix socket, acepta un mensaje
JSON `{"cmd": "...", "args": [...]}`, lanza el proceso con `os/exec`, y
responde con un ID generado y el estado inicial `QUEUED`. Evidencia a guardar
en `verif/results/adr-01-prototipo/` una vez ejecutado.
