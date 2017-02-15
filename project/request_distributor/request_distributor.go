package request_distributor

import (
	"../global"
	"../network/peers"
	"fmt"
)

func Distribute_requests(peer_update_chan <-chan peers.PeerUpdate,
	network_request_rx_chan <-chan global.Request,
	network_request_tx_chan chan<- global.Request,
	button_request_chan <-chan global.Request) {

	for {
		select {
		case button_request := <-button_request_chan:
			fmt.Println("Got a button request, ", button_request)
		}
	}
}
