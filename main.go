package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/coder/websocket"

	"github.com/pbharrell/bloner-server/connection"
)

var (
	lobby     []*connection.Player
	games     map[int]*Game
	lobbyLock sync.Mutex
	gameLock  sync.Mutex
	gameSeq   int = 0
	playerSeq int = 0
)

func main() {
	games = make(map[int]*Game)

	http.HandleFunc("/ws", wsHandler)

	fmt.Println("Server started on :9000")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		panic(err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"}, // dev only
	})
	if err != nil {
		return
	}

	go handleConnection(conn)
}

func handleConnection(conn *websocket.Conn) {
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	// fmt.Println("New connection:", conn.RemoteAddr())

	// Add connection to lobby
	player := connection.Player{
		PlayerId:    playerSeq,
		Ctx:         context.Background(),
		WS:          conn,
		IsConnected: make(chan bool),
		Data:        make(chan connection.Message),
	}
	playerSeq++

	lobby = append(lobby, &player)

	go player.Listen()

	// TODO: Handle lobby requests!
	if len(lobby) == 4 {
		// Create a new game with these 4 players
		playerSeq = 0
		players := lobby
		lobby = nil
		go runGame(players)
	}
}
