package request_distributor

import (
	"../network/peers"
	"../request"
	"fmt"
)

func Distribute_requests(peer_update_chan <-chan peers.PeerUpdate,
	network_request_rx_chan <-chan request.Request,
	network_request_tx_chan chan<- request.Request,
	button_request_chan <-chan request.Request,
	requests_to_execute_chan chan<- request.Request,
	executed_requests_chan <-chan requesst.Request) {

	for {
		select {
		case button_request := <-button_request_chan:
			requests_to_execute_chan <- button_request
		}
	}
}
