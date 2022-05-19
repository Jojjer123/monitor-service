package deviceMonitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/onosproject/monitor-service/pkg/logger"
	"github.com/onosproject/monitor-service/pkg/types"
	gclient "github.com/openconfig/gnmi/client/gnmi"
)

// TODO: Place file in new folder representing its own module???
// TODO: Split this file into at least one more, for some helpers.

func DeviceMonitor(monitor types.DeviceMonitor) {
	var counterWaitGroup sync.WaitGroup
	var counterChannels []chan string

	for index, req := range monitor.Requests {
		counterWaitGroup.Add(1)
		counterChannels = append(counterChannels, make(chan string))

		go newCounter(req, monitor.DeviceName, monitor.Target, monitor.Adapter, &counterWaitGroup, counterChannels[index])
	}

	alive := true
	for alive {
		cmd := <-monitor.ManagerChannel
		if cmd == "shutdown" {
			for _, ch := range counterChannels {
				ch <- cmd
			}
			alive = false
		} else if cmd == "update" {
			for _, ch := range counterChannels {
				ch <- "shutdown"
			}

			monitor.Requests = <-monitor.RequestsChannel

			for index, req := range monitor.Requests {
				counterWaitGroup.Add(1)
				counterChannels = append(counterChannels, make(chan string))
				go newCounter(req, monitor.DeviceName, monitor.Target, monitor.Adapter, &counterWaitGroup, counterChannels[index])
			}
		}
	}

	counterWaitGroup.Wait()
}

// Requests counters at the given interval, extract response and forward it.
func newCounter(req types.Request, deviceName string, target string, adapter types.Adapter, waitGroup *sync.WaitGroup, counterChannel <-chan string) {
	defer waitGroup.Done()

	ctx := context.Background()

	c, err := createGnmiClient(adapter, target, ctx)
	if err != nil {
		// Restarts process after 10s, however, if the shutdown command is sent on
		// counterChannel, the process will stop.
		select {
		case <-time.After(10 * time.Second):
			waitGroup.Add(1)
			go newCounter(req, deviceName, target, adapter, waitGroup, counterChannel)
			return
		case msg := <-counterChannel:
			if msg == "shutdown" {
				logger.Info("Exits counter now")
				return
			}
		}
	}

	fmt.Printf("Get %v from switch %v: %v\n", req.Counters[0].Name, deviceName, time.Now().UnixNano())

	// Get the counter and send it to the data processing and to possible subscribers.
	response, err := c.(*gclient.Client).Get(ctx, req.GnmiRequest)

	fmt.Printf("Received %v from switch %v: %v\n", req.Counters[0].Name, deviceName, time.Now().UnixNano())

	if err != nil {
		logger.Errorf("Target returned RPC error: %v", err)
	} else {
		extractData(response, req.GnmiRequest, deviceName)
	}

	// Start a ticker which will trigger repeatedly after (interval) milliseconds.
	intervalTicker := time.NewTicker(time.Duration(req.Interval) * time.Millisecond)

	counterIsActive := true
	for counterIsActive {
		select {
		case msg := <-counterChannel:
			if msg == "shutdown" {
				intervalTicker.Stop()
				counterIsActive = false
			}
		case <-intervalTicker.C:

			fmt.Printf("Get %v from switch %v: %v\n", req.Counters[0].Name, deviceName, time.Now().UnixNano())

			// Get the counter and send it to the data processing and to possible subscribers.
			response, err := c.(*gclient.Client).Get(ctx, req.GnmiRequest)

			fmt.Printf("Received %v from switch %v: %v\n", req.Counters[0].Name, deviceName, time.Now().UnixNano())

			if err != nil {
				logger.Errorf("Target returned RPC error: %v", err)
			} else {
				extractData(response, req.GnmiRequest, deviceName)
			}
		}
	}

	logger.Info("Exits counter now")
}