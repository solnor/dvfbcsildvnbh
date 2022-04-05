package driver

import (
	"driver/assigner"
	nodeConfig "driver/config"
	"driver/distributor"
	wd "driver/watchdog"
	"elevator/config"
	"elevator/elevio"
	fsm "elevator/fsm"
	"elevator/timer"
	"flag"
	"fmt"
	"network/peers"
	"os"
	"runtime"
	"strconv"
)

func Elevator_Run() {
	runtime.GOMAXPROCS(100)
	///Declaring variables and default data
	var id string
	var port string

	defaultID := strconv.Itoa(os.Getpid())
	defaultPort := "15657"

	//go run main.go -id "ID" -port "PORT"
	flag.StringVar(&id, "id", defaultID, "ID")
	flag.StringVar(&port, "port", defaultPort, "Set port for this node. Default value set as 15657")
	flag.Parse()

	elevio.Init("localhost:"+port, config.NumFloors)
	fmt.Println("Done with elevio init")
	fsm.Fsm_init()

	fmt.Printf("ID set to: %v. Port set to: %v \n", id, port)
	// backup.BackupInit(id, port)
	/////////////////////////////////////////////////////////////////////////////////////////////////////

	drv_buttons := make(chan config.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	var doorTimeout = make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go timer.Observer(doorTimeout)

	if elevio.GetFloor() == -1 {
		elevio.SetMotorDirection(config.MD_Down)
		fsm.ThisElevator.Dirn = config.MD_Down
		fsm.ThisElevator.Behaviour = config.EB_Moving
	}
	nodeConfig.Node_Init(id)
	// ThisNode := nodeConfig.NewNode(id)
	// ThisNode.Elevator = fsm.Elevator1
	// // nodeConfig.KnownNodes = make(map[string])
	// // nodeConfig.KnownNodes = append(nodeConfig.KnownNodes, &thisNode) //MOVE THIS
	// nodeConfig.KnownNodesTable[id] = &ThisNode

	peerTxEnable := make(chan bool)
	peerUpdateCh := make(chan peers.PeerUpdate)
	nodeUpdateCh := make(chan nodeConfig.Node)

	assignRequest := make(chan config.ButtonEvent)
	reassginOrder := make(chan config.ButtonEvent)
	assignedOrder := make(chan nodeConfig.Order)
	orderRx := make(chan nodeConfig.Order)
	orderUpdate := make(chan nodeConfig.OrderEvent)

	orderCleared := make(chan nodeConfig.Order, 15)
	trackOrder := make(chan nodeConfig.Order)

	// go assigner.AssignOrder(nodeUpdateCh, buttonE, orderAssignment, id)
	go assigner.AssignOrder(id, assignRequest, reassginOrder, assignedOrder)
	go distributor.Distribute(id, assignedOrder, reassginOrder, orderRx, peerUpdateCh, orderUpdate, trackOrder, orderCleared)
	go distributor.TrackOrders(trackOrder, orderCleared, orderUpdate, reassginOrder)
	// orderRx := make(chan nodeConfig.Order)
	// go distributor.ReceiveOrder(orderRx)

	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, id, peerUpdateCh, nodeUpdateCh)

	go wd.Watchdog(&fsm.ThisElevator)
	for {
		select {
		case buttonEvent := <-drv_buttons:
			if buttonEvent.Button == 2 {
				elevio.SetButtonLamp(buttonEvent.Button, buttonEvent.Floor, true)
				fsm.Fsm_onRequestButtonPress(buttonEvent.Floor, buttonEvent.Button)
			} else {
				assignRequest <- buttonEvent
			}
		case floor := <-drv_floors:
			fsm.Fsm_onFloorArrival(floor)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				fsm.ThisElevator.Obstruction = true
			} else {
				fsm.ThisElevator.Obstruction = false
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < config.NumFloors; f++ {
				for b := config.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		case <-doorTimeout:
			fsm.Fsm_onDoorTimeout()

		case a := <-orderRx:
			if a.AssignedId == id {
				fsm.Fsm_onRequestButtonPress(a.Request.Floor, a.Request.Button)
			}
		case order := <-orderUpdate:
			elevio.SetButtonLamp(order.Request.Button, order.Request.Floor, order.Confirmed)
		}
	}
}
