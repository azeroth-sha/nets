package nets

import "time"

type SvrOption func(*server)

// WithKeepalive 选择保活时间
func WithKeepalive(dur time.Duration) SvrOption {
	return func(svr *server) {
		svr.keepalive = dur
	}
}

// WithListenCtl 选择自定义监听回调
func WithListenCtl(ctl ListenCtl) SvrOption {
	return func(svr *server) {
		svr.listenCtl = ctl
	}
}

// WithTick 选择是否启用定时器
func WithTick(flag bool) SvrOption {
	return func(svr *server) {
		svr.tick = flag
	}
}

// WithBufCap 自定义buf容量
func WithBufCap(size int) SvrOption {
	return func(svr *server) {
		svr.buffs.reset(int32(size))
	}
}
