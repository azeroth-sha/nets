package nets

import (
	"context"
	"net"
	"strings"
)

// listenAddr 通过协议连接监听
func listenAddr(protoAddr string, conf *net.ListenConfig) (net.Listener, error) {
	proto, addr := parseAddr(protoAddr)
	return conf.Listen(context.Background(), proto, addr)
}

// dial 通过协议连接拨号
func dial(protoAddr string, conf *net.Dialer) (net.Conn, error) {
	proto, addr := parseAddr(protoAddr)
	return conf.DialContext(context.Background(), proto, addr)
}

// parseAddr 协议地址解析
func parseAddr(protoAddr string) (proto, addr string) {
	scheme := "tcp"
	if idx := strings.Index(protoAddr, "://"); idx != -1 && len(protoAddr[:idx]) > 0 {
		scheme = protoAddr[:idx]
		protoAddr = protoAddr[idx+3:]
	}
	return scheme, protoAddr
}
