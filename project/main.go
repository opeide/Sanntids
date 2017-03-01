package main

import (
	"./hardware_interface"
	"./network/bcast"
	"./network/localip"
	"./network/peers"
	"./request"
	"./request_distributor"
	"./request_executor"
	"fmt"
	"os"
)

const (
	peer_update_port     = 20110
	network_request_port = 20111
)

func main() {
	button_request_chan := make(chan request.Request)
	floor_changes_chan := make(chan int)
	go hardware_interface.Read_and_write_to_hardware(
		button_request_chan, 
		floor_changes_chan)

	requests_to_execute_chan := make(chan request.Request)
	executed_requests_chan := make(chan request.Request)
	go request_executor.Execute_requests(
		requests_to_execute_chan, 
		executed_requests_chan, 
		floor_changes_chan)

	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		localIP = "DISCONNECTED"
	}
	id := fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())

	peer_update_chan := make(chan peers.PeerUpdate)
	peer_tx_enable_chan := make(chan bool) // Currently not in use, but needed to run the peers.Receiver
	go peers.Transmitter(peer_update_port, id, peer_tx_enable_chan)
	go peers.Receiver(peer_update_port, peer_update_chan)
	network_request_rx_chan := make(chan request.Request)
	network_request_tx_chan := make(chan request.Request)
	go bcast.Transmitter(network_request_port, network_request_tx_chan)
	go bcast.Receiver(network_request_port, network_request_rx_chan)

	go request_distributor.Distribute_requests(
		peer_update_chan,
		network_request_rx_chan,
		network_request_tx_chan,
		button_request_chan,
		requests_to_execute_chan,
		executed_requests_chan)

	select {}
}
