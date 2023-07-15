package nets

import (
	"net"
	"sync/atomic"
	"time"
)

type CliHandler interface {
	OnBoot(cli Client) (err error) // 启动客户端触发
	OnShutdown(cli Client)         // 关闭客户端触发
	OnTick() (dur time.Duration)   // 定时触发
	ConnHandler
}

type Client interface {
	Serve() error           // 启动服务
	Shutdown() error        // 关停服务
	NewConn() (Conn, error) // 创建新测连接
	Conns() int32           // 当前连接数
}

type client struct {
	running   int32
	closed    chan struct{}
	protoAddr string
	conf      *net.Dialer
	handler   CliHandler
	tick      bool
	conns     int32
	buffs     *Buffs
}

func (c *client) ticker() {
	if !c.tick {
		return
	}
	tk := time.NewTicker(c.handler.OnTick())
	defer tk.Stop()
BREAK:
	for {
		select {
		case <-c.closed:
			break BREAK
		case <-tk.C:
			tk.Reset(c.handler.OnTick())
		}
	}
}

func (c *client) Serve() error {
	if atomic.SwapInt32(&c.running, 1) != 0 {
		return ErrRunning
	}
	if err := c.handler.OnBoot(c); err != nil {
		_ = c.Shutdown()
		return err
	}
	c.closed = make(chan struct{})
	go c.ticker()
	return nil
}

func (c *client) Shutdown() error {
	if atomic.SwapInt32(&c.running, 0) != 1 {
		return ErrStopped
	}
	close(c.closed)
	c.handler.OnShutdown(c)
	return nil
}

func (c *client) NewConn() (Conn, error) {
	conn, err := dial(c.protoAddr, c.conf)
	if err != nil {
		return nil, err
	}
	return newConn(c.buffs, c.handler, &c.conns, conn), nil
}

func (c *client) Conns() int32 {
	return atomic.LoadInt32(&c.conns)
}

// NewClient 返回一个新的服务对象
func NewClient(protoAddr string, handler CliHandler, opts ...CliOption) Client {
	cli := &client{
		protoAddr: protoAddr,
		conf:      new(net.Dialer),
		handler:   handler,
		buffs:     NewBuffs(4 << 10),
	}
	for _, option := range opts {
		option(cli)
	}
	return cli
}
