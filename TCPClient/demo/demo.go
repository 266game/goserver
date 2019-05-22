package main

import (
	"fmt"
	"log"

	conn "github.com/266game/goserver/Connection"
	tcpclient "github.com/266game/goserver/TCPClient"
)

func main() {

	pClient := tcpclient.NewTCPClient()

	pClient.OnRead = func(pData *conn.TData) {
		buf := pData.GetBuffer()
		nLen := pData.GetLength()
		log.Println("收到包了长度是", nLen, "\n", string(buf), "\n", buf)
	}

	pClient.OnConnect = func(pConn *conn.TConnection) {
		log.Println(pConn.GetTCPConn().RemoteAddr(), "连接成功")
	}

	pClient.Connect("127.0.0.1:8083")

	strSending := ""
	for {
		fmt.Scanln(&strSending) //Scanln 扫描来自标准输入的文本，将空格分隔的值依次存放到后续的参数内，直到碰到换行
		if strSending == "\\q" {
			return
		}
		log.Println("需要发送内容是", strSending)
		pClient.WritePack([]byte(strSending))
	}
}
