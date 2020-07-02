package conn

import (
	"net"
	"time"
	// "golang.org/x/net/websocket"
)

// IConn 连接接口
type IConn interface {
	Read(b []byte) (int, error)
	Write(buff []byte) (int, error)
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	Close() error
}
