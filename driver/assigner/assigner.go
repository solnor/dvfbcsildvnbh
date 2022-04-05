package assigner

import (
	nodeConfig "driver/config"
	"driver/cost"
	elevConfig "elevator/config"
	"fmt"
	"time"
)

//Calculation of best cost-effective elevator
//TODO: Cab orders?
// func assignOrderByCost(nodeUpdateCh chan<- nodeConfig.Node) { //orders elevio.ButtonEvent) {
// 	minimumCost := 100000
// 	var elevatorCost int64
// 	var calculatedElevator *nodeConfig.Node

// 	for _, n := range nodes {
// 		elevatorCost = cost.TimeToIdle(n.Elevator)
// 		if elevatorCost < minimumCost && n.Available {
// 			minimumCost = elevatorCost
// 			calculatedElevator = e
// 		}
// 		//failcheck?
// 	}
// 	return calculatedElevator

// }

// func AssignOrder(nodeUpdateCh <-chan nodeConfig.Node, request chan elevConfig.ButtonEvent, assignedOrder chan Order, id string) { //orders elevio.ButtonEvent) {

// 	orderRx := make(chan Order)
// 	orderTx := make(chan Order)

// 	go bcast.Transmitter(15647, orderTx)
// 	go bcast.Receiver(15647, 0, orderRx)

// 	for {
// 		var order Order
// 		updated := false
// 		select {
// 		case <-nodeUpdateCh:
// 			// nodeUpdateCh <- n
// 		// case <-time.After(15 * time.Millisecond):
// 		case r := <-request:
// 			updated = true
// 			order.Request = r
// 			order.Id = "21771"
// 		}
// 		if updated {
// 			assignedOrder <- order
// 		}
// 	}

// 	// minimumCost := 100000
// 	// var elevatorCost int64
// 	// var calculatedElevator *nodeConfig.Node

// 	// for _, n := range nodes {
// 	// 	elevatorCost = cost.TimeToIdle(n.Elevator)
// 	// 	if elevatorCost < minimumCost && n.Available {
// 	// 		minimumCost = elevatorCost
// 	// 		calculatedElevator = e
// 	// 	}
// 	// 	//failcheck?
// 	// }
// 	// return calculatedElevator

// }

func AssignOrder2(id string, requestCh, reassignCh chan elevConfig.ButtonEvent, assignedOrder chan nodeConfig.Order) {
	// nodeConfig.KnownNodes

	var calculatedElevator *nodeConfig.Node
	var order nodeConfig.Order
	var request elevConfig.ButtonEvent
	for {
		orderReceived := false
		orderAssigned := false
		select {
		case r := <-requestCh:
			request = r
			orderReceived = true
		case r := <-reassignCh:
			request = r
			orderReceived = true
		}
		if orderReceived {
			// fmt.Printf("%q", r)
			var minimumCost int64 = 100000
			var elevatorCost int64
			nodeConfig.KnownNodesMutex.RLock()
			for _, node := range nodeConfig.KnownNodesTable {
				//fmt.Println(node)
				eCopy := elevConfig.DupElevator(node.Elevator)
				eCopy.Requests[request.Floor][request.Button] = 1
				elevatorCost = cost.TimeToIdle(eCopy)
				if elevatorCost < minimumCost && node.Available { //&& r.Button != 2 {
					minimumCost = elevatorCost
					calculatedElevator = node
					orderAssigned = true
					// fmt.Println("node available: ", node.Available)

					//fmt.Println("Calculated Elevator: ", calculatedElevator)
				}
				//failcheck?
			}
			nodeConfig.KnownNodesMutex.RUnlock()
			if orderAssigned {
				// fmt.Println("minimumCost: ", minimumCost)
				order.SenderId = id
				order.AssignedId = calculatedElevator.Id
				order.Request = request
				order.Timestamp = time.Now() //.Add(1000 * time.Millisecond) // hardcoded
				order.Acks = nil
				order.Acks = append(order.Acks, id)
				// order.OneAckIsEnough = false
				order.State = nodeConfig.Order_New
				order.Cost = minimumCost

				// CreateOrder()
				fmt.Printf("Assigner: Assigned ID: %s\n", order.AssignedId)

				assignedOrder <- order
			}

		}
	}
	// return calculatedElevator

}

// func CreateOrder()

// func calculateCost()
