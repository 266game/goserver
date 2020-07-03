package main

import (
	"log"
	"time"

	"github.com/266game/goserver/conn"
	"github.com/266game/goserver/ws"
)

func main() {
	pServer := &ws.TWSServer{}

	pServer.Start(":4567")

	// pServer.OnRead = func(pData *conn.TData) {
	// 	pData.Print()
	// }

	pServer.OnRun = func(pConnection *conn.TConnection) {
		buf := make([]byte, 4096)
		for {
			nLen, err := pConnection.Read(buf)
			if err != nil {
				return
			}
			log.Println("实际接收的包长", nLen, err)
			log.Println(string(buf[:nLen]))

			go pConnection.Write(buf[:nLen])
		}
	}

	time.Sleep(time.Second * 100)
}
