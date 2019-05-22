package wxclient

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	conn "github.com/266game/goserver/Connection"
)

// TNetMsgXpHeader 4字节网络结构体 套件
type TNetMsgXpHeader struct {
	PackLength    uint32 // 4字节封包长度(含包头)，可变
	Headlength    uint16 // 2字节表示头部长度,固定值，0x10
	ClientVersion uint16 // 2字节表示协议版本，固定值，0x01
	CmdID         uint32 // 4字节cmdid，可变
	Seq           uint32 // 4字节封包编号，可变
}

// TWxClient wx客户端
type TWxClient struct {
	strAddress    string // 需要连接的服务器地址
	AutoReconnect bool
	bClose        bool // 关闭状态

	pConnection *conn.TConnection // 连接消息

	OnRun        func()                  //
	OnRead       func(*conn.TData)       // 读取回调
	OnConnect    func(*conn.TConnection) // 连接成功
	OnDisconnect func(*conn.TConnection) // 断开成功
}

// NewWxClient 新建
func NewWxClient() *TWxClient {

	return &TWxClient{}
}

// Connect 连接服务器
func (self *TWxClient) Connect(strAddress string) {
	self.strAddress = strAddress
	self.bClose = false

	log.Println("Connect 地址", strAddress)
	go self.run() // 尝试连接
}

// WritePack 发包
func (self *TWxClient) WritePack(buff []byte, pHeader *TNetMsgXpHeader) {
	newbuf := new(bytes.Buffer)

	pHeader.PackLength = 16 + uint32(len(buff)) // 长度
	pHeader.Headlength = 0x10
	pHeader.ClientVersion = 0x01

	// 扁平化拆包
	err := binary.Write(newbuf, binary.BigEndian, *pHeader)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}

	self.pConnection.Write(newbuf.Bytes())
}

// 拨号
func (self *TWxClient) dial() *net.TCPConn {
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
func (self *TWxClient) run() {
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

	time.Sleep(time.Second * 3) // 3秒后继续
	self.run()
}

// Close 关闭连接
func (self *TWxClient) Close() {
	self.bClose = true
	self.pConnection.Close()
}

// 拆包
func (self *TWxClient) unpack(buf []byte, nLen int, pConnection *conn.TConnection) error {
	// 我们规定前两个字节是包的实际长度, 我们认为棋牌游戏当中是不可能超过单个包10K的容量
	nPackageLen := int(buf[0])<<32 + int(buf[1])<<16 + int(buf[2])<<8 + int(buf[3])

	if nPackageLen == nLen {
		// 包长符合, 包满足,直接派发
		log.Println("包长符合")
		self.OnRead(conn.NewData(buf[:], nPackageLen, pConnection))
		return nil
	}

	if nPackageLen < nLen {
		// 这个包需要拆包处理
		self.OnRead(conn.NewData(buf[:], nPackageLen, pConnection))
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
