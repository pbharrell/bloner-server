package connection

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

type Message struct {
	// Suppported types:
	// Lobby Types:
	//   MO: lobby_req; data = lobby_id
	//   MT: lobby_assign; data = { lobby_id, player_id }
	//
	// Game Init Types:
	//   MT: game_start; data = player_id
	//
	// State Types:
	//   MT: state_req; data = nil
	//   MO: state_res; data = gameState (full)
	//   MT: state_res; data = gameState (full)
	//
	// Turn Types:
	//   MO: state_update; data = gameState (changed)
	//   MT: state_update; data = gameState (changed)
	//

	Type string `json:"type"`
	Data any    `json:"data"` // payload
}

type Server struct {
	Conn net.Conn
	Data chan Message
}

func (s *Server) Listen() {
	conn := s.Conn

	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		raw := scanner.Bytes()

		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			println("invalid JSON")
			continue
		}

		s.Data <- msg
	}
}

func (s *Server) Send(msg Message) {
	data, _ := json.Marshal(msg)
	data = append(data, '\n') // so clients can read line by line
	_, err := s.Conn.Write(data)
	if err != nil {
		fmt.Println("send error:", err)
	}
}

type Player struct {
	PlayerId    int
	Conn        net.Conn
	IsConnected chan bool
	Data        chan Message
}

func (p *Player) Listen() {
	conn := p.Conn

	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	p.IsConnected <- true

	for scanner.Scan() {
		raw := scanner.Bytes()

		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			println("invalid JSON")
			continue
		}

		p.Data <- msg
	}
	p.IsConnected <- false
}

func (p *Player) Send(msg Message) {
	data, _ := json.Marshal(msg)
	data = append(data, '\n') // so clients can read line by line
	_, err := p.Conn.Write(data)
	if err != nil {
		fmt.Println("send error:", err)
	}
}
