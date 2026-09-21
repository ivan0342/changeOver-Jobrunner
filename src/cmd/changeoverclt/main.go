package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
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

	for {
		mensaje := bufio.NewReader(os.Stdin)
		fmt.Print(">")
		// 2. Enviamos un mensaje al servidor (Ojo: debe terminar en \n)
		text, _ := mensaje.ReadString('\n')
		_, err = conexion.Write([]byte(text))

		respuesta, err := bufio.NewReader(conexion).ReadString('\n')
		if err != nil {
			fmt.Println("Error en la respuesta del servidor")
			return
		}
		fmt.Println("RESPUES RECIBIDA:", respuesta)

	}

}
