package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"code.wolfmud.org/WolfMUD.git/config"
)

type tcpClient struct {
	tcpConn net.Conn
}

func NewTCPClient() *tcpClient {
	tcpConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port))
	if err != nil {
		panic(err)
	}
	return &tcpClient{
		tcpConn: tcpConn,
	}
}

func (tc *tcpClient) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path[1:]
	tc.tcpConn.Write([]byte(message))
	answer, err := ioutil.ReadAll(tc.tcpConn)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	w.Write(answer)
}
