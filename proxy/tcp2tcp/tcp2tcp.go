package main

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// TCP 2 TCP 的代理转发

func main() {
	port2host(":5678", "127.0.0.1:4567") // 写死转发
}

func startServer(address string) net.Listener {
	log.Println("[+]", "尝试开启服务器在:["+address+"]")
	server, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln("[x]", "监听地址 ["+address+"] 失败.")
	}
	log.Println("[√]", "开始监听地址:["+address+"]")
	return server
}

func port2host(allowPort string, targetAddress string) {
	server := startServer("0.0.0.0:" + allowPort)
	for {
		conn := accept(server)
		if conn == nil {
			continue
		}
		//println(targetAddress)
		go func(targetAddress string) {
			log.Println("[+]", "开始连接服务器:["+targetAddress+"]")
			target, err := net.Dial("tcp", targetAddress)
			if err != nil {
				// temporarily unavailable, don't use fatal.
				log.Println("[x]", "连接目标地址 ["+targetAddress+"] 失败. 将在 ", 5, " 秒后进行重试. ")
				conn.Close()
				log.Println("[←]", "close the connect at local:["+conn.LocalAddr().String()+"] and remote:["+conn.RemoteAddr().String()+"]")
				time.Sleep(5 * time.Second)
				return
			}
			log.Println("[→]", "连接目标地址 ["+targetAddress+"] 成功.")
			forward(target, conn)
		}(targetAddress)
	}
}

func accept(listener net.Listener) net.Conn {
	conn, err := listener.Accept()
	if err != nil {
		log.Println("[x]", "接受连接 ["+conn.RemoteAddr().String()+"] 失败.", err.Error())
		return nil
	}
	log.Println("[√]", "接受一个新客户端. 远端地址是:["+conn.RemoteAddr().String()+"], 本地地址是:["+conn.LocalAddr().String()+"]")
	return conn
}

func forward(conn1 net.Conn, conn2 net.Conn) {
	var wg sync.WaitGroup
	// wait tow goroutines
	wg.Add(2)
	go connCopy(conn1, conn2, &wg)
	go connCopy(conn2, conn1, &wg)
	//blocking when the wg is locked
	wg.Wait()
}

func connCopy(conn1 net.Conn, conn2 net.Conn, wg *sync.WaitGroup) {
	io.Copy(conn1, conn2)
	conn1.Close()
	wg.Done()
}
