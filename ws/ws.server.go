package ws

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/266game/goserver/conn"
	"golang.org/x/net/websocket"
)

// TWSServer websocket服务器类
type TWSServer struct {
	strAddress  string // 服务器地址
	MaxConnNum  int    // 最大连接数
	HTTPTimeOut time.Duration
	pListener   *net.TCPListener
	mutexConns  sync.Mutex
	wgLn        sync.WaitGroup
	// wgConns     sync.WaitGroup

	OnRun              func(*conn.TConnection) // 自处理循环回调
	OnRead             func(*conn.TData)       // 读取回调(buf, 包长, sessionid)
	OnClientConnect    func(*conn.TConnection) // 客户端连接上来了
	OnClientDisconnect func(*conn.TConnection) // 客户端断开了
}

// Start 开启
func (self *TWSServer) Start(strAddress string) {
	self.strAddress = strAddress

	go self.run()
}

// 开始监听
func (self *TWSServer) listen() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", self.strAddress)
	if err != nil {
		log.Println("错误", err)
	}

	pListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Println("错误", err)
	}

	self.pListener = pListener
}

//
func (self *TWSServer) run() {
	self.listen()
	time.Sleep(time.Millisecond * 16)

	httpServer := &http.Server{
		Addr: self.strAddress,
		Handler: websocket.Server{
			Handler: websocket.Handler(func(tcpConn *websocket.Conn) {

				self.wgLn.Add(1)
				defer self.wgLn.Done()

				pConnection := conn.CreateConnection(tcpConn)
				strRemoteAddr := pConnection.RemoteAddr()
				log.Println("监听到客户端的", strRemoteAddr, "连接")

				defer func() {
					log.Println(strRemoteAddr, "断开连接")

					if self.OnClientDisconnect != nil {
						self.OnClientDisconnect(pConnection)
					}

					pConnection.Close()
					// self.wgConns.Done()
				}()

				// 自带循环解包系统
				if self.OnRun != nil {
					log.Println("self on run")
					self.OnRun(pConnection)
					return
				}

				// 默认循环解包系统
				if self.OnRead != nil {
					buf := make([]byte, 4096)
					for {
						//
						nLen, err := pConnection.Read(buf)
						if err != nil {
							return
						}
						//
						log.Println("实际接收的包长", nLen, err)
						err = self.unpack(buf[0:nLen], nLen, pConnection)
					}
				}

			}),
			Handshake: func(config *websocket.Config, r *http.Request) error {
				log.Println("handshack")
				config.Origin, _ = url.ParseRequestURI("ws://" + r.RemoteAddr + r.URL.RequestURI())

				return nil
			},
		},
		ReadTimeout:    self.HTTPTimeOut,
		WriteTimeout:   self.HTTPTimeOut,
		MaxHeaderBytes: 1024,
	}

	httpServer.Serve(self.pListener)

}

// Close 关闭
func (self *TWSServer) Close() {
	self.pListener.Close()
	self.wgLn.Wait()
}

// 拆包
func (self *TWSServer) unpack(buf []byte, nLen int, pConnection *conn.TConnection) error {
	// 我们规定前两个字节是包的实际长度, 我们认为棋牌游戏当中是不可能超过单个包10K的容量
	nPackageLen := int(buf[0]) + int(buf[1])<<8

	if nPackageLen == nLen {
		// 包长符合, 包满足,直接派发
		log.Println("包长符合, 包满足,直接派发", nLen)
		// pSession := self.session(, pConnection)
		self.OnRead(conn.NewData(buf[2:nPackageLen], nPackageLen-2, pConnection))
		return nil
	}

	if nPackageLen < nLen {
		// 这个包需要拆包处理
		// pSession := self.session(buf[2:nPackageLen], nPackageLen-2, pConnection)
		self.OnRead(conn.NewData(buf[2:nPackageLen], nPackageLen-2, pConnection))
		self.unpack(buf[nPackageLen:nLen], nLen-nPackageLen, pConnection)
		return nil
	}

	// 还需要粘包
	buf1 := make([]byte, 4096)

	// 重新取一次包
	nLen1, err := pConnection.Read(buf1)

	if err != nil {
		return err
	}

	buf = append(buf, buf1...)
	self.unpack(buf, nLen+nLen1, pConnection)

	return nil
}
