package clienthandler

import "net/http"

type PingHandler struct {
}

func (r PingHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
}
