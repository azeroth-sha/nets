package nets

import "time"

type SvrOption func(*server)

// WithSvrKeepalive 选择保活时间
func WithSvrKeepalive(dur time.Duration) SvrOption {
	return func(svr *server) {
		svr.conf.KeepAlive = dur
	}
}

// WithSvrListenCtl 选择自定义监听回调
func WithSvrListenCtl(ctl NetCtl) SvrOption {
	return func(svr *server) {
		svr.conf.Control = ctl
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
		svr.buffs.reset(int32(size))
	}
}

type CliOption func(*client)

// WithCliKeepalive 选择保活时间
func WithCliKeepalive(dur time.Duration) CliOption {
	return func(cli *client) {
		cli.conf.KeepAlive = dur
	}
}

// WithCliListenCtl 选择自定义监听回调
func WithCliListenCtl(ctl NetCtl) CliOption {
	return func(cli *client) {
		cli.conf.Control = ctl
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
		cli.buffs.reset(int32(size))
	}
}

// WithCliTimeout 自定义连接超时时间
func WithCliTimeout(dur time.Duration) CliOption {
	return func(cli *client) {
		cli.conf.Timeout = dur
	}
}
