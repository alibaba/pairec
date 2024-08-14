package aliyun_igraph_go_sdk

import (
	"errors"
	"net"
	"net/url"
)

var (
	localIp   net.IP
	queryIP   string
	configStr = "config{outfmt=json&no_cache=false&cache_only=false}"
)

func init() {
	localIp, _ = LocalIP()
	if localIp != nil {
		queryIP = "ip=127.0.0.1"
	} else {
		queryIP = "ip=" + localIp.String()
	}
}

type ReadRequest struct {
	QueryString string            `json:"query_string"`
	QueryParams map[string]string `json:"query_params"`
}

func (r *ReadRequest) Validate() error {
	if len(r.QueryString) == 0 {
		return InvalidParamsError{"Empty query string"}
	}
	return nil
}

func (r *ReadRequest) AddQueryParam(key string, value string) *ReadRequest {
	r.QueryParams[key] = value
	return r
}

func (r *ReadRequest) SetQueryParams(params map[string]string) *ReadRequest {
	r.QueryParams = params
	return r
}

func (r *ReadRequest) BuildUri() string {
	rawUrl := queryIP + "?" + url.QueryEscape(configStr+"&&"+r.QueryString)
	return rawUrl
}

func LocalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			return ip, nil
		}
	}

	return nil, errors.New("no IP")
}
