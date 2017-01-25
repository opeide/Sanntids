package main

import (
	"fmt"
	"net"
	"os"
)


const (
	CONN_HOST = "localhost" 
	CONN_PORT = "34933"
	CONN_TYPE = "tcp"
)

func main(){
	connection, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT) 
	if err != nil{
		fmt.Println("Error dialing: ", err.Error())
		os.Exit(1)
	}
	connection.Write([]byte("Hello! Love, Client"))

	buf := make([]byte, 1024)
	_ , err = connection.Read(buf)
	if err != nil{
		fmt.Println("Error reading: ", err.Error())
	}
	fmt.Println(string(buf[:]))

	connection.Close()
}