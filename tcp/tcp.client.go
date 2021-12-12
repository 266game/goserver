package tcp

import (
	"errors"
	"log"
	"net"
	"time"

	"github.com/266game/goserver/conn"
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

	OnRun        func(*conn.TConnection) //
	OnRead       func(*conn.TData)       // 读取回调
	OnConnect    func(*conn.TConnection) // 连接成功
	OnDisconnect func(*conn.TConnection) // 断开成功
}

// NewTCPClient 新建
func NewTCPClient() *TTCPClient {
	return &TTCPClient{}
}

// Connect 连接服务器
func (m *TTCPClient) Connect(strAddress string) {
	m.strAddress = strAddress
	m.bClose = false

	log.Println("Connect 地址", strAddress)
	go m.run() // 尝试连接
}

// Write 发送包
func (m *TTCPClient) Write(buff []byte) (int, error) {
	if m.pConnection == nil {
		log.Println("客户端未连接")
		return -1, errors.New("client have not connected")
	}
	return m.pConnection.Write(buff)
}

// WritePack 发送封包, 并且自动粘头
func (m *TTCPClient) WritePack(buff []byte) (int, error) {
	if m.pConnection == nil {
		log.Println("客户端未连接")
		return -1, errors.New("client have not connected")
	}
	return m.pConnection.WritePack(buff)
}

// 拨号
func (m *TTCPClient) dial() *net.TCPConn {
	for {
		tcpAddr, err := net.ResolveTCPAddr("tcp", m.strAddress)
		if err != nil {
			log.Println("错误", err)
		}
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err == nil || m.bClose {
			return conn
		}

		// log.Println("连接到", m.strAddress, "错误", err)
		time.Sleep(time.Second * 3) // 3秒后继续自动重新连接
		continue
	}
}

// 客户端尝试连接
func (m *TTCPClient) run() {
	tcpConn := m.dial() // 拨号与等待

	if tcpConn == nil {
		return
	}

	m.pConnection = conn.CreateConnection(tcpConn)
	strRemoteAddr := tcpConn.RemoteAddr()
	// 如果关闭了, 那么就关闭连接
	if m.bClose {
		m.pConnection.Close()
		m.pConnection = nil
		return
	}

	if m.OnConnect != nil {
		// 连接回调
		m.OnConnect(m.pConnection)
	}

	// 在这里进行收包处理
	func() {
		if m.OnRun != nil {
			m.OnRun(m.pConnection)
			return
		}

		// 默认循环解包系统
		if m.OnRead != nil {
			// 先定义一个4096的包长长度作为缓冲区
			buf := make([]byte, 4096)
			for {
				//
				nLen, err := m.pConnection.Read(buf)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println("实际接收的包长", nLen, err)
				_ = m.unpack(buf[0:nLen], nLen, m.pConnection)
			}
		} else {
			log.Println("找不到处理网络的回调函数")
		}
	}()

	if m.OnDisconnect != nil {
		m.OnDisconnect(m.pConnection)
	}
	log.Println(strRemoteAddr, "断开连接")
	// cleanup
	m.pConnection.Close()
	m.pConnection = nil

	if m.AutoReconnect {
		time.Sleep(time.Second * 3) // 3秒后继续
		m.run()
	}
}

// Close 关闭连接
func (m *TTCPClient) Close() {
	m.bClose = true
	m.pConnection.Close()
}

// 拆包
func (m *TTCPClient) unpack(buf []byte, nLen int, pConnection *conn.TConnection) error {
	// 我们规定前两个字节是包的实际长度, 我们认为棋牌游戏当中是不可能超过单个包10K的容量
	nPackageLen := int(buf[0]) + int(buf[1])<<8

	if nPackageLen == nLen {
		// 包长符合, 包满足,直接派发
		log.Println("包长符合")
		m.OnRead(conn.NewData(buf[2:nPackageLen], nPackageLen-2, pConnection))
		return nil
	}

	if nPackageLen < nLen {
		// 这个包需要拆包处理
		m.OnRead(conn.NewData(buf[2:nPackageLen], nPackageLen-2, pConnection))
		m.unpack(buf[nPackageLen:nLen], nLen-nPackageLen, pConnection)
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
	m.unpack(buf, nLen+nLen1, pConnection)

	return nil
}
