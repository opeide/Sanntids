package main

import (
	"fmt"
	"net"
	"time"
)

const (
	LOCAL_PORT  = "20011"
	REMOTE_ADDR = "129.241.187.43"
	REMOTE_PORT = "20011"
	CONN_TYPE   = "udp"
)

func check_error(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func main() {
	fmt.Println("Starting client...")
	localAddr, err := net.ResolveUDPAddr(CONN_TYPE, ":"+LOCAL_PORT)
	check_error(err)
	remoteAddr, err := net.ResolveUDPAddr(CONN_TYPE, REMOTE_ADDR+":"+REMOTE_PORT)
	check_error(err)

	udpConn, err := net.ListenUDP(CONN_TYPE, localAddr)
	check_error(err)

	defer udpConn.Close()

	buf := make([]byte, 1024)
	for {
		fmt.Println("Sending message ...")
		_, err := udpConn.WriteToUDP([]byte("UDP! Love, Client."), remoteAddr)
		check_error(err)
		fmt.Println("Message sent. ")

		fmt.Println("Reading message ...")
		msgLen, addr, err := udpConn.ReadFromUDP(buf)
		check_error(err)
		fmt.Println("Message received from", addr)
		fmt.Println(string(buf[0:msgLen]))

		time.Sleep(500 * time.Millisecond)
	}
}
