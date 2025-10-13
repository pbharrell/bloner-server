package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

type Player struct {
	PlayerId    int
	Conn        net.Conn
	IsConnected chan bool
	Data        chan Message
}

func (p *Player) listen() {
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

func (p *Player) send(msg Message) {
	data, _ := json.Marshal(msg)
	data = append(data, '\n') // so clients can read line by line
	_, err := p.Conn.Write(data)
	if err != nil {
		fmt.Println("send error:", err)
	}
}
