package main

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"code.wolfmud.org/WolfMUD.git/comms"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/stats"
	"code.wolfmud.org/WolfMUD.git/zones"

	"github.com/path22/wolfmud_web/pkg/server"
)

func main() {
	go RunTCPServer()
	RunWebServer()
}

func RunTCPServer() {
	stats.Start()
	zones.Load()
	comms.Listen(config.Server.Host, config.Server.Port)
}

func RunWebServer() {
	time.Sleep(time.Second * 5)
	tcpConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port))
	if err != nil {
		panic(err)
	}
	webSrv, err := server.New("127.0.0.1", "8080")
	_ = webSrv
	for {
		tcpConn.Write([]byte("\n"))
		message, _ := bufio.NewReader(tcpConn).ReadString('\n')
		fmt.Println(message)
		time.Sleep(time.Second * 5)
	}
}
