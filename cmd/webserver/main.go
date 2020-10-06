package main

import (
	"code.wolfmud.org/WolfMUD.git/comms"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/stats"
	"code.wolfmud.org/WolfMUD.git/zones"
)

func main() {
	stats.Start()
	zones.Load()
	comms.Listen(config.Server.Host, config.Server.Port)
}