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
		fsm.Elevator1.Dirn = config.MD_Down
		fsm.Elevator1.Behaviour = config.EB_Moving
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

	buttonE := make(chan config.ButtonEvent)
	reassginOrder := make(chan config.ButtonEvent)
	assignedOrder := make(chan nodeConfig.Order)
	orderRx := make(chan nodeConfig.Order)
	orderUpdate := make(chan nodeConfig.OrderEvent)

	orderCleared := make(chan nodeConfig.Order, 15)
	trackOrder := make(chan nodeConfig.Order)

	// go assigner.AssignOrder(nodeUpdateCh, buttonE, orderAssignment, id)
	go assigner.AssignOrder2(id, buttonE, reassginOrder, assignedOrder)
	go distributor.Distribute(id, assignedOrder, reassginOrder, orderRx, peerUpdateCh, orderUpdate, trackOrder, orderCleared)
	go distributor.TrackOrders(trackOrder, orderCleared, orderUpdate, reassginOrder)
	// orderRx := make(chan nodeConfig.Order)
	// go distributor.ReceiveOrder(orderRx)

	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, id, peerUpdateCh, nodeUpdateCh)

	go wd.Watchdog(&fsm.Elevator1)
	fmt.Println("For loop:")
	for {
		select {
		// case <-nodeUpdateCh:
		// case n := <-nodeUpdateCh:
		// nodeIsKnown := false
		// for _, node := range driverConfig.KnownNodes {
		// 	if node.Id == n.Id {
		// 		nodeIsKnown = true
		// 	}
		// }
		// if nodeIsKnown {
		// 	fmt.Printf("From main.Receiver: id: %s, n.Floor: %d\n", n.Id, n.Elevator.Floor) //n) //n.Elevator.Floor)
		// 	nodeToUpdate, _, _ := peers.GetNodeWithId(n.Id)
		// 	*nodeToUpdate = n
		// } else {
		// 	peers.OnNewNode2(n)
		// }
		// fmt.Printf(" Floor: %d\n", driverConfig.KnownNodes[0].Elevator.Floor)
		// node, err := peers.GetNodeWithId("10")
		// if err != 0 {

		// }
		// case p := <-peerUpdateCh:
		// 	fmt.Println("Got peer update")
		// 	// for _,_ := range p.New {
		// 	// fmt.Println(p.New)
		// 	// fmt.Println(p.Peers)
		// 	fmt.Printf(" Peers: %q\n", p.Peers)
		// 	fmt.Printf(" New: %q\n", p.New)
		// fmt.Printf(" New: %q\n", driverConfig.KnownNodes[0])
		// if len(p.New) > 0 && len(p.Peers) >= 1 {
		// // // // peers.OnNewNode(p)
		// // // // var node *driverConfig.Node
		// // // // // ik := 0
		// // // // node, _ = peers.GetNodeWithId("10")
		// // // // fmt.Println(node.Elevator.Floor)

		case a := <-drv_buttons:
			if a.Button == 2 {
				elevio.SetButtonLamp(a.Button, a.Floor, true)
				fsm.Fsm_onRequestButtonPress(a.Floor, a.Button)
			} else {
				buttonE <- a
			}
			// fsm.Fsm_onRequestButtonPress(a.Floor, a.Button)
			// fmt.Printf("Cost: %d", cost.TimeToIdle(fsm.Elevator1))
		case a := <-drv_floors:
			fsm.Fsm_onFloorArrival(a)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				fsm.Elevator1.Obstruction = true
			} else {
				fsm.Elevator1.Obstruction = false
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
