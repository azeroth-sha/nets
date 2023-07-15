package nets

import (
	"errors"
	"net"
)

var (
	ErrRunning = errors.New(`running`)
	ErrStopped = errors.New(`stopped`)
	ErrClosed  = net.ErrClosed
)

/*
	服务端选项
*/

type SvrOption func(*server)

// WithSvrListenConf 选择连接配置
func WithSvrListenConf(conf net.ListenConfig) SvrOption {
	return func(svr *server) {
		svr.conf = &conf
	}
}

// WithSvrTick 选择是否启用定时器
func WithSvrTick(flag bool) SvrOption {
	return func(svr *server) {
		svr.tick = flag
	}
}

// WithSvrBufCap 自定义buf容量
func WithSvrBufCap(size int) SvrOption {
	return func(svr *server) {
		svr.buffs.bufCap = uint32(size)
	}
}

/*
	客户端选项
*/

type CliOption func(*client)

// WithCliDialConf 选择连接配置
func WithCliDialConf(conf net.Dialer) CliOption {
	return func(cli *client) {
		cli.conf = &conf
	}
}

// WithCliTick 选择是否启用定时器
func WithCliTick(flag bool) CliOption {
	return func(cli *client) {
		cli.tick = flag
	}
}

// WithCliBufCap 自定义buf容量
func WithCliBufCap(size int) CliOption {
	return func(cli *client) {
		cli.buffs.bufCap = uint32(size)
	}
}
