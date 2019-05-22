package connection

import "net"

// TConnection 上下文会话
type TConnection struct {
	nIndex   uint32
	pTCPConn *net.TCPConn
	// 	buffer   []byte //
	// 	nLen     int    // 包长
}

// GetTCPConn 得到TCP连接指针
func (self *TConnection) GetTCPConn() *net.TCPConn {
	return self.pTCPConn
}

// Read 读取字节
func (self *TConnection) Read(b []byte) (int, error) {
	return self.pTCPConn.Read(b)
}

// Write 写字节
func (self *TConnection) Write(buff []byte) (int, error) {
	return self.pTCPConn.Write(buff)
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
	return self.pTCPConn.Write(buffReal)
}

// //GetData 获取数据
// func (self *TSession) GetData() ([]byte, int) {
// 	return self.buffer, self.nLen
// }

// // GetBuffer 获取Buffer
// func (self *TSession) GetBuffer() []byte {
// 	return self.buffer
// }

// Close 关闭连接
func (self *TConnection) Close() error {
	deleteConnection(self.nIndex) // 从MAP中移除
	return self.pTCPConn.Close()
}
