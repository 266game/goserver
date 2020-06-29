package tcp

// TData 数据
type TData struct {
	buffer      []byte       //
	nLen        int          // 包长
	pConnection *TConnection //
}

//NewData 设置数据
func NewData(buff []byte, nLen int, p *TConnection) *TData {
	pData := &TData{}
	pData.buffer = buff
	pData.nLen = nLen
	pData.pConnection = p
	return pData
}

// GetBuffer 获取buffer
func (self *TData) GetBuffer() []byte {
	return self.buffer
}

// GetIndex 获取自增索引
func (self *TData) GetIndex() uint32 {
	return self.pConnection.nIndex
}

// GetConnection 获取连接
func (self *TData) GetConnection() *TConnection {
	return self.pConnection
}

// GetLength 获取长度
func (self *TData) GetLength() int {
	return self.nLen
}
