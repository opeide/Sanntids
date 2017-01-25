package main

import (
	"fmt"
	"net"
	"os"
)

const (
	REMOTE_IP   = "129.241.187.43"
	REMOTE_PORT = "34933"
	CONN_TYPE   = "tcp"
)

func main() {
	fmt.Println("Launching server...")
	ln, err := net.ListenTCP(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer ln.Close()

	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)

	for {
		connection, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(connection)
	}
}

func handleRequest(connection net.Conn) {
	buf := make([]byte, 1024)
	_, err := connection.Read(buf)
	if err != nil {
		fmt.Println("Error reading: ", err.Error())
	}
	connection.Write([]byte("Message received! Love, server."))
	connection.Close()
}
