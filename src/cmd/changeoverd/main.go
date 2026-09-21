package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
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
		go manejarCliente(conexion)
	}

}

func manejarCliente(conexion net.Conn) {
	defer conexion.Close()
	fmt.Println("el cliente se pudo conectar desde: ", conexion.RemoteAddr())

	lector := bufio.NewReader(conexion)
	for {
		mensaje, err := lector.ReadString('\n')
		if err != nil {
			fmt.Println("[SERVIDOR] cliente desconectado")
			return
		}
		fmt.Println("el mensaje es", mensaje)
		respuesta := strings.ToUpper(mensaje)
		conexion.Write([]byte("Servidor dice: " + respuesta))
	}

}
