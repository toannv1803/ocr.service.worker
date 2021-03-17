package module

import (
	"net"
	"net/http"
	"time"
)

func CreateClient() *http.Client {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Second,
		}).Dial,
		TLSHandshakeTimeout: time.Second,
		MaxIdleConns:        100,
		MaxConnsPerHost:     100,
		MaxIdleConnsPerHost: 100,
	}
	return &http.Client{
		Timeout:   40 * time.Second,
		Transport: netTransport,
	}
}
