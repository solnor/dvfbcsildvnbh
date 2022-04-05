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

const interval = 250 * time.Millisecond
const timeout = 500 * time.Millisecond

func Transmitter(port int, id string, transmitEnable <-chan bool) {

	nodeUpdateTx := make(chan nodeConfig.Node)
	go bcast.Transmitter(port, nodeUpdateTx)
	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(150 * time.Millisecond):
		}
		if enable {

			nodeConfig.KnownNodesMutex.RLock()
			node := nodeConfig.KnownNodesTable[id]
			nodeConfig.KnownNodesMutex.RUnlock()
			if node != nil {
				node.Elevator = fsm.Elevator1
				nodeUpdateTx <- *node
			}
		}
	}

}

func Receiver(port int, thisId string, peerUpdateCh chan<- PeerUpdate, nodeUpdateCh chan<- nodeConfig.Node) {
	// var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	nodeUpdateRx := make(chan nodeConfig.Node)
	go bcast.Receiver(port, 0, nodeUpdateRx)

	for {
		updated := false
		id := ""
		select {
		case nodeUpdate := <-nodeUpdateRx:
			id = nodeUpdate.Id
			nodeConfig.KnownNodesMutex.RLock()
			node := nodeConfig.KnownNodesTable[string(id)]
			nodeConfig.KnownNodesMutex.RUnlock()
			if node != nil {
				if node.Id == thisId {
					nodeConfig.KnownNodesMutex.Lock()
					nodeConfig.KnownNodesTable[id].Elevator = nodeUpdate.Elevator
					nodeConfig.KnownNodesMutex.Unlock()
				} else {
					nodeConfig.KnownNodesMutex.Lock()
					nodeConfig.KnownNodesTable[id].Available = nodeUpdate.Available
					nodeConfig.KnownNodesTable[id].Elevator = nodeUpdate.Elevator
					nodeConfig.KnownNodesMutex.Unlock()
				}
			} else {
				OnNewNode(nodeUpdate)
			}
		case <-time.After(interval - 50*time.Millisecond):
		}

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
		// Each node updates the availability of other nodes based on which nodes can be reached.
		for _, lostNode := range p.Lost {
			nodeConfig.KnownNodesMutex.RLock()
			node := nodeConfig.KnownNodesTable[lostNode]
			nodeConfig.KnownNodesMutex.RUnlock()
			if node != nil {
				if node.Id != thisId {
					nodeConfig.KnownNodesMutex.Lock()
					nodeConfig.KnownNodesTable[lostNode].Available = false
					nodeConfig.KnownNodesMutex.Unlock()
				}
			}
		}
		for _, connNode := range p.Peers {
			nodeConfig.KnownNodesMutex.RLock()
			node := nodeConfig.KnownNodesTable[connNode]
			nodeConfig.KnownNodesMutex.RUnlock()
			if node != nil {
				if node.Id != thisId {
					nodeConfig.KnownNodesMutex.Lock()
					nodeConfig.KnownNodesTable[connNode].Available = true
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
		time.Sleep(150 * time.Millisecond)
	}
}

func OnNewNode(newNode nodeConfig.Node) {
	node := nodeConfig.NewNode(newNode.Id)
	nodeConfig.KnownNodesMutex.Lock()
	nodeConfig.KnownNodesTable[newNode.Id] = &node
	nodeConfig.KnownNodesMutex.Unlock()
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
