package tcp

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/266game/goserver/conn"
)

// TTCPServer 服务器类
type TTCPServer struct {
	strAddress string // 服务器地址
	MaxConnNum int
	pListener  *net.TCPListener // 监听者
	mutexConns sync.Mutex
	wgLn       sync.WaitGroup
	wgConns    sync.WaitGroup

	OnRun              func(*conn.TConnection) // 自处理循环回调
	OnRead             func(*conn.TData)       // 读取回调(buf, 包长, sessionid)
	OnClientConnect    func(*conn.TConnection) // 客户端连接上来了
	OnClientDisconnect func(*conn.TConnection) // 客户端断开了
}

func init() {
	//设置答应日志每一行前的标志信息，这里设置了日期，打印时间，当前go文件的文件名
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// NewTCPServer 新建
func NewTCPServer() *TTCPServer {
	return &TTCPServer{}
}

// Start 启动服务器
func (self *TTCPServer) Start(strAddress string) {
	self.strAddress = strAddress

	log.Println("Start 地址", strAddress)
	go self.run()
}

// Stop 停服
func (self *TTCPServer) Stop() {
	self.pListener.Close()
	self.wgLn.Wait()
	self.wgConns.Wait()
}

// 开始监听
func (self *TTCPServer) listen() {
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

// 运行
func (self *TTCPServer) run() {
	self.listen()
	time.Sleep(time.Millisecond * 16)
	self.wgLn.Add(1)
	defer self.wgLn.Done()

	var tempDelay time.Duration
	for {
		tcpConn, err := self.pListener.AcceptTCP()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Println("accept error: ", err, "; retrying in ", tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		tempDelay = 0

		self.wgConns.Add(1)

		pConnection := conn.CreateConnection(tcpConn)
		strRemoteAddr := pConnection.RemoteAddr()
		log.Println("监听到客户端的", strRemoteAddr, "连接")
		if self.OnClientConnect != nil {
			self.OnClientConnect(pConnection)
		}

		go func() {
			defer func() {
				log.Println(strRemoteAddr, "断开连接")

				if self.OnClientDisconnect != nil {
					self.OnClientDisconnect(pConnection)
				}

				pConnection.Close()
				self.wgConns.Done()
			}()

			// if self.OnClose

			// 自带循环解包系统
			if self.OnRun != nil {
				self.OnRun(pConnection)
				return
			}
			// 默认循环解包系统
			if self.OnRead != nil {
				// 先定义一个4096的包长长度作为缓冲区
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
		}()
	}
}

// 拆包
func (self *TTCPServer) unpack(buf []byte, nLen int, pConnection *conn.TConnection) error {
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
