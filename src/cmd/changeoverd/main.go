package main

import (
	"changeover/src/internal/jobmanager"
	"changeover/src/internal/protocol"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

func main() {
	jm := jobmanager.NewManager()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error al abrir puerto", err)
		return
	}
	defer listener.Close()
	fmt.Println("Servidor escuchando en el puerto 8080")

	//aquie se espera a que alguien se conecte con el
	for {
		conexion, err := listener.Accept()
		if err != nil {
			fmt.Println("error al conectar con el cliente")
			continue
		}
		go manejarCliente(conexion, jm)
	}

}

func manejarCliente(conexion net.Conn, jm *jobmanager.JobManager) {
	defer conexion.Close()
	fmt.Println("el cliente se pudo conectar desde: ", conexion.RemoteAddr())

	decoder := json.NewDecoder(conexion)
	encoder := json.NewEncoder(conexion)
	for {
		var tokensRecibidos protocol.Request
		//var response protocol.Response;
		// Decode se queda esperando bytes, lee el JSON y lo transforma de vuelta a []string
		err := decoder.Decode(&tokensRecibidos)
		if err != nil {
			if err == io.EOF {
				fmt.Println("[SERVIDOR] El cliente cerró la conexión.")
			} else {
				fmt.Println("[SERVIDOR] Error al decodificar JSON:", err)
			}
			return
		}
		var resp protocol.Response
		switch tokensRecibidos.Type {
		case "status", "cancel":
			// para status/cancel, lo que el cliente manda en Cmd es el job_id
			// (ver ajuste sugerido abajo en el cliente)
			resp = jm.DefFunc(tokensRecibidos.Type, protocol.Job{}, tokensRecibidos.Cmd)
		case "submit":
			trabajo := protocol.Job{Comando: tokensRecibidos.Cmd, Argumentos: tokensRecibidos.Args}
			resp = jm.DefFunc(tokensRecibidos.Type, trabajo, "")
		default:
			resp = jm.DefFunc(tokensRecibidos.Type, protocol.Job{}, "")
		}

		encoder.Encode(resp)

	}

}
