package main

import (
	"fmt"
	"net"
)

const (
	LOCAL_PORT = "30000"
	CONN_TYPE  = "udp"
)

func check_error(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func main() {
	fmt.Println("Listening for UDP on port", LOCAL_PORT)
	listenerAddr, err := net.ResolveUDPAddr(CONN_TYPE, ":"+LOCAL_PORT)
	check_error(err)

	serverConn, err := net.ListenUDP(CONN_TYPE, listenerAddr)
	check_error(err)
	defer serverConn.Close()

	buf := make([]byte, 1024)

	for {
		msgLen, addr, err := serverConn.ReadFromUDP(buf)
		check_error(err)
		fmt.Println("Received massage from ", addr)
		fmt.Println(string(buf[0:msgLen]))
	}
}
