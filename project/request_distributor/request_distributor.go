package request_distributor

import (
	"../message_structs"
	"../network/peers"
	"fmt"
)

func Distribute_requests(id string,
	peer_update_chan <-chan peers.PeerUpdate,
	network_request_rx_chan <-chan message_structs.Request,
	network_request_tx_chan chan<- message_structs.Request,
	button_request_chan <-chan message_structs.Request,
	requests_to_execute_chan chan<- message_structs.Request,
	executed_requests_chan <-chan message_structs.Request, 
	set_lamp_chan chan<- message_structs.Set_lamp_message) {

	for {
		select {
		case button_request := <-button_request_chan:
			button_request.Primary_responsible_elevator = id
			requests_to_execute_chan <- button_request
		case executed_request := <-executed_requests_chan:
			fmt.Println("Distributor: Executor completed request: ", executed_request)
		}
	}
}
