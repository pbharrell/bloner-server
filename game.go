package main

import (
	"encoding/json"
	"fmt"

	"github.com/pbharrell/bloner-server/connection"
)

func runGame(lobby *Lobby) {
	if len(players) < 4 {
		println("Error! Failed to start game due to insufficient number of players")
	}

	lobby.lock.Lock()
	defer lobby.lock.Unlock()

	lobby.mainPid = lobby.players[0].PlayerId
	lobby.state = connection.GameState{}

	fmt.Println("Starting game with players:")

	lobby.getPlayer(lobby.mainPid).Send(connection.Message{
		Type: "state_req",
		Data: "",
	})

	// Initial broadcast
	lobby.broadcast(connection.Message{
		Type: "info",
		Data: "Game has started with 4 players!",
	})
}

func PlayerMsgHandler(msg connection.Message, p *connection.Player) {
	// Marshal Data back into JSON bytes
	raw, err := json.Marshal(msg.Data)
	if err != nil {
		println("marshal error:", err)
		return
	}

	fmt.Printf("Handling message of type \"%v\" from player with id: %v", msg.Type, p.PlayerId)
	if p.LobbyId < 0 {
		switch msg.Type {
		case "lobby_req":
			var request connection.LobbyRequest
			if err := json.Unmarshal(raw, &request); err != nil {
				println("Connection request unmarshal error:", err)
				return
			}
			p.LobbyRequestHandler(request)

		default:
			fmt.Printf("Unsupported message type received before player assigned lobby: %v\n", msg.Type)
		}

	} else {
		lobby := getLobby(p.LobbyId)
		switch msg.Type {
		case "state_res":
			var state connection.StateResponse
			if err := json.Unmarshal(raw, &state); err != nil {
				println("Game state response unmarshal error:", err)
				return
			}
			lobby.StateResponseHandler(state)

		case "turn_info":
			var state connection.TurnInfo
			if err := json.Unmarshal(raw, &state); err != nil {
				println("Turn info unmarshal error:", err)
				return
			}
			lobby.TurnInfoHandler(state)

		default:
			fmt.Printf("Unsupported type in `game.msgHandler()` %v\n", msg.Type)
		}
	}
}

func (l *Lobby) StateResponseHandler(state connection.StateResponse) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.state = state
	println("Handled game state response!")

	l.broadcast(connection.Message{
		Type: "state_res",
		Data: state,
	})
}

func (l *Lobby) TurnInfoHandler(state connection.TurnInfo) {
	l.lock.Lock()
	defer l.lock.Unlock()

	println("Handled turn info!")

	l.broadcast(connection.Message{
		Type: "turn_info",
		Data: state,
	})
}

func (l *Lobby) broadcast(msg connection.Message) {
	for _, p := range l.players {
		p.Send(msg)
	}
}
