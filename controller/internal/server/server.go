package server

import (
	"log"
)

type Server struct {
	// 占位
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start() error {
	log.Println("Server started")
	return nil
}

func (s *Server) Stop() {
	log.Println("Server stopped")
}
