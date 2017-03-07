package main

import (
	"./hardware_interface"
	"./message_structs"
	"./network/bcast"
	"./network/localip"
	"./network/peers"
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
	button_request_chan := make(chan message_structs.Request, 1) // 50 because ???????????????????????????
	floor_changes_chan := make(chan int, 1)
	set_motor_direction_chan := make(chan int, 1)
	set_lamp_chan := make(chan message_structs.Set_lamp_message, 1)

	go hardware_interface.Read_and_write_to_hardware(
		button_request_chan,
		floor_changes_chan,
		set_motor_direction_chan,
		set_lamp_chan)

	requests_to_execute_chan := make(chan message_structs.Request, 1)
	executed_requests_chan := make(chan message_structs.Request, 1)
	elevator_state_changes_chan := make(chan message_structs.Elevator_state, 1) // Burde være 3 eller noe slikt? Ettersom det kommer 3 på rad av og til. 

	go request_executor.Execute_requests(
		requests_to_execute_chan,
		executed_requests_chan,
		floor_changes_chan,
		set_motor_direction_chan, 
		set_lamp_chan, 
		elevator_state_changes_chan)

	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		localIP = "DISCONNECTED"
	}
	id := fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	fmt.Println("MY ID: ",id)

	peer_update_chan := make(chan peers.PeerUpdate, 1)
	peer_tx_enable_chan := make(chan bool, 1) // Currently not in use, but needed to run the peers.Receiver

	go peers.Transmitter(peer_update_port, id, peer_tx_enable_chan)
	go peers.Receiver(peer_update_port, peer_update_chan)

	network_request_rx_chan := make(chan message_structs.Request, 1)
	network_request_tx_chan := make(chan message_structs.Request, 1)

	go bcast.Transmitter(network_request_port, network_request_tx_chan)
	go bcast.Receiver(network_request_port, network_request_rx_chan)

	go request_distributor.Distribute_requests(
		id,
		peer_update_chan,
		network_request_rx_chan,
		network_request_tx_chan,
		button_request_chan,
		requests_to_execute_chan,
		executed_requests_chan, 
		set_lamp_chan, 
		elevator_state_changes_chan)

	select {}
}
