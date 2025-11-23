package main

import (
	"encoding/json"
	"fmt"

	"github.com/pbharrell/bloner-server/connection"
)

type Game struct {
	mainPid int
	players map[int]*connection.Player
	state   connection.GameState
}

func runGame(players []*connection.Player) {
	if len(players) < 4 {
		println("Error! Failed to start game due to insufficient number of players")
	}

	game := &Game{
		mainPid: players[0].PlayerId,
		players: make(map[int]*connection.Player),
		state:   connection.GameState{},
	}

	gameLock.Lock()
	games[gameSeq] = game
	gameSeq++
	gameLock.Unlock()

	fmt.Println("Starting game with players:")

	for _, p := range players {
		pid := p.PlayerId
		game.players[pid] = p

		fmt.Printf("  Player %d: %s\n", pid, p.Conn.RemoteAddr())
		p.Send(connection.Message{
			Type: "lobby_assign",
			Data: connection.LobbyAssign{
				LobbyId:  gameSeq,
				PlayerId: pid,
			},
		})
	}

	game.players[game.mainPid].Send(connection.Message{
		Type: "state_req",
		Data: "",
	})

	connected := true
	for {
		for _, p := range players {
			select {
			case msg := <-p.Data:
				game.msgHandler(msg)
				break
			case isPConnected := <-p.IsConnected:
				connected = connected && isPConnected
				break
			default:
				break
			}
		}

		if !connected {
			break
		}
	}

	// Initial broadcast
	game.broadcast(connection.Message{
		Type: "info",
		Data: "Game has started with 4 players!",
	})
}

func (g *Game) msgHandler(msg connection.Message) {
	// Marshal Data back into JSON bytes
	raw, err := json.Marshal(msg.Data)
	if err != nil {
		println("marshal error:", err)
		return
	}

	switch msg.Type {
	case "state_res":
		var state connection.StateResponse
		if err := json.Unmarshal(raw, &state); err != nil {
			println("GameState unmarshal error:", err)
			return
		}
		g.StateResponseHandler(state)

	case "turn_info":
		var state connection.TurnInfo
		if err := json.Unmarshal(raw, &state); err != nil {
			println("GameState unmarshal error:", err)
			return
		}
		g.TurnInfoHandler(state)

	default:
		fmt.Printf("Unsupported type in `game.msgHandler()` %v\n", msg.Type)
	}
}

func (g *Game) StateResponseHandler(state connection.StateResponse) {
	g.state = state
	println("Handled game state response!")

	for _, p := range g.players {
		p.Send(connection.Message{
			Type: "state_res",
			Data: state,
		})
	}
}

func (g *Game) TurnInfoHandler(state connection.TurnInfo) {
	println("Handled turn info!")

	for _, p := range g.players {
		p.Send(connection.Message{
			Type: "turn_info",
			Data: state,
		})
	}
}

func (g *Game) broadcast(msg connection.Message) {
	for _, p := range g.players {
		p.Send(msg)
	}
}
