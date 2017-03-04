package request_distributor

import (
	"../message_structs"
	"../network/peers"
	//"fmt"
)

func Distribute_requests(peer_update_chan <-chan peers.PeerUpdate,
	network_request_rx_chan <-chan message_structs.Request,
	network_request_tx_chan chan<- message_structs.Request,
	button_request_chan <-chan message_structs.Request,
	requests_to_execute_chan chan<- message_structs.Request,
	executed_requests_chan <-chan message_structs.Request) {

	for {
		select {
		case button_request := <-button_request_chan:
			requests_to_execute_chan <- button_request
		}
	}
}
