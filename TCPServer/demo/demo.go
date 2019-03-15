﻿package main

import (
	"fmt"
	"time"

	conn "github.com/266game/goserver/Connection"
	tcpserver "github.com/266game/goserver/TCPServer"
)

func main() {
	pServer := tcpserver.NewTCPServer()

	pServer.OnRead = func(pData *conn.TData) {
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

	pServer.Start(":4567")

	time.Sleep(time.Hour * 10)
}
