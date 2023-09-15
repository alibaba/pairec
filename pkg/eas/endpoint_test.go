package eas

import (
	"fmt"
	"testing"
)

func TestBaseEndpoint_changed(t *testing.T) {
	be := newBaseEndpoint()

	be.setEndpoints(map[string]int{
		"192.168.1.1": 10,
		"192.168.1.2": 10,
		"192.168.1.3": 10,
		"192.168.1.4": 10,
		"192.168.1.5": 10,
	})

	endpoints := map[string]int{
		"192.168.1.1": 10,
		"192.168.1.2": 10,
		"192.168.1.4": 10,
		"192.168.1.5": 10,
	}

	if !be.changed(endpoints) {
		t.Fatalf("endpoints changed detect failed")
	}

	endpoints = map[string]int{}

	if !be.changed(endpoints) {
		t.Fatalf("endpoints changed detect failed")
	}

	endpoints = map[string]int{
		"192.168.1.2": 10,
		"192.168.1.1": 10,
		"192.168.1.5": 10,
		"192.168.1.3": 10,
		"192.168.1.4": 10,
	}

	if be.changed(endpoints) {
		t.Fatalf("endpoints changed detect failed")
	}
}

func TestBaseEndpoint_TryNext(t *testing.T) {
	be := newBaseEndpoint()

	endpoints := map[string]int{
		"192.168.1.1": 10,
		"192.168.1.2": 10,
		"192.168.1.3": 20,
		"192.168.1.4": 10,
		"192.168.1.5": 30,
	}

	be.setEndpoints(endpoints)
	for name := range endpoints {
		for i := 0; i < 100; i++ {
			endpoint := be.TryNext(name)
			if endpoint == name {
				t.Fatalf("get a legacy endpoint")
			}
		}
	}


	results := make(map[string]int)
	for i := 0; i < 10000; i+=1 {
		endpoint := be.TryNext("")
		if count, ok := results[endpoint]; !ok {
			results[endpoint] = 1
		} else {
			results[endpoint] = count + 1
		}
	}

	if val, exist := results["192.168.1.1"]; !exist || val != 1250 {
		t.Fatalf("wrrr failed")
	}
	if val, exist := results["192.168.1.2"]; !exist || val != 1250 {
		t.Fatalf("wrrr failed")
	}
	if val, exist := results["192.168.1.3"]; !exist || val != 2500 {
		t.Fatalf("wrrr failed")
	}
	if val, exist := results["192.168.1.4"]; !exist || val != 1250 {
		t.Fatalf("wrrr failed")
	}
	if val, exist := results["192.168.1.5"]; !exist || val != 3750 {
		t.Fatalf("wrrr failed")
	}

	fmt.Printf("%v\n", results)
}
