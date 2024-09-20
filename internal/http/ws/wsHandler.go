package ws

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/sunikka/clich-backend/internal/utils"
	"golang.org/x/net/websocket"
)

type message struct {
	senderID string
	content  string
}

type Server struct {
	conns map[*websocket.Conn]*utils.UserInfo
	mutex *sync.Mutex
}

func NewServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]*utils.UserInfo),
		mutex: &sync.Mutex{},
	}
}

func (s *Server) HandleConn(ws *websocket.Conn) {
	s.mutex.Lock()
	var client utils.UserInfo
	err := websocket.JSON.Receive(ws, &client)
	if err != nil {
		log.Println("Error receiving user data:", err)
		return
	}

	log.Printf("%s has connected to the server \n", client.Name)

	s.conns[ws] = &client
	s.mutex.Unlock()

	s.readLoop(ws, client)
}

func (s *Server) readLoop(ws *websocket.Conn, client utils.UserInfo) {
	buf := make([]byte, 1024)

	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("%s disconnected", client.Name)
				s.dropConnection(ws)
				break
			}
			fmt.Println("read error: ", err)
			continue
		}

		msg := buf[:n]
		log.Printf("%s sent a message: %s", client.Name, string(msg))

		s.broadcast(msg)
	}
}

func (s *Server) broadcast(b []byte) {
	for ws := range s.conns {
		go func(ws *websocket.Conn) {
			_, err := ws.Write(b)

			if err != nil {
				fmt.Printf("write error: %v, dropping client connection", err)
				s.dropConnection(ws)
			}
		}(ws)
	}
}

func (s *Server) dropConnection(ws *websocket.Conn) {
	s.mutex.Lock()
	delete(s.conns, ws)
	s.mutex.Unlock()
}
