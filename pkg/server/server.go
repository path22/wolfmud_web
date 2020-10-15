package server

import (
	"fmt"
	webconfig "github.com/path22/wolfmud_web/pkg/config"
	"github.com/path22/wolfmud_web/pkg/sessions"
	"net"
	"net/http"
)

type Server struct {
	Host string
	Port string
	srv  http.Server
}

func New(conf *webconfig.System) *Server {
	s := new(Server)
	s.Host = conf.Address
	s.Port = conf.Port

	webSrvMux, shutdown := routing(conf)
	s.srv = http.Server{
		Handler: webSrvMux,
	}
	s.srv.RegisterOnShutdown(shutdown)
	return s
}

func routing(conf *webconfig.System) (*http.ServeMux, func()) {
	webSrvMux := http.NewServeMux()
	sess := sessions.New(conf)
	webSrvMux.HandleFunc("/", sess.Interface)
	webSrvMux.HandleFunc("/command", sess.Command)
	webSrvMux.HandleFunc("/message", sess.Message)
	return webSrvMux, sess.Shutdown
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
