package main

import (
	"./elev"
	"./global"
	"./network/bcast"
	"./network/localip"
	"./network/peers"
	//"./request_distributor"
	"fmt"
	"os"
	"os/exec"
	"time"
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
	peer_tx_enable := make(chan bool) // Currently not in use, but needed to run the peers.Receiver
	go peers.Transmitter(peer_update_port, id, peer_tx_enable)
	go peers.Receiver(peer_update_port, peer_update_chan)
	network_request_rx := make(chan global.Request)
	network_request_tx := make(chan global.Request)
	go bcast.Transmitter(network_request_port, network_request_tx)
	go bcast.Receiver(network_request_port, network_request_rx)

	var last_num int = 0
	var peers []string
	peer_update := <-peer_update_chan
	peers = peer_update.Peers

	wait_for_peers_interval := time.Second
	timer_channel := time.NewTimer(wait_for_peers_interval).C
	select {
	case peer_update = <-peer_update_chan:
		peers = peer_update.Peers
	case <-timer_channel:
	}

	if len(peers) > 1 { // We are also in this list
		// Be a backup
		fmt.Println("I am a backup")
	backup_loop:
		for {
			select {
			case peer_update = <-peer_update_chan:
				if len(peer_update.Lost) > 0 {
					break backup_loop
				}
			case received_message := <-network_request_rx:
				last_num = received_message.Floor
			}
		}
	}

	err = exec.Command("gnome-terminal", "-x", "sh", "-c", "go run main.go").Run()
	if err != nil {
		fmt.Println("Could not make a backup. ")
		fmt.Println(err)
	}

	// Be a primary
	fmt.Println("I am a primary, last_num =", last_num)
	for {
		num := last_num + 1
		backup_message := global.Request{Floor: num}
		network_request_tx <- backup_message
		fmt.Println(num)

		last_num = num
		time.Sleep(time.Second)
	}

	//go request_distributor.Distribute_requests(peer_update_chan, network_request_rx, network_request_tx)

	select {}
}
