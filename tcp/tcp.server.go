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
	// mutexConns sync.Mutex
	wgLn    sync.WaitGroup
	wgConns sync.WaitGroup

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
func (m *TTCPServer) Start(strAddress string) {
	m.strAddress = strAddress

	log.Println("Start 地址", strAddress)
	go m.run()
}

// Stop 停服
func (m *TTCPServer) Stop() {
	m.pListener.Close()
	m.wgLn.Wait()
	m.wgConns.Wait()
}

// 开始监听
func (m *TTCPServer) listen() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", m.strAddress)
	if err != nil {
		log.Println("错误", err)
	}

	pListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Println("错误", err)
	}

	m.pListener = pListener
}

// 运行
func (m *TTCPServer) run() {
	m.listen()
	time.Sleep(time.Millisecond * 16)
	m.wgLn.Add(1)
	defer m.wgLn.Done()

	var tempDelay time.Duration
	for {
		tcpConn, err := m.pListener.AcceptTCP()
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

		m.wgConns.Add(1)

		pConnection := conn.CreateConnection(tcpConn)
		strRemoteAddr := pConnection.RemoteAddr()
		log.Println("监听到客户端的", strRemoteAddr, "连接")
		if m.OnClientConnect != nil {
			m.OnClientConnect(pConnection)
		}

		go func() {
			defer func() {
				log.Println(strRemoteAddr, "断开连接")

				if m.OnClientDisconnect != nil {
					m.OnClientDisconnect(pConnection)
				}

				pConnection.Close()
				m.wgConns.Done()
			}()

			// if m.OnClose

			// 自带循环解包系统
			if m.OnRun != nil {
				m.OnRun(pConnection)
				return
			}
			// 默认循环解包系统
			if m.OnRead != nil {
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
					_ = m.unpack(buf[0:nLen], nLen, pConnection)
				}
			}
		}()
	}
}

// 拆包
func (m *TTCPServer) unpack(buf []byte, nLen int, pConnection *conn.TConnection) error {
	// 我们规定前两个字节是包的实际长度, 我们认为棋牌游戏当中是不可能超过单个包10K的容量
	nPackageLen := int(buf[0]) + int(buf[1])<<8

	if nPackageLen == nLen {
		// 包长符合, 包满足,直接派发
		log.Println("包长符合, 包满足,直接派发", nLen)
		// pSession := m.session(, pConnection)
		m.OnRead(conn.NewData(buf[2:nPackageLen], nPackageLen-2, pConnection))
		return nil
	}

	if nPackageLen < nLen {
		// 这个包需要拆包处理
		// pSession := m.session(buf[2:nPackageLen], nPackageLen-2, pConnection)
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
