package main

import (
	"sync"

	northboundInterface "github.com/onosproject/monitor-service/pkg/northbound"
)

// Starts Northbound server of the monitor-service
func main() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	go northboundInterface.Northbound(&waitGroup)

	waitGroup.Wait()
}
