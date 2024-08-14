package netutil

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var DeafaultDnsResolver *dnsResolver

var (
	ErrEmptyAddress = errors.New("not found address")
)

func init() {
	DeafaultDnsResolver = newDnsResolver()

	go DeafaultDnsResolver.loopResolve()
}

type hostName struct {
	host  string
	addrs []string
	i     *uint32
}

func newHostName(address string) *hostName {
	h := &hostName{
		host: address,
	}

	val := uint32(0)
	h.i = &val
	return h
}

// only resolve ip4 addr
func (h *hostName) lookupHost() {
	addrs, err := net.LookupHost(h.host)
	if err == nil {
		ip4Addrs := make([]string, 0, len(addrs))
		for _, addr := range addrs {
			if ip := net.ParseIP(addr); ip != nil {
				if ip.To4() != nil {
					ip4Addrs = append(ip4Addrs, addr)
				}
			}
		}

		h.addrs = ip4Addrs
	}
}

func (h *hostName) getSingleAddress() (string, error) {
	l := len(h.addrs)
	if l == 0 {
		return "", ErrEmptyAddress
	}
	if l == 1 {
		return h.addrs[0], nil
	}

	index := int(atomic.LoadUint32(h.i)) % len(h.addrs)
	atomic.AddUint32(h.i, 1)

	return h.addrs[index], nil
}

type dnsResolver struct {
	hosts []*hostName
	lock  sync.Mutex
}

func newDnsResolver() *dnsResolver {
	return &dnsResolver{
		hosts: make([]*hostName, 0, 8),
	}
}

// return address ip by host
// if host first resolve , add to the dnsResolver, return same with the host
// host in dnsResolver, but addrs is empty, return ErrEmptyAddress
func GetAddrByHost(host string) (string, error) {
	if DeafaultDnsResolver.existHost(host) {
		return DeafaultDnsResolver.getAddrByHost(host)
	}

	go func() {
		DeafaultDnsResolver.addHost(host)
	}()

	return host, nil
}

func (d *dnsResolver) getAddrByHost(host string) (string, error) {
	for _, h := range d.hosts {
		if h.host == host {
			return h.getSingleAddress()
		}
	}

	return host, nil
}
func (d *dnsResolver) existHost(host string) bool {
	for _, h := range d.hosts {
		if h.host == host {
			return true
		}
	}
	return false
}
func (d *dnsResolver) addHost(host string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	found := false
	for _, h := range d.hosts {
		if h.host == host {
			found = true
		}
	}

	if !found {
		h := newHostName(host)
		h.lookupHost()
		d.hosts = append(d.hosts, h)
	}
}

func (d *dnsResolver) loopResolve() {

	for {
		for _, h := range d.hosts {
			h.lookupHost()
		}

		time.Sleep(time.Minute)
	}
}
