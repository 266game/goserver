package conn

import (
	"fmt"
)

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
func (m *TData) GetBuffer() []byte {
	return m.buffer
}

// GetIndex 获取自增索引
func (m *TData) GetIndex() uint64 {
	return m.pConnection.GetIndex()
}

// GetConnection 获取连接
func (m *TData) GetConnection() *TConnection {
	return m.pConnection
}

// GetLength 获取长度
func (m *TData) GetLength() int {
	return m.nLen
}

// Print 打印
func (m *TData) Print() {
	buf := m.GetBuffer()
	nLen := m.GetLength()

	fmt.Print("     00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F")
	for i := 0; i < nLen; i++ {
		if i%16 == 0 {
			fmt.Printf("\n%04d", i/16)
		}
		fmt.Printf(" %02x", buf[i])

	}
	fmt.Println("\n    ", string(buf)) //打印出来
}
