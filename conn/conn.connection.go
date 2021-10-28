package conn

import (
	"errors"
	"net"
	"sync"
	"time"
)

// TConnection 上下文会话
type TConnection struct {
	nIndex         uint64
	pConn          IConn
	mutexConns     sync.Mutex // 锁
	pAes           *TAes
	nTag           int64       // 自定义tag, 存储用
	pCustomPointer interface{} // 自定义的指针, 存储数据用
}

// GetIndex 获取索引值
func (m *TConnection) GetIndex() uint64 {
	return m.nIndex
}

// GetConn 得到连接指针
func (m *TConnection) GetConn() IConn {
	return m.pConn
}

// Read 读取字节
func (m *TConnection) Read(b []byte) (int, error) {
	return m.pConn.Read(b)
}

// Write 写字节
func (m *TConnection) Write(buff []byte) (int, error) {
	m.mutexConns.Lock()
	defer m.mutexConns.Unlock()
	return m.pConn.Write(buff)
}

// LocalAddr 本地socket端口地址
func (m *TConnection) LocalAddr() net.Addr {
	return m.pConn.LocalAddr()
}

// RemoteAddr 远程socket端口地址
func (m *TConnection) RemoteAddr() net.Addr {
	return m.pConn.RemoteAddr()
}

// RemoteAddrHost 获取远程socket端口地址的IP地址
func (m *TConnection) RemoteAddrHost() string {
	strHost, _, err := net.SplitHostPort(m.RemoteAddr().String())
	if err != nil {
		return ""
	}
	return strHost
}

// SetDeadline 设置超时时间
// t = 0 意味着I/O操作不会超时。
func (m *TConnection) SetDeadline(t time.Time) error {
	return m.pConn.SetDeadline(t)
}

// SetReadDeadline 设置读取的超时时间
// t = 0 意味着I/O操作不会超时。
func (m *TConnection) SetReadDeadline(t time.Time) error {
	return m.pConn.SetReadDeadline(t)
}

func (m *TConnection) SetAesKey(binAesKey []byte) {
	m.pAes = NewAES(binAesKey)
}

func (m *TConnection) UnpackAes(buff []byte) ([]byte, error) {
	if m.pAes == nil {
		return buff, errors.New("")
	}

	buff2, err := m.pAes.UnAES(buff)

	if err != nil {
		return buff, err
	}

	return buff2, nil
}

// WriteAesPack 发送AES加密后的包
func (m *TConnection) WriteAesPack(buff []byte) (int, error) {
	if m.pAes == nil {
		return 0, errors.New("no aes, please create first")
	}

	buff, err := m.pAes.CoAES(buff)
	if err != nil {
		return 0, err
	}
	return m.WritePack(buff)
}

// WritePack 写字节, 并且自动补齐包头部分
func (m *TConnection) WritePack(buff []byte) (int, error) {
	// 在这里要进行一轮组包
	nLen := len(buff) + 2 // 需要补充2个字节的包长头
	if nLen > 65535 {
		return 0, nil
	}
	// 补齐二进制包头
	buffLen := [2]byte{}
	buffLen[0] = byte(nLen)
	buffLen[1] = byte(nLen >> 8)

	// 拼接带长度的Buff
	buffReal := append(buffLen[:], buff...)
	m.mutexConns.Lock()
	defer m.mutexConns.Unlock()
	return m.pConn.Write(buffReal)
}

// Close 关闭连接
func (m *TConnection) Close() error {
	m.mutexConns.Lock()
	defer m.mutexConns.Unlock()

	deleteConnection(m.nIndex) // 从MAP中移除
	return m.pConn.Close()
}

//
func (m *TConnection) SetCustomPointer(p interface{}) {
	m.pCustomPointer = p
}
func (m *TConnection) GetCustomPointer() interface{} {
	return m.pCustomPointer
}
func (m *TConnection) SetTag(n int64) {
	m.nTag = n
}
func (m *TConnection) GetTag() int64 {
	return m.nTag
}
