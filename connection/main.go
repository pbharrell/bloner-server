package connection

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coder/websocket"
)

type TurnType uint8

const (
	TrumpPass TurnType = iota
	TrumpPick
	TrumpDiscard
	CardPlay
)

type Card struct {
	Suit   uint8
	Number uint8
}

const (
	InvalidSuit   = 255
	InvalidNumber = 255
)

type PlayerState struct {
	PlayerId int `json:"playerId"`
	Cards    []Card
}

type TeamState struct {
	TricksWon   int `json:"tricksWon"`
	PlayerState [2]PlayerState
}

type GameState struct {
	PlayerId        int    `json:"playerId"`
	ActivePlayer    int    `json:"activePlayer"`
	TrumpDrawPlayer int    `json:"trumpDrawId"`
	TrumpSuit       int    `json:"trumpSuit"`
	DrawPile        []Card `json:"drawPile"`
	PlayPile        []Card `json:"playPile"`
	TeamState       [2]TeamState
}

type GameStateUpdate struct {
	PlayerId     int      `json:"playerId"`
	TurnType     TurnType `json:"typeType"`
	TrumpPick    int8     `json:"trumpPick"`
	TrumpDiscard Card     `json:"trumpDiscard"`
	CardPlay     Card     `json:"cardPlay"`
}

type Message struct {
	// Suppported types:
	// Lobby Types:
	//   MO: lobby_req; data = lobbyId
	//   MT: lobby_assign; data = { lobbyId, playerId }
	//
	// Game Init Types:
	//   MT: game_start; data = player_id? Maybe nil
	//
	// State Types:
	//   MT: state_req; data = nil
	//   MO: state_res; data = gameState (full)
	//   MT: state_res; data = gameState (full)
	//
	// Turn Types:
	//   MO: turn_info; data = gameStateUpdate (changed)
	//   MT: turn_info; data = gameStateUpdate (changed)
	//

	Type string `json:"type"`
	Data any    `json:"data"` // payload
}

// PAYLOAD MO TYPES
type LobbyRequest struct {
	LobbyId int16 `json:"lobbyId"`
}

// PAYLOAD MT TYPES
type LobbyAssign struct {
	LobbyId  int `json:"lobbyId"`
	PlayerId int `json:"playerId"`
}

// PAYLOAD UNIVERSAL TYPES
type StateResponse = GameState

type TurnInfo = GameStateUpdate

// Internal types
type Server struct {
	Ctx  context.Context
	WS   *websocket.Conn
	Data chan Message
}

func (s *Server) Listen() {
	defer s.WS.Close(websocket.StatusNormalClosure, "disconnect")

	for {
		msgType, data, err := s.WS.Read(s.Ctx)
		if err != nil {
			return
		}

		if msgType != websocket.MessageBinary {
			continue
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			println("invalid JSON")
			continue
		}

		s.Data <- msg
	}
}

func (s *Server) Send(msg Message) {
	data, _ := json.Marshal(msg)
	data = append(data, '\n') // so clients can read line by line
	err := s.WS.Write(s.Ctx, websocket.MessageBinary, data)
	if err != nil {
		fmt.Println("send error:", err)
	}
}

type Player struct {
	PlayerId    int
	Ctx         context.Context
	WS          *websocket.Conn
	IsConnected chan bool
	Data        chan Message
}

func (p *Player) Listen() {
	defer p.WS.Close(websocket.StatusNormalClosure, "disconnect")

	for {
		msgType, data, err := p.WS.Read(p.Ctx)
		if err != nil {
			return
		}

		if msgType != websocket.MessageBinary {
			continue
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			println("invalid JSON")
			continue
		}

		p.Data <- msg
	}

	// conn := p.WS
	//
	// defer conn.Close()
	// scanner := bufio.NewScanner(conn)
	//
	// p.IsConnected <- true
	//
	// for scanner.Scan() {
	// 	raw := scanner.Bytes()
	//
	// 	var msg Message
	// 	if err := json.Unmarshal(raw, &msg); err != nil {
	// 		println("invalid JSON")
	// 		continue
	// 	}
	//
	// 	p.Data <- msg
	// }
	// p.IsConnected <- false
}

func (p *Player) Send(msg Message) {
	data, _ := json.Marshal(msg)
	data = append(data, '\n') // so clients can read line by line
	err := p.WS.Write(p.Ctx, websocket.MessageBinary, data)
	if err != nil {
		fmt.Println("send error:", err)
	}
}
