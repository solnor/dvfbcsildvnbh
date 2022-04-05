package assigner

import (
	nodeConfig "driver/config"
	"driver/cost"
	elevConfig "elevator/config"
	"fmt"
	"time"
)

func AssignOrder(id string, requestCh, reassignCh chan elevConfig.ButtonEvent, assignedOrder chan nodeConfig.Order) {

	var calculatedElevator *nodeConfig.Node
	var order nodeConfig.Order
	var request elevConfig.ButtonEvent
	for {
		orderAssigned := false
		select {
		case r := <-requestCh:
			request = r
		case r := <-reassignCh:
			request = r
		}

		var minimumCost int64 = 500
		var elevatorCost int64

		nodeConfig.KnownNodesMutex.RLock()
		for _, node := range nodeConfig.KnownNodesTable {
			eCopy := elevConfig.DupElevator(node.Elevator)
			eCopy.Requests[request.Floor][request.Button] = 1
			elevatorCost = cost.TimeToIdle(eCopy)
			if elevatorCost < minimumCost && node.Available {
				minimumCost = elevatorCost
				calculatedElevator = node
				orderAssigned = true
			}
		}
		nodeConfig.KnownNodesMutex.RUnlock()

		if orderAssigned {
			order = createOrder(id, calculatedElevator.Id, request, minimumCost)
			fmt.Printf("Assigned order to ID: %s\n", order.AssignedId)
			assignedOrder <- order
		}
	}
}

func createOrder(id, assignedId string, request elevConfig.ButtonEvent, cost int64) nodeConfig.Order {
	var order nodeConfig.Order
	order.SenderId = id
	order.AssignedId = assignedId
	order.Request = request
	order.Timestamp = time.Now()
	order.Acks = nil
	order.Acks = append(order.Acks, id)
	order.State = nodeConfig.Order_New
	order.Cost = cost

	return order
}
