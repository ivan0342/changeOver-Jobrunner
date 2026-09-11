# Planificación inicial — changeOver (JobRunner)

## 1. Matriz de roles

| Área | Responsable principal | Revisor |
|---|---|---|
| Producto (alcance, requisitos, decisiones funcionales) | Iván Lozano Sánchez | Luis Francisco Rosas Vega |
| Verificación (pruebas, matriz de trazabilidad, evidencia) | Luis Francisco Rosas Vega | Iván Lozano Sánchez |
| Ingeniería (arquitectura, implementación, ADR) | Iván Lozano Sánchez | Luis Francisco Rosas Vega |


## 2. Cronograma con fechas reales

Fechas confirmadas:

| Periodo | Corte del curso | Hito(s) técnico(s) a cubrir | Foco principal |
|---|---|---|---|
| 10–16 sep | — | Avance 0 | Repo, README, roles, cronograma, Issues, ADR planteados |
| 17–23 sep | — | Hito 0 | Arquitectura preliminar, RF/RNF baselined, plan de verificación en borrador |
| 24 sep–1 oct | **Avance 1 (2 oct)** | Hito 1 | Envío local, proceso hijo, estados, stdout/stderr, cola inicial, pruebas unitarias, RT-1 |
| 2–12 oct | — | Hito 2 (inicio) | Límite de concurrencia, cola estable, cancelación, persistencia |
| 13–23 oct | — | Hito 2 (cierre) | Recuperación tras reinicio, bitácora, RT-2, incidente de concurrencia/persistencia registrado |
| 24–30 oct | **Avance 2 (30 oct)** | Hito 3 | Protocolo de red, cliente remoto, restricción LAN/VPN, RT-3, Change Request del cliente |
| 31 oct–17 nov | — | Hito 4 | Documentación completa, matriz sin huecos, verificación cruzada entre equipos |
| 18 nov–1 dic | **Entrega final (1 dic)** | Hito 5 | Versión etiquetada, demostración final, defensa individual, incidente no anunciado |

## 3. Riesgos y dependencias conocidas

| Riesgo/dependencia | Impacto | Mitigación inicial |
|---|---|---|
| Equipo de solo 2 personas frente a alcance amplio (RF-01 a RF-30, RNF-01 a RNF-34) | Alto | Priorizar RF/RNF obligatorios sobre extensiones opcionales; repartir Producto/Verificación/Ingeniería con revisión cruzada |
| Poca experiencia previa con concurrencia en Go (si se confirma como lenguaje) | Medio | Dedicar la Semana 1–2 a un prototipo pequeño de concurrencia antes del Hito 1 |
| Cambio de alcance solicitado por el cliente (Change Request obligatorio en Hito 3) | Medio | Reservar tiempo de buffer en semanas 10–12 para análisis de impacto |
| Dependencia de acceso a una red LAN/VPN real para la demostración del Hito 3 | Medio | Confirmar con anticipación el entorno de prueba (dos equipos en la misma red) |

## 4. Issues a crear en el repositorio

1. **[Arranque] Configurar repositorio y estructura de carpetas** — checklist de la entrega Avance 0, incluyendo `.github/ISSUE_TEMPLATE/` con plantillas de historia/requisito, defecto, tarea técnica y deuda/mejora (mínimo exigido por los estándares).
2. **[Avance 1] Preparar Hito 0 — Inicio y línea base** — arquitectura preliminar, RF/RNF baselined, plan de verificación en borrador.
3. **[ADR] ADR-01 — Elección de lenguaje y modelo de concurrencia**
4. **[ADR] ADR-02 — Mecanismo de persistencia de metadatos**
5. **[ADR] ADR-03 — Protocolo de comunicación remota (framing y versión)**
6. **[Verificación] Definir plan de pruebas y estructura de la matriz de trazabilidad** — crear `verif/verification-plan/traceability-matrix.md` con las columnas obligatorias (Req ID, Descripción, Prioridad, Método, Caso(s), Evidencia, Resultado, Defecto/Excepción) y todas las filas RF-01…RF-30 y RNF-01…RNF-34 en estado "Pendiente".
7. **[Documentación] Esqueleto de guía de usuario y guía técnica** — crear `docs/user-guide/` y `docs/technical-guide/` con la lista mínima de contenidos técnicos (arquitectura, IPC, estados, protocolo, formato persistente, señales, errores, configuración, amenazas y límites).

### Issues adicionales que exigen los estándares (abrir también en Avance 0)

8. **[ADR] ADR-04 — Política de cancelación y escalamiento** — el documento de estándares exige un mínimo de **5 ADR** en `docs/decisions/`, no 3.
9. **[ADR] ADR-05 — Tratamiento de solicitudes duplicadas y semántica de reintentos**
10. **[Registro IA] Crear `docs/ai-usage/` con la primera entrada** — debe incluir, desde ahora, un ejemplo de resultado de IA aceptado, uno modificado y uno rechazado (con objetivo, resultado recibido, revisión realizada, cambios aplicados, prueba agregada y aprendizaje).
11. **[Higiene] Configurar `.gitignore` y verificar ausencia de secretos/credenciales** antes del primer push.

## 5. Primeras decisiones a resolver mediante ADR

### ADR-01 — Lenguaje y modelo de concurrencia
- **Alternativas a comparar:** Go (goroutines/channels) vs. C (hilos/procesos + señales) vs. Rust (async/threads).
- **Afecta a:** RF-04, RF-05, RNF-04, RNF-07, RNF-09.

### ADR-02 — Persistencia de metadatos y resultados
- **Alternativas a comparar:** archivos planos con escritura atómica (write-then-rename) vs. SQLite.
- **Afecta a:** RF-12, RF-13, RNF-06, RNF-11, RNF-28.

### ADR-03 — Protocolo de comunicación remota
- **Alternativas a comparar:** protocolo binario propio con framing por longitud vs. JSON delimitado por línea (newline-delimited JSON) sobre TCP.
- **Afecta a:** RF-18, RF-21, RF-22, RNF-13, RNF-24, RNF-32.

### ADR-04 — Política de cancelación y escalamiento
- **Alternativas a comparar:** señal de terminación simple (SIGTERM) con timeout fijo vs. escalamiento gradual (SIGTERM → espera → SIGKILL) con política configurable.
- **Afecta a:** RF-10, RF-26, RF-30.

### ADR-05 — Tratamiento de solicitudes duplicadas
- **Alternativas a comparar:** idempotencia por clave de solicitud (request ID provisto por el cliente) vs. detección por huella de contenido (hash del comando + timestamp) en el servidor.
- **Afecta a:** RF-27, RNF-27.
