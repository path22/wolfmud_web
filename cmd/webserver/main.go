package main

import (
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
	webSrv := server.New("127.0.0.1", "8080")
	webSrv.Run()
}
