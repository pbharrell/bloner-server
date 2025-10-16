package main

import (
	"fmt"
	"net"
	"sync"
)

var (
	lobby     []*Player
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
	player := Player{
		PlayerId:    playerSeq,
		Conn:        conn,
		IsConnected: make(chan bool),
		Data:        make(chan Message),
	}
	playerSeq++

	lobby = append(lobby, &player)

	go player.listen()

	if len(lobby) == 4 {
		// Create a new game with these 4 players
		playerSeq = 0
		players := lobby
		lobby = nil
		go runGame(players)
	}
}
