package nets

import (
	"context"
	"net"
	"syscall"
	"time"
)

// ListenCtl 监听方法回调函数
type ListenCtl func(network, address string, c syscall.RawConn) error

// Listen 通过协议连接监听端口
func Listen(protoAddr string, keepalive time.Duration, ctl ListenCtl) (net.Listener, error) {
	proto, addr, err := parseAddr(protoAddr)
	if err != nil {
		return nil, err
	}
	conf := net.ListenConfig{
		Control:   ctl,
		KeepAlive: keepalive,
	}
	return conf.Listen(context.Background(), proto, addr)
}
