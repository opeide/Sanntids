package main

import (
	"fmt"
	"net"
)

const (
	REMOTE_IP   = "129.241.187.43"
	REMOTE_PORT = "34933"
	LOCAL_IP    = "129.241.187.159"
	LOCAL_PORT  = "20011"
	CONN_TYPE   = "tcp"
)

func check_error(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func main() {
	remoteAddr, err := net.ResolveTCPAddr(CONN_TYPE, REMOTE_IP+":"+REMOTE_PORT)
	check_error(err)
	localAddr, err := net.ResolveTCPAddr(CONN_TYPE, LOCAL_IP+":"+LOCAL_PORT)
	check_error(err)
	conn, err := net.DialTCP(CONN_TYPE, localAddr, remoteAddr)
	check_error(err)
	defer conn.Close()

	// Get welcome message
	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading: ", err.Error())
	}
	fmt.Println(string(buf[:]))

	conn.Write([]byte("Echo this!"))
	buf = make([]byte, 1024)
	_, err = conn.Read(buf)
	check_error(err)
	fmt.Println(string(buf[:]))

	conn.Close()

	/*_, err = conn.Write([]byte("Connect to: " + LOCAL_IP + ":" + LOCAL_PORT))
	check_error(err)
	conn.Close()

	listener, err := net.ListenTCP(CONN_TYPE, localAddr)
	check_error(err)
	newConn, err := listener.AcceptTCP()
	check_error(err)
	defer newConn.Close()

	// Write and read the echo
	newConn.Write([]byte("Echo this!"))
	_, err = newConn.Read(buf)
	if err != nil {
		fmt.Println("Error reading: ", err.Error())
	}
	fmt.Println(string(buf[:]))

	newConn.Close()*/
}
