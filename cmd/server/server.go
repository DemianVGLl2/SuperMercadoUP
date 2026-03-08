package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/DemianVGLl2/SuperMercadoUP/internal/logger"
)

func main() {
	if err := logger.Init("server.txt"); err != nil {
		log.Fatal("Error iniciando logger: ", err)
	}

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

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	logger.Log("SERVER", "CONNECT", addr)

	input := bufio.NewScanner(conn)
	for input.Scan() {
		line := strings.TrimSpace(input.Text())
		if line == "" {
			continue
		}

		fmt.Fprintf(conn, "ECHO: %s\n", line)
	}

	logger.Log("SERVER", "DISCONNECT", addr)
}
