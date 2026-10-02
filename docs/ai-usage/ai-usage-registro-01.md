# Registro de uso de IA — Entrada 01

**Herramienta utilizada:** Claude (Anthropic), vía chat asistido durante el
diseño y desarrollo del núcleo local (Avance 1).
**Periodo cubierto:** 18 de septiembre – 1 de octubre de 2026.
**Responsable del registro:** Iván Lozano Sánchez.

---

## Caso 1 — ACEPTADO: estructura de protocolo `Request`/`Response`

**Objetivo:** definir cómo estructurar los mensajes entre cliente y servicio,
evitando los problemas de mandar texto plano o arreglos de strings sin tipo.

**Resultado recibido:** la IA propuso dos structs de Go (`Request` con campos
`Type`, `Cmd`, `Args`, `JobID`; `Response` con `OK`, `Error`, `JobID`, `Job`,
`Jobs`), explicando que un campo `Type` explícito evita tener que "adivinar"
qué operación se pide a partir de la posición de las palabras en un mensaje.

**Revisión realizada:** se comparó contra la alternativa que ya se tenía
(arreglo plano de `[]string`), confirmando que el struct resolvía problemas
reales: argumentos con espacios, variedad de tipos de dato en la respuesta
(número para `ExitCode`, texto para `Stdout`/`Stderr`), y legibilidad del
código.

**Cambios aplicados:** se adoptó la estructura base casi sin cambios,
ajustando únicamente nombres de campos al español donde fue natural para el
equipo (`Comando`, `Argumentos` en el struct `Job`).

**Prueba agregada:** validado manualmente al correr el flujo completo
`submit` → `status` → `list` con JSON real viajando por el socket TCP.

**Aprendizaje:** una estructura de datos explícita (con campos con nombre)
es más fácil de extender con el tiempo que un formato posicional, incluso si
al inicio parece más trabajo escribirla.

---

## Caso 2 — MODIFICADO: condición de bloqueo (deadlock) en el worker

**Objetivo:** depurar por qué el servicio dejaba de responder (`list`,
`status`, nuevos `submit`) cuando había varios trabajos de larga duración
(`sleep 40`) corriendo al mismo tiempo.

**Resultado recibido:** al construir la lógica de cancelación junto con la
IA, el flujo sugerido inicialmente colocaba `j.mu.Lock()` envolviendo la
llamada a `cmd.Wait()` dentro del worker.

**Revisión realizada:** al probar con varios `sleep 40` simultáneos y un
`list` intercalado, el programa se quedó colgado (hubo que interrumpir con
Ctrl+C). Se investigó la causa junto con la IA, identificando que un mutex
que envuelve una operación de duración impredecible (como esperar a que
termine un proceso) bloquea innecesariamente a todas las demás goroutines
que necesitan ese mismo candado para operaciones rápidas (leer el mapa).

**Cambios aplicados:** se movió `cmd.Wait()` fuera de la sección protegida
por el mutex; el lock solo se vuelve a pedir justo antes y después de leer o
escribir el mapa `storage`, nunca durante la espera del proceso.

**Prueba agregada:** se repitió el escenario original (5+ `submit sleep 40`
seguidos de un `list`) y se confirmó que el servicio responde de inmediato
sin bloquearse, aunque los procesos de larga duración sigan corriendo de
fondo.

**Aprendizaje:** un mutex debe proteger la sección de código más corta
posible que toca memoria compartida; nunca debe envolver una llamada
bloqueante de duración impredecible (I/O, espera de procesos, red).

---

## Caso 3 — RECHAZADO/AJUSTADO: transporte Unix socket vs. TCP

**Objetivo:** elegir el mecanismo de comunicación entre el cliente CLI y el
servicio para la fase de núcleo local (ADR-01).

**Resultado recibido:** la IA recomendó un Unix domain socket
(`/tmp/changeover.sock`) para esta fase, argumentando que evita exponer un
puerto de red antes de que exista la capa de seguridad de Hito 3, y que
permite una migración directa a TCP más adelante gracias al paquete `net` de
Go.

**Revisión realizada:** el equipo decidió no seguir esa recomendación al
momento de escribir el código, implementando directamente un socket TCP en
`localhost:8080` desde el primer prototipo cliente-servidor.

**Cambios aplicados:** el ADR-01 fue actualizado para reflejar la decisión
real (TCP desde el inicio), documentando la razón: facilidad para probar con
herramientas estándar de red, y que el riesgo de seguridad es mínimo al estar
limitado a `localhost` durante esta fase, sin exponerse aún a una LAN/VPN
real (eso se decidirá con sus propias restricciones en Hito 3 / RF-20).

**Prueba agregada:** N/A — es una decisión de arquitectura, no de código,
verificada por el funcionamiento correcto del cliente-servidor sobre TCP en
todas las pruebas realizadas.

**Aprendizaje:** una recomendación de la IA no se adopta automáticamente;
el equipo evaluó el trade-off (simplicidad de prueba vs. pureza arquitectónica
de separar "local" de "remoto") y tomó una decisión propia, justificada y
documentada, tal como exige el Project Brief ("la frase 'lo recomendó la IA'
no constituye una justificación técnica").

---

## Caso 4 — ACEPTADO: casos de prueba TC-001 a TC-006

**Objetivo:** escribir pruebas automatizadas que cubrieran la funcionalidad
mínima de Avance 1 (envío, concurrencia, estados, consulta/listado,
cancelación, stdout/stderr/exit code), siguiendo el formato TC-XXX exigido
por el Plan de Verificación.

**Resultado recibido:** la IA propuso 6 funciones de prueba en Go
(`jobmanager_test.go`), usando un helper `waitForJobCompletion` con polling
para evitar falsos negativos por condiciones de carrera en las pruebas
mismas (en vez de usar `time.Sleep` fijo, que sería frágil).

**Revisión realizada:** se corrieron las pruebas con `go test -v` y también
con `go test -race`, confirmando que pasaban de forma consistente en varias
corridas, incluyendo el caso de concurrencia con 10 trabajos simultáneos
(TC-002).

**Cambios aplicados:** se adoptaron prácticamente sin cambios, ya que las
corridas confirmaron el comportamiento esperado contra la implementación
real del equipo.

**Prueba agregada:** las 6 pruebas en sí son la evidencia; se guardó su
salida en `verif/results/test_results.txt`.

**Aprendizaje:** un helper de espera con polling (en vez de un `sleep` fijo)
hace las pruebas de concurrencia más confiables y menos propensas a fallar
por timing en máquinas más lentas o más rápidas.

---

## Caso 5 — MODIFICADO: script de verificación `verify.sh`

**Objetivo:** contar con un script que compilara el sistema completo,
corriera las pruebas unitarias, y ejecutara un flujo funcional real
(servicio + cliente reales, no solo las funciones internas) de principio a
fin con un solo comando, cumpliendo RNF-20 y el entregable "script inicial
de verificación" de Avance 1.

**Resultado recibido:** una primera versión del script fue generada con otra
herramienta de IA (Gemini), asumiendo que el cliente aceptaba subcomandos
por línea de comandos (`changeoverclt run -- "comando"`, `changeoverclt
status <id>`).

**Revisión realizada:** al intentar correrlo, falló de inmediato — el
cliente real del equipo es **interactivo** (lee comandos de un loop con
`stdin`), no acepta argumentos así. Se identificó el desajuste comparando la
firma real de `main.go` del cliente contra lo que el script asumía.

**Cambios aplicados:** se reescribió el script para levantar el servicio en
background y mandarle los comandos al cliente real a través de un pipe
(`echo "submit ..." | ./changeoverclt`), simulando lo que un usuario
escribiría. Durante las pruebas, esto reveló un segundo bug real: el cliente
no manejaba el fin de entrada (EOF) del pipe y entraba en un loop infinito;
se corrigió agregando detección explícita de `io.EOF` en la lectura de
`stdin` del cliente.

**Prueba agregada:** se corrió el script completo con `timeout` como medida
de seguridad, confirmando que termina limpio con código de salida 0 y las 5
secciones (compilación, pruebas, `-race`, flujo funcional, cierre del
servicio) completas.

**Aprendizaje:** un script de verificación generado sin acceso al código real
del proyecto puede asumir una interfaz que no corresponde a la implementación
real; siempre hay que probarlo contra el sistema real antes de darlo por
válido, y la prueba misma puede revelar bugs adicionales no relacionados con
el script (como el manejo de EOF).

---

## Caso 6 — ACEPTADO: documentación (guía técnica, guía de usuario, ADRs)

**Objetivo:** producir la guía técnica (arquitectura, modelo de estados,
manejo de concurrencia), la guía de usuario, y los documentos ADR-02/ADR-04,
reflejando fielmente las decisiones y el comportamiento real del sistema ya
implementado.

**Resultado recibido:** la IA redactó los documentos basándose en el código
fuente real compartido en la conversación (no en una descripción genérica),
incluyendo diagramas de texto del flujo cliente-servidor y del modelo de
estados, y documentando explícitamente las limitaciones conocidas de esta
fase (sin persistencia, sin escalamiento de cancelación, etc.).

**Revisión realizada:** se contrastó cada afirmación técnica del documento
contra el comportamiento observado al correr el sistema (por ejemplo, el
diagrama de estados se verificó contra las transiciones reales vistas en las
pruebas manuales).

**Cambios aplicados:** se mantuvo la estructura propuesta; el equipo
verificó que las limitaciones declaradas coincidieran con la realidad del
código antes de aceptar el documento como definitivo.

**Prueba agregada:** N/A — es documentación, verificada por inspección
cruzada contra el código y las pruebas ya pasadas.

**Aprendizaje:** pedir que la documentación se genere a partir del código
real (no de una descripción abstracta) reduce el riesgo de que el documento
describa un sistema distinto al que realmente existe — error que sí ocurrió
con el script de verificación del Caso 5, generado sin ese contexto.
