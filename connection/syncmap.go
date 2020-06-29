package conn

import (
	"net"
	"sync"
)

var nAutoIncrease uint32
var mpConnection sync.Map
var mpChan sync.Map // 通道的ChanID

func init() {
	nAutoIncrease = 0
}

// CreateConnection 创建一个新的连接
func CreateConnection(pTCPConn *net.TCPConn) *TConnection {
	nAutoIncrease++
	n := nAutoIncrease
	pConnection := &TConnection{}
	pConnection.nIndex = n
	pConnection.pTCPConn = pTCPConn
	mpConnection.Store(n, pConnection)
	return pConnection
}

// FindConnection 查找
func FindConnection(nIndex uint32) *TConnection {
	v, ok := mpConnection.Load(nIndex)
	if ok {
		return v.(*TConnection)
	}
	return nil
}

func deleteConnection(nIndex uint32) {
	mpConnection.Delete(nIndex)
}

// CreateChan 创建一个通道
func CreateChan() (chan *TData, uint32) {
	nAutoIncrease++
	n := nAutoIncrease

	ch := make(chan *TData, 1)
	mpChan.Store(n, ch)
	return ch, n
}

// FindChan 删除掉chan
func FindChan(nIndex uint32) chan *TData {
	v, ok := mpChan.Load(nIndex)
	if ok {
		return v.(chan *TData)
	}
	return nil
}

// DeleteChan 删除掉chan
func DeleteChan(nIndex uint32) {
	mpChan.Delete(nIndex)
}
