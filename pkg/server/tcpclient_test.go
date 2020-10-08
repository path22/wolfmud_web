package server

import (
	"bufio"
	"code.wolfmud.org/WolfMUD.git/comms"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/stats"
	"code.wolfmud.org/WolfMUD.git/zones"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestTCPConnection(t *testing.T) {
	go func() {
		stats.Start()
		zones.Load()
		comms.Listen(config.Server.Host, config.Server.Port)
	}()
	time.Sleep(time.Second * 5)
	tcpConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port))
	if err != nil {
		panic(err)
	}
	tcpConn.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
	message := "test"
	tcpConn.Write([]byte(message + "\n"))
	r := bufio.NewReader(tcpConn)
	var answer string
	func() {
		defer func() {
			err := recover().(error)
			if !strings.Contains(err.Error(), "timeout") {
				panic(err)
			}
		}()
		for {
			str, err := r.ReadString('\n')
			if err != nil {
				panic(err)
			}
			answer += str
		}
	}()
	fmt.Println(answer)
	tcpConn.Close()
}
