package watchdog

import (
	nodeConfig "driver/config"
	elevConfig "elevator/config"
	"fmt"
	"time"
)

func Watchdog(elev *elevConfig.Elevator) {
	var timer *time.Timer
	timer = time.NewTimer(time.Duration(1 * time.Second))
	timer.Stop()
	const interval = 150
	lastFloor := elev.Floor
	lastBehaviour := elev.Behaviour
	for {
		select {
		case <-timer.C:
			fmt.Printf("Litta bjeffa: %t (fra mini-honnja)\n", nodeConfig.KnownNodes[0].Available)

			node := nodeConfig.KnownNodes[0]
			node.Available = false
			fmt.Printf("Litta bjeffa: %t (fra mini-honnja)\n", nodeConfig.KnownNodes[0].Available)

		case <-time.After(interval * time.Millisecond):
			// fmt.Printf("sdsdsdasd\n")
			// fmt.Println(elev.Behaviour)
		}
		// time.Sleep(250*time.Millisecond)
		if elev.Floor != lastFloor {
			timer.Reset(10 * time.Second)
			nodeConfig.KnownNodes[0].Available = true
			lastFloor = elev.Floor
		} else if elev.Behaviour != lastBehaviour {
			timer.Reset(10 * time.Second)
			nodeConfig.KnownNodes[0].Available = true
			lastBehaviour = elev.Behaviour
		} else if elev.Behaviour == elevConfig.EB_Idle {
			timer.Reset(10 * time.Second)
			nodeConfig.KnownNodes[0].Available = true
		}
		time.Sleep(200)
	}
}
