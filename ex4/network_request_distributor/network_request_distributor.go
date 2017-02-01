package network_request_distributor

import (
	"net"
	"time"
)

const (
	com_port             = "20011"
	i_am_elevator_string = "I am an elevator"
)

var local_ip string = get_local_ip()

func get_local_ip() string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	// Taken from jniltinho on GitHub and changed a bit
	// at https://gist.github.com/jniltinho/9787946
	for _, address := range addresses {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	// End copied segment

	return ""
}

func Start() {
	new_elevators := make(chan string)
	local_address, _ := net.ResolveUDPAddr("udp", ":"+com_port)
	udp_connection, _ := net.ListenUDP("udp", local_address)
	defer udp_connection.Close()
	remote_address, _ := net.ResolveUDPAddr("udp", net.IPv4bcast.String()+":"+com_port)
	go say_i_am_elevator(udp_connection, remote_address)
	go listen_for_new_elevators(udp_connection, new_elevators)
}

func say_i_am_elevator(udp_connection *net.UDPConn, remote_address *net.UDPAddr) {
	for {
		udp_connection.WriteToUDP([]byte(i_am_elevator_string), remote_address)
		time.Sleep(500 * time.Millisecond)
	}
}

func listen_for_new_elevators(udp_connection *net.UDPConn, new_elevators chan) {
	buffer := make([]byte, 1024)
	for {
		message_length, address, _ := udp_connection.ReadFromUDP(buffer)
		if string(buffer[0:message_length]) == i_am_elevator_string{
			new_elevators = <- 
		}

		time.Sleep(500 * time.Millisecond)
	}
}
