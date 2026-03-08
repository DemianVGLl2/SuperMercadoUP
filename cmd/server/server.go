package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/DemianVGLl2/SuperMercadoUP/internal/logger"
	"github.com/DemianVGLl2/SuperMercadoUP/internal/models"
)

func main() {
	if err := logger.Init("server.log"); err != nil {
		log.Fatal("Error iniciando logger: ", err)
	}

	store := models.NewStore()

	listener, err := net.Listen("tcp", "localhost:8000")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Listening to: " + listener.Addr().String())

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn, store)
	}
}

func handleConn(conn net.Conn, store *models.Store) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	logger.Log("SERVER", "CONNECT", addr)

	cart := models.Cart{}
	_ = cart // esto solo es para evitar errores

	input := bufio.NewScanner(conn)
	for input.Scan() {
		line := strings.TrimSpace(input.Text())
		if line == "" {
			continue
		}

		// Aquí Sophía puede hacer todo lo de los comandos con el carrito :D

		fmt.Fprintf(conn, "ECHO: %s\n", line)
	}

	logger.Log("SERVER", "DISCONNECT", addr)
}
