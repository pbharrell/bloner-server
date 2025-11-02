package main

import (
	"fmt"
	"net"
	"sync"

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
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}
	fmt.Println("Server started on :9000")
	games = make(map[int]*Game)

	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	fmt.Println("New connection:", conn.RemoteAddr())

	// Add connection to lobby
	player := connection.Player{
		PlayerId:    playerSeq,
		Conn:        conn,
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
