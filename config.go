package proxy_client

import (
	"net/http"
)

// Config provides settings for client proxy
type Config struct {
	ProxyURL         string
	MaxConn          int
	HandshakeTimeout int
	ClientTimeOut    int
	LogRequest       func(req *http.Request)
	LogResponse      func(resp *http.Response)
}
