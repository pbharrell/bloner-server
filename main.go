package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/pbharrell/bloner-server/connection"
)

type Lobby struct {
	lock    *sync.Mutex
	id      int
	mainPid int
	players []*connection.Player
	state   connection.GameState
}

var (
	players    map[int]*connection.Player
	playerSeq  int = 0
	lobbies    map[int]Lobby
	lobbySeq   int = 0
	playerLock sync.Mutex
	lobbyLock  sync.Mutex
	gameLock   sync.Mutex
)

func (lobby *Lobby) getPlayer(id int) *connection.Player {
	for _, player := range lobby.players {
		if player.PlayerId == id {
			return player
		}
	}

	return nil
}

func getLobby(id int) *Lobby {
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	l, ok := lobbies[id]
	if !ok {
		return nil
	} else {
		return &l
	}
}

func main() {
	connection.SetLobbyRequestCallback(lobbyRequestCallback)

	players = make(map[int]*connection.Player)
	lobbies = make(map[int]Lobby)

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

func lobbyRequestCallback(request connection.LobbyRequest, player *connection.Player) int {
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	l, ok := lobbies[int(request)]
	if !ok {
		l = Lobby{
			id:      lobbySeq,
			mainPid: -1,
			players: []*connection.Player{player},
		}
		lobbies[l.id] = l
		lobbySeq++

	} else {
		l.players = append(l.players, player)
		lobbies[l.id] = l
	}

	if len(l.players) == 4 {
		go runGame(&l)
	}

	return l.id
}

func handleConnection(conn *websocket.Conn) {
	playerLock.Lock()

	fmt.Println("New connection!")

	// Add connection to lobby
	player := connection.Player{
		PlayerId:    playerSeq,
		Ctx:         context.Background(),
		WS:          conn,
		IsConnected: make(chan bool),
		Data:        make(chan connection.Message),
	}

	players[player.PlayerId] = &player
	playerSeq++
	playerLock.Unlock()

	go player.Listen()

	connected := true
	for {
		select {
		case msg := <-player.Data:
			PlayerMsgHandler(msg, &player)
			break
		case isPConnected := <-player.IsConnected:
			connected = connected && isPConnected
			break
		default:
			break
		}

		if !connected {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

}
