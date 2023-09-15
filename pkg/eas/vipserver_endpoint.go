package eas

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type vipServerEndpoint struct {
	baseEndpoint
	domain string
	client http.Client
}

// newVipServerEndpoint returns an instance for vipServerEndpoint
func newVipServerEndpoint(domain string) *vipServerEndpoint {
	domain = strings.Replace(domain, "http://", "", 1)
	domain = strings.Replace(domain, "https://", "", 1)
	if domain[len(domain)-1] == '/' {
		domain = domain[:len(domain)-1]
	}

	return &vipServerEndpoint{
		baseEndpoint: *newBaseEndpoint(),
		domain:       domain,
		client:       http.Client{},
	}
}

// getServer randomly gets a server from vipserver server list
func (v *vipServerEndpoint) getServer() (string, error) {
	url := "http://jmenv.tbsite.net:8080/vipserver/serverlist"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("failed to create request for %v: %v\n", url, err)
		return "", err
	}
	resp, err := v.client.Do(req)
	if resp.StatusCode != 200 || err != nil {
		fmt.Printf("failed to query %v: %v\n", url, err)
		return resp.Status, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("failed to read vipserver server list from %v: %v, %v, %v", url, resp.StatusCode, body, err)
		return resp.Status, err
	}
	serverList := strings.Split(strings.Trim(string(body[:]), " "), "\n")
	rand.Seed(time.Now().UTC().UnixNano())
	return serverList[rand.Intn(len(serverList)-1)], nil
}

// sync with server, get server list and set endpoints
func (v *vipServerEndpoint) Sync() {
	server, err := v.getServer()
	if err != nil {
		return
	}
	url := fmt.Sprintf("http://%s/vipserver/api/srvIPXT?dom=%s&clusters=DEFAULT", server, v.domain)

	endpoints := make(map[string]int)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("failed to create request for %v: %v\n", url, err)
		return
	}
	resp, err := v.client.Do(req)
	if err != nil {
		fmt.Printf("failed to query %v: %v\n", url, err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("failed to read service endpoints from %v: %v, %v, %v\n", url, resp.Status, body, err)
		return
	}

	mp := make(map[string]interface{})
	json.Unmarshal(body, &mp)
	for _, hostmap := range mp["hosts"].([]interface{}) {
		host := hostmap.(map[string]interface{})
		if host["valid"].(bool) {
			name := fmt.Sprintf("%v:%v", host["ip"], host["port"])
			endpoints[name] = int(host["weight"].(float64))
		}
	}

	v.setEndpoints(endpoints)
}
