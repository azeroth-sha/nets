package nets

import (
	"context"
	"net"
	"syscall"
)

// NetCtl 连接处理方法
type NetCtl func(network, address string, c syscall.RawConn) error

// Listen 通过协议连接监听
func Listen(protoAddr string, conf *net.ListenConfig) (net.Listener, error) {
	proto, addr, err := parseAddr(protoAddr)
	if err != nil {
		return nil, err
	}
	return conf.Listen(context.Background(), proto, addr)
}

// Dial 通过协议连接拨号
func Dial(ctx context.Context, protoAddr string, conf *net.Dialer) (net.Conn, error) {
	proto, addr, err := parseAddr(protoAddr)
	if err != nil {
		return nil, err
	}
	return conf.DialContext(ctx, proto, addr)
}
