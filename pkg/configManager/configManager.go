package deviceManager

import (
	"fmt"
	"sync"

	reqBuilder "github.com/onosproject/monitor-service/pkg/requestBuilder"
	"github.com/onosproject/monitor-service/pkg/types"
)

var deviceMonitorStore []types.DeviceMonitor

func ConfigManager(waitGroup *sync.WaitGroup, adminChannel chan types.ConfigAdminChannelMessage) {
	defer waitGroup.Done()

	// TODO: Remove deviceMonitorWaitGroup and add better way of keeping module "alive".

	var deviceMonitorWaitGroup sync.WaitGroup

	var adminMessage types.ConfigAdminChannelMessage
	adminMessage.ExecuteSetCmd = executeAdminSetCmd

	adminChannel <- adminMessage

	deviceMonitorWaitGroup.Wait()
}

func executeAdminSetCmd(cmd string, target string, configIndex ...int) string {
	switch cmd {
	case "Create":
		// Get slice of the different paths with their intervals and the appropriate
		// adapter if one is necessary
		requests, adapter := reqBuilder.GetConfig(target, configIndex[0])
		createDeviceMonitor(requests, adapter, target)
	case "Update":
		requests, _ := reqBuilder.GetConfig(target, configIndex[0])
		updateDeviceMonitor(requests, target)
	case "Delete":
		deleteDeviceMonitor(target)
	default:
		fmt.Println("Could not find command: " + cmd)
		return "Command not found!"
	}

	return "Successfully executed command sent"
}

func updateDeviceMonitor(requests []types.Request, target string) {
	for _, monitor := range deviceMonitorStore {
		if monitor.Target == target {
			monitor.ManagerChannel <- "update"
			monitor.RequestsChannel <- requests
			return
		}
	}
}

func deleteDeviceMonitor(target string) {
	for index, monitor := range deviceMonitorStore {
		if monitor.Target == target {
			monitor.ManagerChannel <- "shutdown"

			deviceMonitorStore[index] = deviceMonitorStore[len(deviceMonitorStore)-1]
			deviceMonitorStore = deviceMonitorStore[:len(deviceMonitorStore)-1]
			return
		}
	}
}

func createDeviceMonitor(requests []types.Request, adapter types.Adapter, target string) {
	managerChannel := make(chan string)
	requestsChannel := make(chan []types.Request)

	// Consider checking Requests to update only if changed.
	monitor := types.DeviceMonitor{
		Target:          target,
		Adapter:         adapter,
		Requests:        requests,
		RequestsChannel: requestsChannel,
		ManagerChannel:  managerChannel,
	}

	deviceMonitorStore = append(deviceMonitorStore, monitor)

	// fmt.Println("Starting deviceMonitor now...")
	go deviceMonitor(monitor)
}
