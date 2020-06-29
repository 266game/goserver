package main

import (
	"fmt"
	"time"

	"github.com/266game/goserver/tcp"
)

func main() {
	pServer := tcp.NewTCPServer()

	pServer.OnRead = func(pData *tcp.TData) {
		buf := pData.GetBuffer()
		nLen := pData.GetLength()

		fmt.Print("     00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F")
		for i := 0; i < nLen; i++ {
			if i%16 == 0 {
				fmt.Printf("\n%04d", i/16)
			}
			fmt.Printf(" %02x", buf[i])

		}
		fmt.Println("\n", string(buf)) //打印出来
		pData.GetIndex()
	}

	pServer.Start(":4567")

	pClient := tcp.NewTCPClient()

	pClient.OnRead = func(pData *tcp.TData) {
		buf := pData.GetBuffer()
		nLen := pData.GetLength()

		fmt.Print("     00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F")
		for i := 0; i < nLen; i++ {
			if i%16 == 0 {
				fmt.Printf("\n%04d", i/16)
			}
			fmt.Printf(" %02x", buf[i])

		}
		fmt.Println("\n", string(buf)) //打印出来
	}

	pClient.Connect("127.0.0.1:4567")

	for {
		pClient.WritePack([]byte("hello world"))
		time.Sleep(time.Second * 1)
	}

}
