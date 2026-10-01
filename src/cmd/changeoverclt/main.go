package main

import (
	"bufio"
	"changeover/src/internal/protocol"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conexion, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println(" [CLIENTE] No se pudo conectar al servidor:", err)
		return
	}
	defer conexion.Close()
	fmt.Println(" [CLIENTE] ¡Conectado exitosamente!")

	encoder := json.NewEncoder(conexion)
	decoder := json.NewDecoder(conexion)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		// 1. Leer la entrada del usuario
		mensaje, _ := reader.ReadString('\n')
		mensaje = strings.TrimSpace(mensaje)

		// 2. Separar las palabras del mensaje
		mensajeDividido := strings.Fields(mensaje)
		if len(mensajeDividido) == 0 {
			fmt.Println("comando vacío, intenta: submit <cmd>, status <id>, list, cancel <id>")
			continue
		}

		// 3. Armar el Request según el tipo de operación
		var req protocol.Request
		switch mensajeDividido[0] {
		case "submit":
			if len(mensajeDividido) < 2 {
				fmt.Println("uso: submit <comando> [args...]")
				continue
			}
			req = protocol.Request{
				Type: "submit",
				Cmd:  mensajeDividido[1],
				Args: mensajeDividido[2:],
			}

		case "status", "cancel":
			if len(mensajeDividido) < 2 {
				fmt.Println("uso:", mensajeDividido[0], "<job_id>")
				continue
			}
			// El job_id viaja en el campo Cmd para estas dos operaciones.
			req = protocol.Request{
				Type: mensajeDividido[0],
				Cmd:  mensajeDividido[1],
			}

		case "list":
			req = protocol.Request{Type: "list"}

		default:
			fmt.Println("comando desconocido:", mensajeDividido[0])
			continue
		}

		// 4. Enviar el Request como JSON
		if err := encoder.Encode(req); err != nil {
			fmt.Println("error al enviar datos al servidor:", err)
			continue
		}

		// 5. Recibir la respuesta del servidor
		var respuesta protocol.Response
		if err := decoder.Decode(&respuesta); err != nil {
			fmt.Println("Error o desconexión al recibir respuesta del servidor:", err)
			break
		}

		// 6. Mostrar el resultado según el tipo de operación
		switch req.Type {
		case "submit":
			if respuesta.Ok {
				fmt.Printf(" [RESPUESTA SERVIDOR] Status: OK | JobID: %s\n", respuesta.JobID)
			} else {
				fmt.Printf(" [RESPUESTA SERVIDOR] Status: ERROR | Detalle: %s\n", respuesta.Error)
			}

		case "status":
			if respuesta.Ok {
				fmt.Printf(" JobID: %s | Estado: %s | ExitCode: %d\n", respuesta.JobID, respuesta.Status, respuesta.ExitCode)
				if respuesta.Stdout != "" {
					fmt.Printf("--- stdout ---\n%s\n", respuesta.Stdout)
				}
				if respuesta.Stderr != "" {
					fmt.Printf("--- stderr ---\n%s\n", respuesta.Stderr)
				}
			} else {
				fmt.Printf(" [RESPUESTA SERVIDOR] Status: ERROR | Detalle: %s\n", respuesta.Error)
			}

		case "list":
			if respuesta.Ok {
				if len(respuesta.Jobs) == 0 {
					fmt.Println("(no hay trabajos registrados todavía)")
				}
				for _, jobItem := range respuesta.Jobs {
					fmt.Printf("%s\t%s\tstdout=%q\n", jobItem.ID, jobItem.Estado, jobItem.Stdout)
				}
			} else {
				fmt.Printf(" [RESPUESTA SERVIDOR] Status: ERROR | Detalle: %s\n", respuesta.Error)
			}

		case "cancel":
			if respuesta.Ok {
				fmt.Println(" [RESPUESTA SERVIDOR] Trabajo cancelado.")
			} else {
				fmt.Printf(" [RESPUESTA SERVIDOR] Status: ERROR | Detalle: %s\n", respuesta.Error)
			}
		}
	}
}
