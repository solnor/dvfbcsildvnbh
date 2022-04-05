package distributor

import (
	nodeConfig "driver/config"
	elevConfig "elevator/config"
	"fmt"
	"network/bcast"
	"network/peers"
	"time"
)

func getUpdatedAckList(id string, current, new []string) []string {
	temp := make([]string, 0)
	orderAcked := false
	for _, receivedAck := range new {
		newAck := true
		for _, currentAck := range current {
			if receivedAck == currentAck {
				newAck = false
			}
		}
		if receivedAck == id {
			orderAcked = true
		}
		if newAck {
			// tempAck = append(tempAck, receivedAck)
			temp = append(temp, receivedAck) // Make a list of every ack node doesn't know about
			// orderUpdated = true
		}
	}
	// fmt.Printf("Temp ack in getUpdatedAckList: %s\n", temp)
	if !orderAcked {
		fmt.Printf("Fucking fuck appended with %s\n", id)
		temp = append(temp, id)
	}
	temp = append(current, temp...)
	return temp
}

func contains(slice []string, value string) bool {
	for _, element := range slice {
		if element == value {
			return true
		}
	}
	return false
}

func getUpdatedAckList2(id string, current, new []string) []string {
	temp := make([]string, 0)
	for _, ack := range new {
		if !contains(current, ack) {
			temp = append(temp, ack)
		}
	}
	current = append(current, temp...)
	if !contains(current, id) {
		current = append(current, id)
	}
	return current
}

func getUpdatedAckList3(id string, current, new []string) []string {
	temp := make([]string, 0)
	for _, ack := range new {
		if !contains(current, ack) {
			temp = append(temp, ack)
		}
	}
	current = append(current, temp...)
	return current
}

func addSelfToAckList(id string, current, new []string) []string {
	temp := make([]string, 0)
	orderAcked := false
	for _, ack := range current {
		if ack == id {
			orderAcked = true
			break
		}
	}
	if !orderAcked {
		temp = append(current, id)
		return temp
		// orderUpdated = true
	}
	return current
	// currentStage = nodeConfig.Order_Ack
}

//Args: id, orderCh = Assinger/Distributor interface, orderOut = Driver/Distributor, peerUpdateCh
func Distribute(id string, orderCh chan nodeConfig.Order, reassignCh chan elevConfig.ButtonEvent, orderOut chan nodeConfig.Order, peerUpdateCh chan peers.PeerUpdate, orderUpdate chan nodeConfig.OrderEvent) {

	orderTx := make(chan nodeConfig.Order)
	orderRx := make(chan nodeConfig.Order)
	go bcast.Transmitter(15648, orderTx)
	go bcast.Receiver(15648, 0, orderRx)

	// var orderstates [4][2] ?? struct{owner, list of acks, timeout}

	// ordermessage {assigned, sender, floor, dir, type[new, ack, confirmed, cleared]}

	const ORDER_SEND_TIMEOUT_MS = 1500

	//OrderList = make([][2]nodeConfig.Node, elevConfig.NumFloors)
	var localOrders [4][2]nodeConfig.Order
	// var prevOrder [4][2]int
	var availableNodes []string
	// var lastState [4][2]nodeConfig.OrderType
	enoughAcks := 2
	// acksIncomingOrder := 0
	// acksEstablishedOrder := 0
	// TIMEOUT := 1000
	// orderReceived := false

	for floor, floors := range localOrders {
		for button, _ := range floors {
			localOrders[floor][button].State = nodeConfig.Order_Cleared
			// prevOrder[floor][button] = 0
			// fmt.Println(localOrders[floor][button])
		}
	}

	flr := 0
	btn := 0

	for {
		select {
		case order := <-orderCh:
			// fmt.Printf("??? %q\n", order)
			orderTx <- order
		case p := <-peerUpdateCh:
			fmt.Println("Got peer update")
			// for _,_ := range p.New {
			// fmt.Println(p.New)
			// fmt.Println(p.Peers)
			fmt.Printf(" Peers: %q\n", p.Peers)
			fmt.Printf(" New: %q\n", p.New)
			availableNodes = p.Peers
			enoughAcks = len(availableNodes)
			fmt.Println(enoughAcks)

		case order := <-orderRx:
			currentOrder := localOrders[order.Request.Floor][order.Request.Button]
			// fmt.Printf("%p\n", &currentOrder)
			currentState := currentOrder.State
			// lastState := currentOrder.State
			if currentOrder.State == nodeConfig.Order_Cleared && order.State == nodeConfig.Order_New {
				fmt.Printf("[%s]: new order\n", time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
				currentOrder.State = nodeConfig.Order_New
			}

			switch currentOrder.State {
			case nodeConfig.Order_New:
				currentOrder.AssignedId = order.AssignedId
				currentOrder.SenderId = order.SenderId
				currentOrder.Timestamp = order.Timestamp
				currentOrder.Request.Floor = order.Request.Floor
				currentOrder.Request.Button = order.Request.Button
				currentOrder.Acks = nil
				// currentOrder.Acks = append(currentOrder.Acks, id)
				switch id {
				case currentOrder.AssignedId:
					orderOut <- currentOrder
					if id == currentOrder.SenderId { // Node can be both assigned and sender
						currentOrder.State = nodeConfig.Order_Ack
					} else {
						currentOrder.State = nodeConfig.Order_Confirmed
						currentOrder.Acks = append(currentOrder.Acks, id)
					}
				case currentOrder.SenderId:
					currentOrder.State = nodeConfig.Order_Ack
				default:
					currentOrder.State = nodeConfig.Order_Confirmed
					currentOrder.Acks = append(currentOrder.Acks, id)
				}
				// fmt.Printf("Ordertype: %d\n", currentOrder.State)
				orderTx <- currentOrder
			case nodeConfig.Order_Ack:
				switch id {
				case currentOrder.SenderId:
					currentOrder.Acks = getUpdatedAckList2(id, currentOrder.Acks, order.Acks)
					// fmt.Printf("Current order acks: %s\n", currentOrder.Acks)
					if len(currentOrder.Acks) >= enoughAcks {
						fmt.Println("Enough acks")
						currentOrder.State = nodeConfig.Order_Confirmed
						//Send to orderClearer
					} else {
						currentOrder.State = nodeConfig.Order_Ack
						TimeOfButtonPress := currentOrder.Timestamp
						if time.Since(TimeOfButtonPress) > time.Duration(ORDER_SEND_TIMEOUT_MS)*time.Millisecond {
							fmt.Printf("[%s]: ", time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
							fmt.Println("Order timeout reached - clearing order")
							// currentOrder = ClearOrder()
							order.State = nodeConfig.Order_Cleared
						} else {
							orderTx <- currentOrder
						}
					}
				}
			case nodeConfig.Order_Confirmed:
				switch id {
				case currentOrder.SenderId:
					// fmt.Println("Sender order in confirmed state")
				default:
					// fmt.Printf("Order.state in confirmed current order: %d\n", order.State)
					if order.State == nodeConfig.Order_Ack {
						// fmt.Println("Sending ack back from order confirmed!")
						currentOrder.Acks = getUpdatedAckList2(id, currentOrder.Acks, order.Acks)
						fmt.Printf("Current order acks from confirmed: %s\n", currentOrder.Acks)
						// fmt.Printf("Current acks: %s\n", currentOrder.Acks)
						orderTx <- currentOrder
					}
				}
			}

			// lastState = currentOrder.State
			localOrders[order.Request.Floor][order.Request.Button].State = currentState
			localOrders[order.Request.Floor][order.Request.Button] = currentOrder
			// fmt.Printf("[%s]: ", time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
			// fmt.Println("SOIJDIOASJDAOISD")
			time.Sleep(100 * time.Millisecond)
		case <-time.After(50 * time.Millisecond):

		}

		// floor++;

		if btn == len(localOrders[1])-1 {
			btn = 0
			flr++
		} else {
			btn++
		}
		if flr == len(localOrders) {
			flr = 0
			btn = 0
		}
		// btn++
		// if floor == len(localOrders)
		order := localOrders[flr][btn]
		// fmt.Printf("Floor: %d\n", floor)
		// fmt.Printf("Button: %d\n", btn)

		switch order.State {
		case nodeConfig.Order_Ack:
			// TimeOfButtonPress := order.Timestamp
			// if time.Since(TimeOfButtonPress) > time.Duration(ORDER_SEND_TIMEOUT_MS)*time.Millisecond {
			// 	fmt.Printf("[%s]", time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
			// 	fmt.Println("Order timeout reached - clearing order")
			// 	// currentOrder = ClearOrder()
			// 	order.State = nodeConfig.Order_Cleared
			// }
		case nodeConfig.Order_Confirmed:
			node, _, err := peers.GetNodeWithId(order.AssignedId)
			if err == 0 {
				// fmt.Printf("req: %d\n", node.Elevator.Requests[floor][btn])
				// fmt.Printf("Currently assigned id: %s\n",order.AssignedId)
				// fmt.Printf("Currently floor: %d\n\n",node.Elevator.Floor)

				var orderEvent nodeConfig.OrderEvent
				orderEvent.Request.Floor = flr
				orderEvent.Request.Button = elevConfig.ToButtonType(btn)
				orderEvent.Confirmed = true
				orderUpdate <- orderEvent

				TimeOfButtonPress := order.Timestamp
				var orderReassignTimeout int64 = 0
				if order.Cost > 0 {
					orderReassignTimeout = order.Cost
				} else {
					orderReassignTimeout = 7
				}
				if time.Since(TimeOfButtonPress) > time.Duration(2000)*time.Millisecond && time.Since(TimeOfButtonPress) < time.Duration(orderReassignTimeout*3500)*time.Millisecond {
					// if node.Available && node.Elevator.Requests[flr][btn] == 0 {
					// 	order.State = nodeConfig.Order_Cleared
					// 	fmt.Printf("[%s]: ", time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
					// 	fmt.Printf("Cleared order at floor %d, btn: %d \n", flr, btn)
					// }
					if node.Available && node.Elevator.Requests[flr][btn] == 0 {
						order.State = nodeConfig.Order_Cleared
						fmt.Printf("[%s]: ", time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
						fmt.Printf("Cleared order at floor %d, btn: %d \n", flr, btn)
						var orderEvent nodeConfig.OrderEvent
						orderEvent.Request.Floor = flr
						orderEvent.Request.Button = elevConfig.ToButtonType(btn)
						orderEvent.Confirmed = false
						orderUpdate <- orderEvent
					}
				}
				if time.Since(TimeOfButtonPress) > time.Duration(orderReassignTimeout*1000)*time.Millisecond {
					fmt.Printf("[%s]: ", time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
					fmt.Printf("Reassigned order at floor %d, btn: %d \n", flr, btn)
					reassignCh <- order.Request
					order.State = nodeConfig.Order_Cleared

				}
			}
		default:
		}
		localOrders[flr][btn] = order

		// if order.State == nodeConfig.Order_Confirmed {
		// 	// fmt.Printf("Got confirmed order at floor %d, button %d\n", floor, button)
		// 	node, _, err := peers.GetNodeWithId(order.AssignedId)
		// 	if err == 0 {
		// 		// nodeConfig.KnownNodes[index]
		// 		fmt.Printf("req: %d\n", node.Elevator.Requests[floor][btn])
		// 		// fmt.Printf("Floor: %d\n", node.Elevator.Floor)
		// 		// time.Sleep(50 * time.Millisecond)
		// 		TimeOfButtonPress := localOrders[floor][btn].Timestamp
		// 		// if inTimeSpan(time.Time(500)*time.Millisecond, time.Time(1500)*time.Millisecond, time.Since(TimeOfButtonPress)) {
		// 		// 	order.State = nodeConfig.Order_Cleared
		// 		// 	fmt.Printf("Cleared order at floor %d, btn: %d \n", floor, btn)
		// 		// } else if time.Since(TimeOfButtonPress) > time.Duration(1500)*time.Millisecond {
		// 		// 	fmt.Printf("Reassigned order at floor %d, btn: %d \n", floor, btn)
		// 		// 	reassignCh<-order.Request
		// 		// 	order.State = nodeConfig.Order_Cleared
		// 		// }
		// 		if time.Since(TimeOfButtonPress) > time.Duration(500)*time.Millisecond && time.Since(TimeOfButtonPress) < time.Duration(1500)*time.Millisecond {
		// 			if node.Available && node.Elevator.Requests[floor][btn] == 0 {
		// 				order.State = nodeConfig.Order_Cleared
		// 				fmt.Printf("Cleared order at floor %d, btn: %d \n", floor, btn)
		// 			}
		// 		}
		// 		if time.Since(TimeOfButtonPress) > time.Duration(1500)*time.Millisecond {
		// 			fmt.Printf("Reassigned order at floor %d, btn: %d \n", floor, btn)
		// 			reassignCh<-order.Request
		// 			order.State = nodeConfig.Order_Cleared
		// 		}

		// 		// 	if node.Available && node.Elevator.Requests[floor][btn] == 0 {

		// 		// 	}
		// 		// } else if time.Since(TimeOfButtonPress) > time.Duration(1500)*time.Millisecond {
		// 		// 	fmt.Printf("Reassigned order at floor %d, btn: %d \n", floor, btn)
		// 		// 	reassignCh<-order.Request
		// 		// 	order.State = nodeConfig.Order_Cleared
		// 		// }
		// 		// if time.Since(TimeOfButtonPress) > time.Duration(15)*time.Second {

		// 		// }

		// 		// if node.Available && node.Elevator.Requests[floor][button] == 0 && prevOrder[floor][button] == 1 {
		// 		// 	order.State = nodeConfig.Order_Cleared
		// 		// 	fmt.Println("Cleared order")
		// 		// }
		// 	}
		// 	localOrders[floor][btn] = order
		// }
		// for floor, floors := range localOrders {
		// 	for button, _ := range floors {
		// 		order := localOrders[floor][button]
		// 		// var node *nodeConfig.Node
		// 		if order.State == nodeConfig.Order_Confirmed {
		// 			// fmt.Printf("Got confirmed order at floor %d, button %d\n", floor, button)
		// 			node, _, err := peers.GetNodeWithId(order.AssignedId)
		// 			if err == 0 {
		// 				// nodeConfig.KnownNodes[index]
		// 				// fmt.Printf("req: %d\n", node.Elevator.Requests[floor][button])
		// 				// fmt.Printf("Floor: %d\n", node.Elevator.Floor)
		// 				// time.Sleep(50 * time.Millisecond)
		// 				TimeOfButtonPress := localOrders[floor][button].Timestamp
		// 				if time.Since(TimeOfButtonPress) > time.Duration(100)*time.Millisecond {

		// 					if node.Available && node.Elevator.Requests[floor][button] == 0 {
		// 						order.State = nodeConfig.Order_Cleared
		// 						fmt.Println("Cleared order")
		// 					}
		// 				}
		// 				// if time.Since(TimeOfButtonPress) > time.Duration(order.Cost*1.5)*time.Second {

		// 				// }
		// 				// if time.Since(TimeOfButtonPress) > time.Duration(15)*time.Second {

		// 				// }

		// 				// if node.Available && node.Elevator.Requests[floor][button] == 0 && prevOrder[floor][button] == 1 {
		// 				// 	order.State = nodeConfig.Order_Cleared
		// 				// 	fmt.Println("Cleared order")
		// 				// }
		// 				prevOrder[floor][button] = node.Elevator.Requests[floor][button]
		// 			}
		// 			localOrders[floor][button] = order
		// 		}

		// 	}
		// }
		// for floor, floors := range localOrders {
		// 	for button, _ := range floors {
		// 		orderStage := localOrders[floor][button].Stage
		// 		if orderStage == nodeConfig.Order_Confirmed {
		// 			localOrders[floor][button].Stage = nodeConfig.Order_Cleared
		// 		}
		// 		if orderStage != nodeConfig.Order_Cleared {
		// 			TimeOfButtonPress := localOrders[floor][button].Timeout
		// 			fmt.Println(TimeOfButtonPress)
		// 			if time.Since(TimeOfButtonPress) > time.Duration(TIMEOUT)*time.Millisecond {
		// 				localOrders[floor][button].Stage = nodeConfig.Order_Cleared
		// 				fmt.Println("Order cleared")
		// 			}
		// 		}

		// 		fmt.Println(localOrders[floor][button])
		// 	}
		// }
	}
}

// func orderClearer(newOrder chan nodeConfig.Order, clearOrderCh, redistributeCh <-chan nodeConfig.Order) {

// 	for {
// 	select {
// 		case
// 	}
// 	for floor, floors := range localOrders {
// 		for button, _ := range floors {
// 			order := localOrders[floor][button]
// 			// var node *nodeConfig.Node
// 			if order.State == nodeConfig.Order_Confirmed {
// 				// fmt.Printf("Got confirmed order at floor %d, button %d\n", floor, button)
// 				node, _, err := peers.GetNodeWithId(order.AssignedId)
// 				if err == 0 {
// 					// nodeConfig.KnownNodes[index]
// 					// fmt.Printf("req: %d\n", node.Elevator.Requests[floor][button])
// 					// fmt.Printf("Floor: %d\n", node.Elevator.Floor)
// 					// time.Sleep(50 * time.Millisecond)
// 					TimeOfButtonPress := localOrders[floor][button].Timestamp
// 					if time.Since(TimeOfButtonPress) > time.Duration(100)*time.Millisecond {

// 						if node.Available && node.Elevator.Requests[floor][button] == 0 {
// 							order.State = nodeConfig.Order_Cleared
// 							fmt.Println("Cleared order")
// 						}
// 					}
// 					if time.Since(TimeOfButtonPress) > time.Duration(order.Cost*1.5)*time.Second {

// 					}
// 					// if time.Since(TimeOfButtonPress) > time.Duration(15)*time.Second {

// 					// }

// 					// if node.Available && node.Elevator.Requests[floor][button] == 0 && prevOrder[floor][button] == 1 {
// 					// 	order.State = nodeConfig.Order_Cleared
// 					// 	fmt.Println("Cleared order")
// 					// }
// 					prevOrder[floor][button] = node.Elevator.Requests[floor][button]
// 				}
// 				localOrders[floor][button] = order
// 			}

// 		}
// 	}
// 	}
// }

// func MakeNewOrderList()  {
// 	OrderList = make([][2]nodeConfig.Node, elevConfig.NumFloors)
// }

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}
