package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

var connections = make(map[string]*net.Conn)

type tcpClient struct {
	tcpConn net.Conn
}

func NewTCPClient() *tcpClient {
	return &tcpClient{}
}

func (tc *tcpClient) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	connection := r.FormValue("connection")
	message := r.FormValue("cmd")
	fmt.Println("Command", connection, message, connections)
	tcpConn, ok := connections[connection]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	conn := *tcpConn
	_, err := conn.Write([]byte(message + "\n"))
	if err != nil {
		encoder.Encode(map[string]interface{}{
			"err": err.Error(),
		})
	}
	var answer string
	rdr := bufio.NewReader(conn)
	conn.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
	func() {
		defer func() {
			err := recover().(error)
			if !strings.Contains(err.Error(), "timeout") {
				panic(err)
			}
		}()
		for {
			line, err := rdr.ReadString('\n')
			if err != nil {
				panic(err)
			}
			answer += line
		}
	}()
	encoder.Encode(map[string]interface{}{
		"ans": string(answer),
	})
}
