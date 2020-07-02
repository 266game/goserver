package conn

import (
	"net"
	"sync"
	"time"
)

// TConnection 上下文会话
type TConnection struct {
	nIndex     uint64
	pConn      IConn
	mutexConns sync.Mutex // 锁
}

// GetIndex 获取索引值
func (self *TConnection) GetIndex() uint64 {
	return self.nIndex
}

// GetConn 得到连接指针
func (self *TConnection) GetConn() IConn {
	return self.pConn
}

// Read 读取字节
func (self *TConnection) Read(b []byte) (int, error) {
	return self.pConn.Read(b)
}

// Write 写字节
func (self *TConnection) Write(buff []byte) (int, error) {
	self.mutexConns.Lock()
	defer self.mutexConns.Unlock()
	return self.pConn.Write(buff)
}

// LocalAddr 本地socket端口地址
func (self *TConnection) LocalAddr() net.Addr {
	return self.pConn.LocalAddr()
}

// RemoteAddr 远程socket端口地址
func (self *TConnection) RemoteAddr() net.Addr {
	return self.pConn.RemoteAddr()
}

// SetDeadline 设置超时时间
// t = 0 意味着I/O操作不会超时。
func (self *TConnection) SetDeadline(t time.Time) error {
	return self.pConn.SetDeadline(t)
}

// SetReadDeadline 设置读取的超时时间
// t = 0 意味着I/O操作不会超时。
func (self *TConnection) SetReadDeadline(t time.Time) error {
	return self.pConn.SetReadDeadline(t)
}

// WritePack 写字节, 并且自动补齐包头部分
func (self *TConnection) WritePack(buff []byte) (int, error) {
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
	self.mutexConns.Lock()
	defer self.mutexConns.Unlock()
	return self.pConn.Write(buffReal)
}

// Close 关闭连接
func (self *TConnection) Close() error {
	self.mutexConns.Lock()
	defer self.mutexConns.Unlock()

	deleteConnection(self.nIndex) // 从MAP中移除
	return self.pConn.Close()
}
