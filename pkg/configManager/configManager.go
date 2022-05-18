package configManager

import (
	"github.com/onosproject/monitor-service/pkg/deviceMonitor"
	"github.com/onosproject/monitor-service/pkg/logger"
	reqBuilder "github.com/onosproject/monitor-service/pkg/requestBuilder"
	"github.com/onosproject/monitor-service/pkg/types"
)

var deviceMonitorStore []types.DeviceMonitor

func ExecuteAdminSetCmd(cmd string, target string, configIndex ...int) string {
	if len(configIndex) > 1 {
		logger.Warn("Config index should not be an array larger than 1")
	}

	switch cmd {
	case "Create":
		// Get slice of the different paths with their intervals and the appropriate adapter if one is necessary
		// Should create new object with all the data inside.
		requests, adapter, deviceName := reqBuilder.GetConfig(target, configIndex[0])
		if len(requests) == 0 {
			return "No configurations to monitor"
		}
		createDeviceMonitor(requests, adapter, target, deviceName)
	case "Update":
		requests, _, _ := reqBuilder.GetConfig(target, configIndex[0])
		if len(requests) == 0 {
			return "No configurations to monitor"
		}
		updateDeviceMonitor(requests, target)
	case "Delete":
		deleteDeviceMonitor(target)
	default:
		logger.Warnf("Could not find command: %v", cmd)
		return "Could not find command: " + cmd
	}

	return "Successfully executed command"
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

	logger.Warn("Could not find device monitor in store")
}

func updateDeviceMonitor(requests []types.Request, target string) {
	for _, monitor := range deviceMonitorStore {
		if monitor.Target == target {
			monitor.ManagerChannel <- "update"
			monitor.RequestsChannel <- requests
			return
		}
	}

	logger.Warn("Could not find device monitor in store")
}

func createDeviceMonitor(requests []types.Request, adapter types.Adapter, target string, deviceName string) {
	// Consider checking Requests to update only if changed.
	monitor := types.DeviceMonitor{
		DeviceName:      deviceName,
		Target:          target,
		Adapter:         adapter,
		Requests:        requests,
		RequestsChannel: make(chan []types.Request),
		ManagerChannel:  make(chan string),
	}

	deviceMonitorStore = append(deviceMonitorStore, monitor)

	go deviceMonitor.DeviceMonitor(monitor)
}
