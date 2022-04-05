package main

import (
	"driver/driver"
	"elevator/config"
	"elevator/elevio"
	"elevator/fsm"
	"flag"
	"fmt"
	"os"
	"strconv"
	// "elevator/config"
	// "elevator/elevio"
	// // . "elevator/requests"
	// fsm "elevator/fsm"
	// timer "elevator/timer"
	// . "../Driver-go/timer"
)

func main() {
	fmt.Printf("Startup\n")
	///Declaring variables and default data
	var id string
	var port string

	defaultID := strconv.Itoa(os.Getpid())
	defaultPort := "15657"

	flag.StringVar(&id, "id", defaultID, "ID")
	flag.StringVar(&port, "port", defaultPort, "Set port for this node. Default value set as 15657")
	flag.Parse()

	elevio.Init("localhost:"+port, config.NumFloors)
	fmt.Println("Done with elevio init")
	fsm.Fsm_init()

	fmt.Printf("ID set to: %v. Port set to: %v \n", id, port)
	driver.Elevator_Run(id)
}
