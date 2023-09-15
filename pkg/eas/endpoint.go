package eas

import (
	"fmt"
	"sync"
	"time"
)

type Endpoint interface {
	Sync()
	TryNext(addr string) string
}

type baseEndpoint struct {
	lock      sync.RWMutex
	endpoints map[string]int
	scheduler wrrscheduler
}

func newBaseEndpoint() *baseEndpoint {
	return &baseEndpoint{
		lock:      sync.RWMutex{},
		endpoints: make(map[string]int),
		scheduler: wrrscheduler{inited: false},
	}
}

// setEndpoints replaces the local endpoint list, every time we get the response from
// the upstream discovery server, we need to update the endpoint list in local memory.
func (ep *baseEndpoint) setEndpoints(endpoints map[string]int) {
	if !ep.changed(endpoints) {
		return
	}

	ep.lock.Lock()
	ep.endpoints = endpoints
	ep.scheduler = wrrScheduler(ep.endpoints)
	ep.lock.Unlock()
}

// get gets an ip address and port using a weight round robin algorithm from endpoint list in local cache.
func (ep *baseEndpoint) get() string {
	retryCount := 0
	for true {
		if ep.scheduler.inited {
			break
		}
		retryCount += 1
		if retryCount == 3000 { // 30s timeout
			fmt.Printf("Timeout when getting the services' endpoint list from upstream\n")
			return ""
		}
		time.Sleep(10 * time.Millisecond)
	}

	ep.lock.RLock()
	addr := ep.scheduler.getNext()
	ep.lock.RUnlock()
	return addr
}

// changed checks whether the latest endpoints returned from the upstream discovery server are identically
// equal to the endpoints in local memory cache.
func (ep *baseEndpoint) changed(endpoints map[string]int) bool {
	if len(ep.endpoints) != len(endpoints) {
		return true
	}

	for key, val := range endpoints {
		if oldVal, exist := ep.endpoints[key]; !exist || oldVal != val {
			return true
		}
	}
	return false
}

// TryNext tries to get an new endpoint from the local endpoint list cache, if we get a new endpoint equals
// the last failed one, continue to retry until we find a new endpoint not equals to the last one.
func (ep *baseEndpoint) TryNext(addr string) string {
	if len(ep.endpoints) > 1 && len(addr) > 0 {
		newAddr := ep.get()
		i := 0
		for addr == newAddr && i < 100 {
			newAddr = ep.get()
			i += 1
		}
		return newAddr
	}
	return ep.get()
}
