package requestBuilder

import (
	"strings"

	"github.com/onosproject/monitor-service/pkg/logger"
	storageInterface "github.com/onosproject/monitor-service/pkg/storage"
	types "github.com/onosproject/monitor-service/pkg/types"
	"github.com/openconfig/gnmi/proto/gnmi"
)

// Builds requests to send to a switch or an adapter.
func GetRequestConf(target string, configSelected int) ([]types.Request, types.Adapter, string) {
	conf := storageInterface.GetConfig(target)

	if len(conf.Configs) == 0 {
		logger.Error("No configurations to monitor")
		return []types.Request{}, types.Adapter{}, ""
	}
	// TODO: Add check for empty config, and dont crash if that is the case.

	var requests []types.Request

	// For each interval and all counters for that interval (intCounters), build a request.
	for _, intCounters := range conf.Configs[configSelected].Counters {
		request := types.Request{
			Interval: intCounters.Interval,
		}

		// For each counter for an interval, build a Counter object.
		for _, counter := range intCounters.Counters {
			request.Counters = append(request.Counters, types.Counter{
				Name: counter.Name,
				Path: getPathFromString(counter.Path),
			})
		}

		// Create gnmi get request.
		r := &gnmi.GetRequest{
			Type: gnmi.GetRequest_STATE,
		}

		for _, counter := range request.Counters {
			r.Path = append(r.Path, &gnmi.Path{
				Target: target,
				Elem:   counter.Path,
			})
		}

		request.GnmiRequest = r
		requests = append(requests, request)
	}

	var adapter types.Adapter

	// Only protocol without need for an adapter is gNMI, for now.
	if conf.Protocol != "GNMI" {
		adapter = storageInterface.GetAdapter(conf.Protocol)
	} else {
		logger.Info("Support for direct communication with switches over gNMI is not yet supported")
	}

	return requests, adapter, conf.DeviceName
}

// Get gNMI path from a string.
func getPathFromString(path string) []*gnmi.PathElem {
	if !strings.Contains(path, "elem:") {
		return nil
	}

	var pathElems []*gnmi.PathElem
	for index, elem := range strings.Split(path, "elem:") {
		if index == 0 {
			continue
		}

		tok := strings.Split(elem, "'")

		newElem := &gnmi.PathElem{
			Name: tok[1],
		}

		// Contains key.
		if len(tok) > 3 {
			keyMap := make(map[string]string)
			keyMap[tok[3]] = tok[5]
			newElem.Key = keyMap
		}

		pathElems = append(pathElems, newElem)
	}

	return pathElems
}
