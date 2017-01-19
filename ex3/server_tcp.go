package main


import (
	"fmt"
	"net"
	"os"
)


const (
	CONN_HOST = "localhost"
	CONN_PORT = "5432"
	CONN_TYPE = "tcp"
)

func main(){
	fmt.Println("Launching server...")
	ln, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil{
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer ln.Close()
	
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	
	for {
		connection, err := ln.Accept()
		if err != nil{
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(connection)
	}
}


func handleRequest(connection net.Conn){
	buf := make([]byte, 1024)
	_ , err := connection.Read(buf)
	if err != nil{
		fmt.Println("Error reading: ", err.Error())
	}
	connection.Write([]byte("Message received! Love, server."))
	connection.Close()
} 