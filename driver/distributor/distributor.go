package distributor

import (
	nodeConfig "driver/config"
	elevConfig "elevator/config"
	"fmt"
	"network/bcast"
	"network/peers"
	"time"
)

func Distribute(id string, orderCh chan nodeConfig.Order, reassignCh chan elevConfig.ButtonEvent, orderOut chan nodeConfig.Order, peerUpdateCh chan peers.PeerUpdate, orderUpdate chan nodeConfig.OrderEvent, trackConfirmedNode, orderCleared chan nodeConfig.Order) {

	orderTx := make(chan nodeConfig.Order)
	orderRx := make(chan nodeConfig.Order)
	go bcast.Transmitter(15648, orderTx)
	go bcast.Receiver(15648, 0, orderRx)

	const ORDER_SEND_TIMEOUT_MS = 1500

	var localOrders [4][2]nodeConfig.Order

	var availableNodes []string
	enoughAcks := 1

	for floor, floors := range localOrders {
		for button, _ := range floors {
			localOrders[floor][button].State = nodeConfig.Order_Cleared
		}
	}

	for {
		select {
		case order := <-orderCh:
			orderTx <- order
		case p := <-peerUpdateCh:
			fmt.Println("Got peer update:")
			fmt.Printf(" Peers: %q\n", p.Peers)
			fmt.Printf(" New: %q\n", p.New)
			availableNodes = p.Peers
			enoughAcks = len(availableNodes)

		case order := <-orderRx:
			currentOrder := localOrders[order.Request.Floor][order.Request.Button]
			currentState := currentOrder.State
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
				currentOrder.Cost = order.Cost

				switch id {
				case currentOrder.AssignedId:
					orderOut <- currentOrder
					if id == currentOrder.SenderId { // Node can be both assigned and sender
						currentOrder.State = nodeConfig.Order_Ack
					} else {
						currentOrder.State = nodeConfig.Order_Confirmed
						currentOrder.Acks = append(currentOrder.Acks, id)
						trackConfirmedNode <- currentOrder
					}
				case currentOrder.SenderId:
					currentOrder.State = nodeConfig.Order_Ack
				default:
					currentOrder.State = nodeConfig.Order_Confirmed
					currentOrder.Acks = append(currentOrder.Acks, id)
					trackConfirmedNode <- currentOrder
				}
				orderTx <- currentOrder
			case nodeConfig.Order_Ack:
				switch id {
				case currentOrder.SenderId:
					currentOrder.Acks = getUpdatedAckList(id, currentOrder.Acks, order.Acks)
					if len(currentOrder.Acks) >= enoughAcks {
						fmt.Printf("Got enough acks: %s. Confirming order.\n", currentOrder.Acks)
						currentOrder.State = nodeConfig.Order_Confirmed
						trackConfirmedNode <- currentOrder
					} else {
						currentOrder.State = nodeConfig.Order_Ack
						TimeOfButtonPress := currentOrder.Timestamp
						if time.Since(TimeOfButtonPress) > time.Duration(ORDER_SEND_TIMEOUT_MS)*time.Millisecond {
							fmt.Println(time.Since(TimeOfButtonPress))
							fmt.Printf("[%s]: ", time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
							fmt.Println("Order timeout reached - clearing order")
							currentOrder.State = nodeConfig.Order_Cleared
							currentOrder.Acks = nil
						} else {
							orderTx <- currentOrder
						}
					}
				}
			case nodeConfig.Order_Confirmed:
				switch id {
				case currentOrder.SenderId:
				default:
					if order.State == nodeConfig.Order_Ack {
						currentOrder.Acks = getUpdatedAckList(id, currentOrder.Acks, order.Acks)
						orderTx <- currentOrder
					}
				}
			}

			localOrders[order.Request.Floor][order.Request.Button].State = currentState
			localOrders[order.Request.Floor][order.Request.Button] = currentOrder

			time.Sleep(100 * time.Millisecond)
		case order := <-orderCleared:
			btn := order.Request.Button
			flr := order.Request.Floor
			localOrders[flr][btn].State = order.State

		}

	}
}
func updateOrderToTrack(order nodeConfig.Order) nodeConfig.OrderUpdate {
	var trackOrder nodeConfig.OrderUpdate
	trackOrder.AssignedId = order.AssignedId
	trackOrder.Request.Button = order.Request.Button
	trackOrder.Request.Floor = order.Request.Floor
	trackOrder.Timestamp = order.Timestamp
	trackOrder.State = order.State
	return trackOrder
}

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}

func TrackOrders(newOrderToTrack, orderCleared chan nodeConfig.Order, confirmedOrder chan nodeConfig.OrderEvent, reassignCh chan elevConfig.ButtonEvent) {
	var confirmedOrders = make([]nodeConfig.Order, 0)
	iterator := 0
	for {
		select {
		case order := <-newOrderToTrack:
			confirmedOrders = append(confirmedOrders, order)
		case <-time.After(50 * time.Millisecond):
		}
		if len(confirmedOrders) > 0 {
			if iterator < len(confirmedOrders)-1 {
				iterator++
			} else {
				iterator = 0
			}

			order := confirmedOrders[iterator]
			orderUpdated := false
			orderReassigned := false

			switch order.State {
			case nodeConfig.Order_Ack:
			case nodeConfig.Order_Confirmed:
				flr := order.Request.Floor
				btn := order.Request.Button
				TimeOfButtonPress := order.Timestamp

				var reassignTime int64 = getReassignmentTimeout(order)

				nodeConfig.KnownNodesMutex.RLock()
				node := nodeConfig.KnownNodesTable[order.AssignedId]
				nodeConfig.KnownNodesMutex.RUnlock()
				if node != nil {
					order.NumRequests++
					order.SumRequests += float64(node.Elevator.Requests[flr][btn])
					order.AvgRequest = order.SumRequests / order.NumRequests
					confirmedOrder <- makeOrderEvent(flr, btn, true) // TODO: Should only be set once

					if time.Since(TimeOfButtonPress) > time.Duration(5000)*time.Millisecond {

						//Each node sends its own data - if request is zero at floor, it should be executed
						// if node.Available && node.Elevator.Requests[flr][btn] == 0 {
						if node.Available && order.AvgRequest <= 0.5 {

							order.State = nodeConfig.Order_Cleared
							orderUpdated = true
							fmt.Printf("[%s]: ", time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
							fmt.Printf("Cleared order at floor %d, btn: %d \n", flr, btn)
							confirmedOrder <- makeOrderEvent(flr, btn, false)
						}

					}
				}

				if time.Since(TimeOfButtonPress) > time.Duration(reassignTime*1000)*time.Millisecond {
					fmt.Printf("[%s]: ", time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
					fmt.Printf("Reassigned order at floor %d, btn: %d \n", flr, btn)
					orderReassigned = true
					order.State = nodeConfig.Order_Cleared
					orderUpdated = true
				}
			}

			if orderUpdated {
				order.State = nodeConfig.Order_Cleared
				confirmedOrders[iterator] = order
				clearedOrder := confirmedOrders[iterator]
				confirmedOrders = remove(confirmedOrders, iterator)
				orderCleared <- clearedOrder
			}
			if orderReassigned {
				order.State = nodeConfig.Order_Cleared
				confirmedOrders[iterator] = order
				confirmedOrders = remove(confirmedOrders, iterator)
				reassignCh <- order.Request
			}
			time.Sleep(150 * time.Millisecond)
		}
	}
}

func getReassignmentTimeout(order nodeConfig.Order) int64 {
	var orderReassignTimeout int64 = 0
	if order.Cost > 0 {
		orderReassignTimeout = order.Cost
	} else {
		orderReassignTimeout = 7
	}
	return orderReassignTimeout
}

func makeOrderEvent(flr int, btn elevConfig.ButtonType, confirmed bool) nodeConfig.OrderEvent {
	var orderEvent nodeConfig.OrderEvent
	orderEvent.Request.Floor = flr
	orderEvent.Request.Button = btn
	orderEvent.Confirmed = confirmed
	return orderEvent
}

func remove(s []nodeConfig.Order, i int) []nodeConfig.Order {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]

}
func contains(slice []string, value string) bool {
	for _, element := range slice {
		if element == value {
			return true
		}
	}
	return false
}

func getUpdatedAckList(id string, current, new []string) []string {
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
