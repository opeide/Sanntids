package main 

import (
	"fmt"
	"net"
)


const (
	LOCAL_ADDR = "localhost"
	LOCAL_PORT = "5432"
	REMOTE_ADDR = "localhost"
	REMOTE_PORT = "5432"
	CONN_TYPE = "udp"
)

func check_error(err error){
	if err != nil{
		fmt.Println("Error: ",err)
	}
}

func main(){
	conn, err := net.Dial(CONN_TYPE, REMOTE_ADDR+":"+REMOTE_PORT)
	check_error(err)
	defer conn.Close()

	for {
		conn.Write([]byte("UDP! Love, Client."))
	}
}