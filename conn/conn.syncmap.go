package conn

import (
	"sync"
)

var nAutoIncrease uint64
var mpConnection sync.Map
var mpChan sync.Map // 通道的ChanID
var mutexAuto sync.Mutex

func init() {
	nAutoIncrease = 0
}

// CreateConnection 创建一个新的连接
func CreateConnection(pConn IConn) *TConnection {
	mutexAuto.Lock()
	nAutoIncrease++
	n := nAutoIncrease
	mutexAuto.Unlock()

	pConnection := &TConnection{}
	pConnection.nIndex = n
	pConnection.pConn = pConn
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

// RangeConnection 遍历
func RangeConnection(f func(key, value interface{}) bool) {
	mutexAuto.Lock()
	mpConnection.Range(f)
	mutexAuto.Unlock()
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
