package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/pbharrell/bloner/state"
)

type Game struct {
	mu      sync.Mutex
	mainPid int
	players map[int]*Player
	state   state.GameState
}

func runGame(players []*Player) {
	if len(players) < 4 {
		println("Error! Failed to start game due to insufficient number of players")
	}

	game := &Game{
		mainPid: players[0].PlayerId,
		players: make(map[int]*Player),
		state:   state.GameState{},
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
		p.send(Message{
			Type: "welcome",
			Data: map[string]any{
				"playerID": pid,
				"time":     time.Now().String(),
			},
		})
	}

	game.players[game.mainPid].send(Message{
		Type: "state_req",
		Data: "",
	})

	connected := true
	for {
		for _, p := range players {
			select {
			case msg := <-p.Data:
				game.msgHandler(msg)
			case isPConnected := <-p.IsConnected:
				connected = connected && isPConnected
				break
			default:
				time.Sleep(1000000000)
			}
		}

		if !connected {
			break
		}
	}

	// Initial broadcast
	game.broadcast(Message{
		Type: "info",
		Data: "Game has started with 4 players!",
	})
}

func (g *Game) msgHandler(msg Message) {
	switch msg.Type {
	case "state_res":
		state, ok := msg.Data.(state.GameState)
		if !ok {
			println("Invalid data type passed with type: `state_res`")
		}
		g.StateResponseHandler(state)
	default:
		fmt.Printf("Unsupported type in `game.msgHandler()` %v\n", msg.Type)
	}
}

func (g *Game) StateResponseHandler(state state.GameState) {
	g.state.Elements = state.Elements
	println("Updated game state!")
}

func (g *Game) broadcast(msg Message) {
	for _, p := range g.players {
		p.send(msg)
	}
}
