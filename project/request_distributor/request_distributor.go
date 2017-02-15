package request_distributor

import (
	"../global"
	"../network/peers"
)

func Distribute_requests(peer_update_chan <-chan peers.PeerUpdate,
	network_request_rx <-chan global.Request,
	network_request_tx chan<- global.Request) {

}
