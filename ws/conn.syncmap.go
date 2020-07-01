package ws

import (
	"sync"

	"golang.org/x/net/websocket"
)

var nAutoIncrease uint64
var mpConnection sync.Map
var mpChan sync.Map // 通道的ChanID

func init() {
	nAutoIncrease = 0
}

// CreateConnection 创建一个新的连接
func CreateConnection(pWSConn *websocket.Conn) *TConnection {
	nAutoIncrease++
	n := nAutoIncrease
	pConnection := &TConnection{}
	pConnection.nIndex = n
	pConnection.pWSConn = pWSConn
	mpConnection.Store(n, pConnection)
	return pConnection
}

// FindConnection 查找
func FindConnection(nIndex uint64) *TConnection {
	v, ok := mpConnection.Load(nIndex)
	if ok {
		return v.(*TConnection)
	}
	return nil
}

func deleteConnection(nIndex uint64) {
	mpConnection.Delete(nIndex)
}

// CreateChan 创建一个通道
func CreateChan() (chan *TData, uint64) {
	nAutoIncrease++
	n := nAutoIncrease

	ch := make(chan *TData, 1)
	mpChan.Store(n, ch)
	return ch, n
}

// FindChan 查找chan
func FindChan(nIndex uint64) chan *TData {
	v, ok := mpChan.Load(nIndex)
	if ok {
		return v.(chan *TData)
	}
	return nil
}

// DeleteChan 删除掉chan
func DeleteChan(nIndex uint64) {
	mpChan.Delete(nIndex)
}
