# Guía de usuario — changeOver

Esta guía permite instalar, iniciar, enviar y consultar un trabajo sin
asistencia del equipo de desarrollo (RNF-23).

## Requisitos previos

- Sistema operativo Linux (probado en Ubuntu, vía WSL2).
- Go 1.22 o superior instalado (`go version` para verificar).

## Instalación

```bash
git clone <url-del-repositorio>
cd changeOver-Jobrunner
go build -o changeoverd ./src/cmd/changeoverd
go build -o changeoverctl ./src/cmd/changeoverctl
```

Esto genera dos binarios en la carpeta actual: `changeoverd` (el servicio) y
`changeoverctl` (el cliente).

## Iniciar el servicio

En una terminal, desde la raíz del proyecto:

```bash
./changeoverd
```

Deberías ver:
```
Servidor escuchando en el puerto 8080
```

Deja esta terminal abierta — el servicio debe seguir corriendo mientras uses
el cliente. Para detenerlo, usa `Ctrl+C`.

## Usar el cliente

En **otra** terminal (sin cerrar la del servicio), desde la raíz del
proyecto:

```bash
./changeoverctl
```

Vas a ver el mensaje de conexión exitosa y un indicador `>` esperando
comandos.

### Enviar un trabajo

```
> submit echo hola mundo
 [RESPUESTA SERVIDOR] Status: OK | JobID: Job-1
```

El sistema regresa un identificador único (`Job-1`), que se usa para
consultarlo o cancelarlo después.

### Consultar el estado de un trabajo

```
> status Job-1
 JobID: Job-1 | Estado: SUCCEEDED | ExitCode: 0
--- stdout ---
hola mundo
```

Si el trabajo aún no termina, el estado aparece como `RUNNING` o `QUEUED`;
puedes volver a consultar las veces que quieras.

### Listar todos los trabajos

```
> list
Job-1	SUCCEEDED	stdout="hola mundo\n"
Job-2	RUNNING	stdout=""
```

### Cancelar un trabajo

```
> submit sleep 40
 [RESPUESTA SERVIDOR] Status: OK | JobID: Job-3

> cancel Job-3
 [RESPUESTA SERVIDOR] Trabajo cancelado.

> status Job-3
 JobID: Job-3 | Estado: CANCELED | ExitCode: 0
```

### Comando inválido (manejo de errores)

```
> submit comando_que_no_existe_xyz
 [RESPUESTA SERVIDOR] Status: OK | JobID: Job-4

> status Job-4
 JobID: Job-4 | Estado: FAILED | ExitCode: -1
```

El servicio no se detiene ni afecta a otros trabajos cuando esto ocurre.

## Códigos de salida del cliente (RF-17)

- `0`: operación completada exitosamente.
- `1`: error reportado por el servicio (ej. ID de trabajo inexistente).
- `2`: uso incorrecto del comando (ej. falta un argumento requerido).

## Limitaciones conocidas en esta versión (Avance 1)

- No hay operación remota todavía: el cliente y el servicio deben correr en
  la misma máquina (`localhost:8080`).
- No hay persistencia: si el servicio se reinicia, se pierde el historial de
  trabajos (ver ADR-02).
- La cancelación es inmediata (sin periodo de gracia): ver ADR-04.
- Los argumentos con espacios dentro de comillas (ej. `echo "hola mundo"`
  como un solo argumento) no se separan correctamente todavía; usa palabras
  sueltas por ahora (`echo hola mundo` se trata como 3 argumentos separados).
