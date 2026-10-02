#!/bin/bash
# verif/scripts/verify.sh
#
# Script de verificación del núcleo local de changeOver (Avance 1).
# Corre la compilación, las pruebas automatizadas con detección de
# condiciones de carrera, y un flujo funcional manual contra el servicio
# real levantado en background.
set -e

echo "=== [1/5] Compilando el servicio y el cliente ==="
go build -o bin/changeoverd ./src/cmd/changeoverd
go build -o bin/changeoverclt ./src/cmd/changeoverclt

echo "=== [2/5] Ejecutando pruebas unitarias de JobManager (TC-001 a TC-006) ==="
go test -v ./src/internal/jobmanager/...

echo "=== [3/5] Verificando ausencia de condiciones de carrera (-race) ==="
go test -race ./src/internal/jobmanager/...

echo "=== [4/5] Levantando el servicio para la prueba de flujo funcional ==="
# El cliente real (changeoverclt) es INTERACTIVO: lee comandos de stdin en
# un loop, no acepta argumentos de línea de comandos. Por eso aquí le
# mandamos los comandos por un pipe, uno por línea, simulando lo que
# escribiría un usuario real.
./bin/changeoverd > /tmp/changeoverd.log 2>&1 &
DAEMON_PID=$!
sleep 1  # dar tiempo al servicio para que empiece a escuchar

echo "--- Enviando flujo: submit, status, list, submit+cancel, comando inválido ---"
{
  echo "submit echo hola mundo"
  sleep 0.3
  echo "submit sleep 5"
  sleep 0.3
  echo "status Job-1"
  echo "list"
  echo "cancel Job-2"
  sleep 0.3
  echo "status Job-2"
  echo "submit comando_que_no_existe_xyz"
  sleep 0.3
  echo "status Job-3"
  echo "list"
} | ./bin/changeoverclt

echo "=== [5/5] Deteniendo el servicio ==="
kill "$DAEMON_PID" 2>/dev/null || true
wait "$DAEMON_PID" 2>/dev/null || true

echo "=== Verificación completada con éxito ==="
echo "Log del servicio guardado en /tmp/changeoverd.log"
