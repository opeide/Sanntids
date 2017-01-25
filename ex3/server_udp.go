package main


import (
	"fmt"
	"net"
)

const (
	SERVER_PORT = "3000"
	CONN_TYPE = "udp"
)

func check_error(err error){
	if err != nil{
		fmt.Println("Error: ",err)
	}
}


func main(){
	serverAddr, err := net.ResolveUDPAddr(CONN_TYPE, ":"+SERVER_PORT)
	check_error(err)

	serverConn, err := net.ListenUDP(CONN_TYPE, serverAddr)
	check_error(err)
	defer serverConn.Close()

	buf := make([]byte, 1024)

	for {
		msg_len, addr, err := serverConn.ReadFromUDP(buf)
		check_error(err)
		fmt.Println("Received massage from ",addr)
		fmt.Println(string(buf[0:msg_len]))
	}
}