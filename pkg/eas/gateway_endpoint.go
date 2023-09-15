package eas

import "strings"

type gatewayEndpoint struct {
	domain string
}

// newGatewayEndpoint returns an instance of gatewayEndpoint
func newGatewayEndpoint(domain string) *gatewayEndpoint {
	domain = strings.Replace(domain, "http://", "", 1)
	domain = strings.Replace(domain, "https://", "", 1)
	if domain[len(domain)-1] == '/' {
		domain = domain[:len(domain)-1]
	}

	return &gatewayEndpoint{
		domain: domain,
	}
}

// TryNext always returns the gateway's domain endpoint
func (g *gatewayEndpoint) TryNext(addr string) string {
	return g.domain
}

// sync does nothing for gateway endpoint
func (g *gatewayEndpoint) Sync() {
	return
}
