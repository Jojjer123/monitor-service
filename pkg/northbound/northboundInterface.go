package northboundInterface

import (
	"fmt"
	"sync"
	"time"

	types "github.com/onosproject/device-monitor/pkg/types"
)

func Northbound(waitGroup *sync.WaitGroup, adminChannel chan types.AdminChannelMessage) {
	fmt.Println("AdminInterface started")
	defer waitGroup.Done()

	// var serverWaitGroup sync.WaitGroup

	// var registerFunction func(chan string, *sync.WaitGroup)

	// select {
	// case x := <-adminChannel:
	// 	{
	// 		registerFunction = x.RegisterFunction
	// 	}
	// }

	go startServer(":11161")

	// // Starts the gRPC server which will be the external interface.
	// go startServer(&serverWaitGroup, registerFunction)

	// Wait for the gNMI server to exit before exiting admin interface.
	// serverWaitGroup.Wait()

	for {
		time.Sleep(10 * time.Second)
	}
}