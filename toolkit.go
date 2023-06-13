package nets

import (
	"net"
	"net/url"
	"strconv"
)

// parseAddr 协议地址解析
func parseAddr(protoAddr string) (proto, addr string, err error) {
	addrUrl, err := url.Parse(protoAddr)
	if err != nil {
		return "", "", err
	}
	proto = `tcp`
	if len(addrUrl.Scheme) != 0 {
		proto = addrUrl.Scheme
	}
	addr = addrUrl.Host
	if _, port, err := net.SplitHostPort(addrUrl.Host); err != nil {
		return "", "", err
	} else if _, e := strconv.Atoi(port); e != nil {
		return "", "", ErrHostPort
	}
	return proto, addr, nil
}
