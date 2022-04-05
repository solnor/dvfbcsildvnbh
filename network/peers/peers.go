package peers

import (
	nodeConfig "driver/config"
	"elevator/fsm"
	"network/bcast"
	"sort"
	"time"
)

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

const interval = 150 * time.Millisecond
const timeout = 1000 * time.Millisecond

func Transmitter(port int, id string, transmitEnable <-chan bool) {

	// conn := conn.DialBroadcastUDP(port)
	// addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	nodeUpdateTx := make(chan nodeConfig.Node)
	go bcast.Transmitter(port, nodeUpdateTx)

	// thisNode := nodeConfig.NewNode(id)
	// thisNode.Elevator = fsm.Elevator1
	// nodeConfig.KnownNodes = make(map[string])
	// nodeConfig.KnownNodes = append(nodeConfig.KnownNodes, &thisNode) //MOVE THIS
	// fmt.Printf("thisNode: %d\n", thisNode.Elevator.Floor)
	// fmt.Println(fsm.Elevator1.Floor)
	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		// case <-time.After(interval - 50*time.Millisecond):
		case <-time.After(100 * time.Millisecond):
		}
		if enable {
			// nodeConfig.KnownNodes

			// n, _, err := GetNodeWithId(id)
			// fmt.Println(nodeConfig.KnownNodes)
			nodeConfig.KnownNodesMutex.RLock()
			node := nodeConfig.KnownNodesTable[id]
			nodeConfig.KnownNodesMutex.RUnlock()
			if node != nil {
				node.Elevator = fsm.Elevator1
				nodeUpdateTx <- *node
			}
			// if err == 0 {
			// 	n.Elevator = fsm.Elevator1
			// 	// nodeConfig.KnownNodes[index].Elevator = fsm.Elevator1
			// 	// fmt.Printf("souhafdiounha %d\n", n.Elevator.Floor)
			// 	nodeUpdateTx <- *n
			// }
			// conn.WriteTo([]byte(id), addr)
		}
		// time.Sleep(30 * time.Millisecond)
		// fmt.Printf("[%s]: ", time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
		// fmt.Println("PEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEER Transdmit")
	}

}

func Receiver(port int, thisId string, peerUpdateCh chan<- PeerUpdate, nodeUpdateCh chan<- nodeConfig.Node) {
	// var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	nodeUpdateRx := make(chan nodeConfig.Node)
	go bcast.Receiver(port, interval, nodeUpdateRx)

	// conn := conn.DialBroadcastUDP(port)
	for {
		// fmt.Printf("Known nodes: %q\n", nodeConfig.KnownNodes)
		updated := false

		// conn.SetReadDeadline(time.Now().Add(interval))
		// n, _, _ := conn.ReadFrom(buf[0:])
		// var n nodeConfig.Node
		id := ""
		select {
		case nodeUpdate := <-nodeUpdateRx:
			// id = n.Id
			// // fmt.Printf("From peers.Receiver: n.Floor: %d\n", n.Elevator.Floor)
			// nodeIsKnown := false
			// for _, node := range nodeConfig.KnownNodes {
			// 	if node.Id == n.Id {
			// 		nodeIsKnown = true
			// 	}
			// }
			// for _, node := range nodeConfig.KnownNodesTable {
			// 	if node.Id == nodeUpdate.Id {
			// 		nodeIsKnown = true
			// 	}
			// 	break
			// }
			nodeConfig.KnownNodesMutex.RLock()
			node := nodeConfig.KnownNodesTable[id]
			nodeConfig.KnownNodesMutex.RUnlock()
			if node != nil {
				if node.Id != thisId {
					*node = nodeUpdate
				}
			} else {
				OnNewNode2(nodeUpdate)
			}
			// if nodeIsKnown {
			// 	// fmt.Printf("From main.Receiver: id: %s, n.Floor: %d\n", n.Id, n.Elevator.Floor) //n) //n.Elevator.Floor)
			// 	nodeToUpdate, _, _ := GetNodeWithId(n.Id)
			// 	if nodeToUpdate.Id != thisId {
			// 		*nodeToUpdate = n
			// 	}
			// 	// fmt.Printf()

			// } else {
			// 	OnNewNode2(n)
			// }
			// nodeUpdateCh <- n
		case <-time.After(interval - 50*time.Millisecond):
		}
		// id := string(buf[:n])

		// id := "10"
		// Adding new connection
		p.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		// Removing dead connection
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		for _, lostNode := range p.Lost {
			// node, _, _ := GetNodeWithId(lostNode)
			nodeConfig.KnownNodesMutex.Lock()
			nodeConfig.KnownNodesTable[lostNode].Available = false
			nodeConfig.KnownNodesMutex.Unlock()
		}
		for _, peer := range p.Peers {
			// node, _, _ := GetNodeWithId(peer)
			nodeConfig.KnownNodesMutex.RLock()
			node := nodeConfig.KnownNodesTable[peer]
			nodeConfig.KnownNodesMutex.RUnlock()
			if node != nil {
				if node.Id != thisId {
					nodeConfig.KnownNodesMutex.Lock()
					nodeConfig.KnownNodesTable[peer].Available = true
					nodeConfig.KnownNodesMutex.Unlock()
				}
			}
		}

		// Sending update
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)
			peerUpdateCh <- p
		}
		// fmt.Printf("[%s]: ", time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
		// fmt.Println("SOIJDIOASJDAOISD")
		// fmt.Printf("BOTTOM OF PEERS Known nodes: %q\n", nodeConfig.KnownNodes)
		time.Sleep(100 * time.Millisecond)
	}
}

// func Receiver2(port int, thisId string, peerUpdateCh chan<- PeerUpdate, nodeUpdateCh chan<- nodeConfig.Node) {
// 	// var buf [1024]byte
// 	var p PeerUpdate
// 	lastSeen := make(map[string]time.Time)

// 	nodeUpdateRx := make(chan nodeConfig.Node)
// 	go bcast.Receiver(port, interval, nodeUpdateRx)

// 	// conn := conn.DialBroadcastUDP(port)
// 	for {
// 		// fmt.Printf("Known nodes: %q\n", nodeConfig.KnownNodes)
// 		updated := false

// 		// conn.SetReadDeadline(time.Now().Add(interval))
// 		// n, _, _ := conn.ReadFrom(buf[0:])
// 		var n nodeConfig.Node
// 		id := ""
// 		select {
// 		case n = <-nodeUpdateRx:
// 			id = n.Id
// 			// fmt.Printf("From peers.Receiver: n.Floor: %d\n", n.Elevator.Floor)
// 			nodeIsKnown := false
// 			for _, node := range nodeConfig.KnownNodes {
// 				if node.Id == n.Id {
// 					nodeIsKnown = true
// 				}
// 			}
// 			if nodeIsKnown {
// 				// fmt.Printf("From main.Receiver: id: %s, n.Floor: %d\n", n.Id, n.Elevator.Floor) //n) //n.Elevator.Floor)
// 				nodeToUpdate, _, _ := GetNodeWithId(n.Id)
// 				if nodeToUpdate.Id != thisId {
// 					*nodeToUpdate = n
// 				}
// 				// fmt.Printf()

// 			} else {
// 				OnNewNode2(n)
// 			}
// 			// nodeUpdateCh <- n
// 		case <-time.After(interval - 50*time.Millisecond):
// 		}
// 		// id := string(buf[:n])

// 		// id := "10"
// 		// Adding new connection
// 		p.New = ""
// 		if id != "" {
// 			if _, idExists := lastSeen[id]; !idExists {
// 				p.New = id
// 				updated = true
// 			}

// 			lastSeen[id] = time.Now()
// 		}

// 		// Removing dead connection
// 		p.Lost = make([]string, 0)
// 		for k, v := range lastSeen {
// 			if time.Now().Sub(v) > timeout {
// 				updated = true
// 				p.Lost = append(p.Lost, k)
// 				delete(lastSeen, k)
// 			}
// 		}

// 		for _, lostNode := range p.Lost {
// 			node, _, _ := GetNodeWithId(lostNode)
// 			node.Available = false
// 		}
// 		for _, peer := range p.Peers {
// 			node, _, _ := GetNodeWithId(peer)
// 			if node.Id != thisId {
// 				node.Available = true
// 			}
// 		}

// 		// Sending update
// 		if updated {
// 			p.Peers = make([]string, 0, len(lastSeen))

// 			for k, _ := range lastSeen {
// 				p.Peers = append(p.Peers, k)
// 			}

// 			sort.Strings(p.Peers)
// 			sort.Strings(p.Lost)
// 			peerUpdateCh <- p
// 		}
// 		// fmt.Printf("[%s]: ", time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
// 		// fmt.Println("SOIJDIOASJDAOISD")
// 		// fmt.Printf("BOTTOM OF PEERS Known nodes: %q\n", nodeConfig.KnownNodes)
// 		time.Sleep(100 * time.Millisecond)
// 	}
// }

// func OnNewNode(node PeerUpdate) {
// 	for _, newNode := range node.New {
// 		newNodeIsKnown := false
// 		for _, peer := range node.Peers {
// 			if string(newNode) == peer {
// 				newNodeIsKnown = true
// 			}
// 		}
// 		if newNodeIsKnown {
// 			n, _, err := GetNodeWithId(node.New)
// 			if err != 0 {
// 				fmt.Printf("Could not find elevator with id %s\n", node.New)
// 				// In case ID is known, but no elevator is associated with the id: Create new node with ID
// 				n := nodeConfig.NewNode(node.New)
// 				nodeConfig.KnownNodes = append(nodeConfig.KnownNodes, &n)
// 			}
// 			// e.undefined = setNodeDataUndefined(e)
// 			n.Available = true
// 		} else {
// 			n := nodeConfig.NewNode(node.New)
// 			nodeConfig.KnownNodes = append(nodeConfig.KnownNodes, &n)
// 		}
// 		// if node.New not in node.Peers {

// 		// } else {

// 		// }
// 	}
// }

func OnNewNode2(newNode nodeConfig.Node) {

	node := nodeConfig.NewNode(newNode.Id)
	nodeConfig.KnownNodesMutex.Lock()
	nodeConfig.KnownNodesTable[newNode.Id] = &node
	nodeConfig.KnownNodesMutex.Unlock()
	// node := nodeConfig.KnownNodesTable[newNode.Id]

	// nodeConfig.KnownNodes = append(nodeConfig.KnownNodes, &n)

	// for _, newNode := range node.New {
	// 	newNodeIsKnown := false
	// 	for _, peer := range node.Peers {
	// 		if string(newNode) == peer {
	// 			newNodeIsKnown = true
	// 		}
	// 	}
	// 	if newNodeIsKnown {
	// 	n, err := GetNodeWithId(node.New)
	// 	if err != 0 {
	// 		fmt.Printf("Could not find elevator with id %s\n", node.New)
	// 		// In case ID is known, but no elevator is associated with the id: Create new node with ID
	// 		n := nodeConfig.NewNode(node.New)
	// 		nodeConfig.KnownNodes = append(nodeConfig.KnownNodes, n)
	// 	}
	// 	// e.undefined = setNodeDataUndefined(e)
	// 	n.Available = true
	// // } else {
	// 	n := nodeConfig.NewNode(node.New)
	// 	nodeConfig.KnownNodes = append(nodeConfig.KnownNodes, n)
	// }
	// if node.New not in node.Peers {

	// } else {

	// }
	// }
}

// func GetNodeWithId(id string) (*nodeConfig.Node, int, int) {
// 	for i, node := range nodeConfig.KnownNodes {
// 		if id == node.Id {
// 			// fmt.Printf("GetNodeWithId: %s\n", node.Id)
// 			return node, i, 0
// 		}
// 	}
// 	// fmt.Println("Uppppppsouhd")
// 	return nil, 0, 1 //errors.New("Could not find node\n")
// }
