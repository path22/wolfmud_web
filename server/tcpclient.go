package server

import "net/http"

type tcpClient struct {
}

func NewTCPClient() *tcpClient {
	return &tcpClient{}
}

func (tc *tcpClient) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path[1:]
	w.Write([]byte(message))
}
