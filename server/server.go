package server

import (
	"fmt"
	"net"
	"net/http"
)

type Server struct {
	Host string
	Port string
	srv  http.Server
}

func New(host string, port string) (*Server, error) {
	s := new(Server)
	webSrvMux := http.NewServeMux()
	webSrvMux.Handle("/", NewTCPClient())
	s.srv = http.Server{
		Handler: webSrvMux,
	}
	webSrvLn, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}
	go func() {
		err := s.srv.Serve(webSrvLn)
		if err != nil {
			panic(err)
		}
	}()
	return s, nil
}
