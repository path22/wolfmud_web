package server

import (
	"code.wolfmud.org/WolfMUD.git/config"
	"fmt"
	"html/template"
	"math/rand"
	"net"
	"net/http"
	"strconv"
)

type Server struct {
	Host string
	Port string
	srv  http.Server
}

var tmpl, err = template.New("main_page").Parse(`
<html>
<head>
</head>
<body>
	<label>command: <input id="command" type="text" name="command"></label>
	<input id="send" type="button" value="Send">
	<h3 id="answer"></h3>
<script>
(function() {
	var command = document.getElementById("command");
	var send = document.getElementById("send");
	send.onclick = function() {
		fetch('/command?connection={{.ConnectionID}}&cmd='+command.value)
		  .then(response => response.json())
		  .then(data => {
			var ans = document.getElementById("answer");
			ans.innerHTML = data.ans
		  });
	}
}())
</script>
</body>
</html>
`)

func New(host string, port string) *Server {
	s := new(Server)
	s.Host = host
	s.Port = port
	webSrvMux := http.NewServeMux()
	webSrvMux.HandleFunc("/page", func(w http.ResponseWriter, r *http.Request) {
		connection := strconv.Itoa(rand.Int())
		data := map[string]interface{}{
			"ConnectionID": connection,
		}
		tcpConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port))
		if err != nil {
			panic(err)
		}
		connections[connection] = &tcpConn
		tmpl.Execute(w, data)
	})
	tcpClient := NewTCPClient()
	webSrvMux.Handle("/command", tcpClient)
	s.srv = http.Server{
		Handler: webSrvMux,
	}
	return s
}

func (s *Server) Run() {
	webSrvLn, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.Host, s.Port))
	if err != nil {
		panic(err)
	}
	err = s.srv.Serve(webSrvLn)
	if err != nil {
		panic(err)
	}
}
