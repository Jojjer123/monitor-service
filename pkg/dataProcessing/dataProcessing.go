package dataProcessing

import (
	"sync"
)

func DataProcessing(waitGroup *sync.WaitGroup) {
	// fmt.Println("DataProcessing started")
	defer waitGroup.Done()
}
