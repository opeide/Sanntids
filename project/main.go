package main

import (
	"./hardware_interface"
	"./message_structs"
	"./network/bcast"
	"./network/localip"
	"./network/peers"
	"./request_distributor"
	"./request_executor"
	"./request_watchdog"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

const (
	peer_update_port     = 20110
	network_request_port = 20111

	check_maker_alive_period_ms = 50 // milli Seconds
)

func main() {

	// Be a backup
	var made_by_pid string
	flag.StringVar(&made_by_pid, "made_by_pid", "", "pid of process that started this process")
	flag.Parse()
	fmt.Println("I was made by process id:", made_by_pid)
	if made_by_pid != "" {
		made_by_pid, _ := strconv.Atoi(made_by_pid)
		maker_alive := true
		for maker_alive {
			select {
			case <-time.After(time.Millisecond * check_maker_alive_period_ms):
				process, err := os.FindProcess(made_by_pid)
				if err != nil {
					fmt.Println("Failed to find process id: ", made_by_pid)
					maker_alive = false
				} else {
					err := process.Signal(syscall.Signal(0))
					//fmt.Println("process.Signal on pid", made_by_pid, "returned:", err)
					if err != nil {
						maker_alive = false
					}
				}
			}
		}
	}

	// be the elevator
	// Init hardware to stop in case moving
	button_request_chan := make(chan message_structs.Request, 1) // 50 because ???????????????????????????
	floor_changes_chan := make(chan int, 1)
	set_motor_direction_chan := make(chan int, 1)
	set_lamp_chan := make(chan message_structs.Set_lamp_message, 1)

	go hardware_interface.Read_and_write_to_hardware(
		button_request_chan,
		floor_changes_chan,
		set_motor_direction_chan,
		set_lamp_chan)

	// Make the backup
	pid := os.Getpid()
	fmt.Println("My process id:", pid)
	err := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run main.go -made_by_pid="+strconv.Itoa(pid)).Run()
	if err != nil {
		fmt.Println("Could not make a backup. ")
		fmt.Println(err)
		//todo: restart computer [in design]
	}

	//<-time.After(time.Second * 10) //todo: make const. let backup be generated before continuing

	requests_to_execute_chan := make(chan message_structs.Request, 1)
	executed_requests_chan := make(chan message_structs.Request, 1)
	local_elevator_state_changes_chan := make(chan message_structs.Elevator_state, 1) // Burde være 3 eller noe slikt? Ettersom det kommer 3 på rad av og til.

	go request_executor.Execute_requests(
		executed_requests_chan,
		set_motor_direction_chan,
		set_lamp_chan,
		local_elevator_state_changes_chan,
		floor_changes_chan,
		requests_to_execute_chan)

	// Network module
	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		fmt.Println("DISCONNECTED FROM NETWORK AT STARTUP. EXITING")
		os.Exit(0) //Lets backup take over (effectively a program restart)
	}
	id := fmt.Sprintf("peer-%s", localIP) //fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())

	time.After(time.Millisecond * 500) //Let others detect this elevator was lost from network

	peer_update_chan := make(chan peers.PeerUpdate, 1)
	peer_tx_enable_chan := make(chan bool, 1) // Currently not in use, but needed to run the peers.Receiver
	go peers.Transmitter(peer_update_port, id, peer_tx_enable_chan)
	go peers.Receiver(peer_update_port, peer_update_chan)
	network_request_rx_chan := make(chan message_structs.Request, 1)
	network_request_tx_chan := make(chan message_structs.Request, 1)
	local_elevator_state_changes_tx_chan := make(chan message_structs.Elevator_state, 1)
	non_local_elevator_state_changes_rx_chan := make(chan message_structs.Elevator_state, 1)
	go bcast.Transmitter(network_request_port, network_request_tx_chan, local_elevator_state_changes_tx_chan)
	go bcast.Receiver(network_request_port, network_request_rx_chan, non_local_elevator_state_changes_rx_chan)

	timed_out_requests_chan := make(chan message_structs.Request, 1)
	request_watchdog.Init(id, timed_out_requests_chan)

	go request_distributor.Distribute_requests(
		id,
		network_request_tx_chan,
		local_elevator_state_changes_tx_chan,
		requests_to_execute_chan,
		set_lamp_chan,
		peer_update_chan,
		network_request_rx_chan,
		non_local_elevator_state_changes_rx_chan,
		button_request_chan,
		executed_requests_chan,
		local_elevator_state_changes_chan,
		timed_out_requests_chan)

	select {}
}
