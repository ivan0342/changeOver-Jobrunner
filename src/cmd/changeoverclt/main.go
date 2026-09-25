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
	// 1. "Marcamos" el número del servidor (IP local y puerto 8080)
	conexion, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println(" [CLIENTE] No se pudo conectar al servidor:", err)
		return
	}

	defer conexion.Close() // Nos aseguramos de colgar al terminar
	fmt.Println(" [CLIENTE] ¡Conectado exitosamente!")

	decoder := json.NewDecoder(conexion)
	for {

		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)

		// 1. Leer la entrada del usuario
		mensaje, _ := reader.ReadString('\n')
		// Limpiamos los saltos de línea (\n o \r\n en Windows)
		mensaje = strings.TrimSpace(mensaje)

		// 3. Separar las palabras del mensaje en un slice temporal
		mensajeDividido := strings.Fields(mensaje)

		// Colocamos el ID primero y luego "expandimos" (...) las palabras dentro del slice

		req := protocol.Request{Type: mensajeDividido[0], Cmd: mensajeDividido[1], Args: mensajeDividido[2:]}
		bytesJson, _ := json.Marshal(req)

		if err != nil {
			fmt.Println("Error al codificar los datos en json", err)
			continue
		}

		//enviamos los datos
		_, err = conexion.Write(bytesJson)
		if err != nil {
			fmt.Println("error al enviar datos al servido", err)
		}

		// 3. RECIBIMOS LA RESPUESTA DEL SERVIDOR
		var respuesta protocol.Response // Asegúrate de tener esta estructura en protocol
		err = decoder.Decode(&respuesta)
		if err != nil {
			fmt.Println("Error o desconexión al recibir respuesta del servidor:", err)
			break
		}

		// 4. Mostramos el resultado en pantalla
		if respuesta.Ok {
			fmt.Printf(" [RESPUESTA SERVIDOR] Status: OK | JobID: %s\n", respuesta.JobID)
		} else {
			fmt.Printf(" [RESPUESTA SERVIDOR] Status: ERROR | Detalle: %s\n", respuesta.Error)
		}

	}

}
