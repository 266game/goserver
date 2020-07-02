package main

import (
	"time"

	"github.com/266game/goserver/conn"
	"github.com/266game/goserver/tcp"
)

func main() {
	pServer := tcp.NewTCPServer()

	pServer.OnRead = func(pData *conn.TData) {
		pData.Print()
	}

	pServer.Start(":4567")

	pClient := tcp.NewTCPClient()

	pClient.OnRead = func(pData *conn.TData) {
		pData.Print()
	}

	pClient.Connect("127.0.0.1:4567")

	for {
		pClient.WritePack([]byte("hello world"))
		time.Sleep(time.Second * 1)
	}

}
