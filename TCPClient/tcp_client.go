package tcpclient

import (
	"errors"
	"log"
	"net"
	"time"

	conn "github.com/266game/goserver/Connection"
)

func init() {
	//设置答应日志每一行前的标志信息，这里设置了日期，打印时间，当前go文件的文件名
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// TTCPClient TCP连接客户端
type TTCPClient struct {
	strAddress    string // 需要连接的服务器地址
	AutoReconnect bool
	bClose        bool              // 关闭状态
	pConnection   *conn.TConnection // 连接消息

	OnRun        func()                  //
	OnRead       func(*conn.TData)       // 读取回调
	OnConnect    func(*conn.TConnection) // 连接成功
	OnDisconnect func(*conn.TConnection) // 断开成功
}

// NewTCPClient 新建
func NewTCPClient() *TTCPClient {
	return &TTCPClient{}
}

// Connect 连接服务器
func (self *TTCPClient) Connect(strAddress string) {
	self.strAddress = strAddress
	self.bClose = false

	log.Println("Connect 地址", strAddress)
	go self.run() // 尝试连接
}

// WritePack 发送封包, 并且自动粘头
func (self *TTCPClient) WritePack(buff []byte) (int, error) {
	if self.pConnection == nil {
		log.Println("客户端未连接")
		return -1, errors.New("client have not connected")
	}
	return self.pConnection.WritePack(buff)
}

// 拨号
func (self *TTCPClient) dial() *net.TCPConn {
	for {
		tcpAddr, err := net.ResolveTCPAddr("tcp", self.strAddress)
		if err != nil {
			log.Println("错误", err)
		}
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err == nil || self.bClose {
			return conn
		}

		// log.Println("连接到", self.strAddress, "错误", err)
		time.Sleep(time.Second * 3) // 3秒后继续自动重新连接
		continue
	}
}

// 客户端尝试连接
func (self *TTCPClient) run() {
	tcpConn := self.dial() // 拨号与等待

	if tcpConn == nil {
		return
	}

	self.pConnection = conn.CreateConnection(tcpConn)
	strRemoteAddr := tcpConn.RemoteAddr()
	// 如果关闭了, 那么就关闭连接
	if self.bClose {
		self.pConnection.Close()
		self.pConnection = nil
		return
	}

	if self.OnConnect != nil {
		// 连接回调
		go self.OnConnect(self.pConnection)
	}

	// 在这里进行收包处理
	func() {
		if self.OnRun != nil {
			self.OnRun()
			return
		}

		// 默认循环解包系统
		if self.OnRead != nil {
			// 先定义一个4096的包长长度作为缓冲区
			buf := make([]byte, 4096)
			for {
				//
				nLen, err := self.pConnection.Read(buf)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println("实际接收的包长", nLen, err)
				err = self.unpack(buf[0:nLen], nLen, self.pConnection)
			}
		} else {
			log.Println("找不到处理网络的回调函数")
		}
	}()

	log.Println(strRemoteAddr, "断开连接")
	// cleanup
	self.pConnection.Close()
	self.pConnection = nil

	if self.AutoReconnect {
		time.Sleep(time.Second * 3) // 3秒后继续
		self.run()
	}
}

// Close 关闭连接
func (self *TTCPClient) Close() {
	self.bClose = true
	self.pConnection.Close()
}

// 拆包
func (self *TTCPClient) unpack(buf []byte, nLen int, pConnection *conn.TConnection) error {
	// 我们规定前两个字节是包的实际长度, 我们认为棋牌游戏当中是不可能超过单个包10K的容量
	nPackageLen := int(buf[0]) + int(buf[1])<<8

	if nPackageLen == nLen {
		// 包长符合, 包满足,直接派发
		log.Println("包长符合")
		self.OnRead(conn.NewData(buf[2:nPackageLen], nPackageLen-2, pConnection))
		return nil
	}

	if nPackageLen < nLen {
		// 这个包需要拆包处理
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

// // 创建一个新的session来使用
// func (self *TTCPClient) session(buff []byte, nLen int, tcpConn *net.TCPConn) *session.TSession {
// 	// 自增索引
// 	nTCPIndex := self.nAutoIncrease
// 	// 构建一个Session
// 	pSession := &session.TSession{} // 上下文会话消息
// 	pSession.SetTCPConn(tcpConn)
// 	pSession.SetAutoIncrease(nTCPIndex)
// 	pSession.SetData(buff, nLen)
// 	// 保存session
// 	self.mpSession.Store(nTCPIndex, pSession)
// 	self.nAutoIncrease++ // 自动递增

// 	// 设置15秒超时后自动干掉
// 	go func() {
// 		time.Sleep(time.Second * 15)
// 		self.mpSession.Delete(nTCPIndex)
// 	}()

// 	return pSession
// }
