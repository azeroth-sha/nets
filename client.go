package nets

import (
	"context"
	"net"
	"sync/atomic"
	"time"
)

type CliHandler interface {
	ConnHandler
	OnBoot(cli Client) (err error) // 启动客户端触发
	OnShutdown(cli Client)         // 关闭客户端触发
	OnTick() (dur time.Duration)   // 定时触发
}

type Client interface {
	Serve() error           // 启动服务
	Shutdown() error        // 关停服务
	NewConn() (Conn, error) // 创建新测连接
	Conns() int32           // 当前连接数
}

type client struct {
	running   int32
	closedCh  chan struct{}
	protoAddr string
	conf      *net.Dialer
	handle    CliHandler
	tick      bool
	conns     int32
	buffs     *buffs
}

func (c *client) ticker() {
	if !c.tick {
		return
	}
	var tm = time.NewTimer(c.handle.OnTick())
	defer tm.Stop()
	for atomic.LoadInt32(&c.running) == 1 {
		select {
		case <-c.closedCh:
			return
		case <-tm.C:
			tm.Reset(c.handle.OnTick())
		}
	}
}

// Serve 启动服务
func (c *client) Serve() error {
	if atomic.SwapInt32(&c.running, 1) != 0 {
		return ErrRunning
	}
	go c.ticker()
	return nil
}

// Shutdown 停止服务
func (c *client) Shutdown() error {
	if atomic.SwapInt32(&c.running, 0) != 1 {
		return ErrShutdown
	}
	close(c.closedCh)
	c.handle.OnShutdown(c)
	return nil
}

func (c *client) NewConn() (Conn, error) {
	var ctx context.Context
	var cancel func()
	if c.conf.Timeout <= 0 {
		ctx = context.Background()
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), c.conf.Timeout)
		defer cancel()
	}
	conn, err := Dial(ctx, c.protoAddr, c.conf)
	if err != nil {
		return nil, err
	}
	return newCliConn(c, conn), nil
}

// Conns 获取当前连接数
func (c *client) Conns() int32 {
	return atomic.LoadInt32(&c.conns)
}

// NewClient 返回一个新的服务对象
func NewClient(protoAddr string, handle CliHandler, opts ...CliOption) Client {
	cli := &client{
		running:   0,
		closedCh:  nil,
		protoAddr: protoAddr,
		conf:      new(net.Dialer),
		handle:    handle,
		tick:      false,
		conns:     0,
		buffs:     newBuffs(),
	}
	for _, opt := range opts {
		opt(cli)
	}
	return cli
}
