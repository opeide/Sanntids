package main

import (
	"./button_request_acquirer"
	"./elev"
	"./global"
	"./network/bcast"
	"./network/localip"
	"./network/peers"
	"./request_distributor"
	"fmt"
	"os"
)

const (
	peer_update_port     = 20110
	network_request_port = 20111
)

func main() {
	elev.Elev_init()

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
	network_request_rx_chan := make(chan global.Request)
	network_request_tx_chan := make(chan global.Request)
	go bcast.Transmitter(network_request_port, network_request_tx_chan)
	go bcast.Receiver(network_request_port, network_request_rx_chan)

	button_request_chan := make(chan global.Request)
	go button_request_acquirer.Acquire_button_requests(button_request_chan)

	go request_distributor.Distribute_requests(peer_update_chan,
		network_request_rx_chan,
		network_request_tx_chan,
		button_request_chan)

	select {}
}
