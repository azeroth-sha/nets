package nets

import (
	"bytes"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Conn interface {
	Read(b []byte) (n int, err error)   // 读取数据
	Write(b []byte) (n int, err error)  // 写入数据
	Close() error                       // 关闭连接
	LocalAddr() net.Addr                // 返回本地网络地址
	RemoteAddr() net.Addr               // 返回远端网络地址
	SetDeadline(t time.Time) error      // 设置综合超时时间
	SetReadDeadline(t time.Time) error  // 设置读取超时时间
	SetWriteDeadline(t time.Time) error // 设置写入超时时间
	IsClosed() bool                     // 连接是否已关闭
	Context() interface{}               // 获取上下文信息
	SetContext(ctx interface{})         // 设置上下文信息
}

type connection struct {
	conn    net.Conn
	closed  int32
	reading int32
	writMu  sync.Mutex
	flag    *int32
	buffer  *bytes.Buffer
	err     error
	handle  Handler
	conns   *int32
	ctx     interface{}
}

func (c *connection) Read(b []byte) (n int, err error) {
	if c.IsClosed() {
		return 0, net.ErrClosed
	} else if atomic.SwapInt32(&c.reading, 1) != 0 {
		return 0, nil
	}
	n, err = c.buffer.Read(b)
	atomic.StoreInt32(&c.reading, 0)
	return n, err
}

func (c *connection) Write(b []byte) (n int, err error) {
	if c.IsClosed() {
		return 0, net.ErrClosed
	}
	c.writMu.Lock()
	n, err = c.conn.Write(b)
	c.writMu.Unlock()
	return n, err
}

func (c *connection) Close() error {
	if atomic.SwapInt32(&c.closed, 1) != 0 {
		return net.ErrClosed
	}
	atomic.AddInt32(c.conns, -1)
	if err := c.conn.Close(); err != nil && c.err == nil {
		c.err = err
	}
	c.handle.OnClosed(c, c.err)
	return c.err
}

func (c *connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *connection) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *connection) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *connection) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *connection) IsClosed() bool {
	return atomic.LoadInt32(&c.closed) == 1
}

func (c *connection) Context() interface{} {
	return c.ctx
}

func (c *connection) SetContext(ctx interface{}) {
	c.ctx = ctx
}

func (c *connection) read() (int, error) {
	if atomic.SwapInt32(&c.reading, 1) != 0 {
		return 0, nil
	}
	buf := getBuf()
	n, err := c.conn.Read(buf[:])
	if n > 0 {
		c.buffer.Write(buf[:n])
	}
	atomic.StoreInt32(&c.reading, 0)
	putBuf(buf)
	return n, err
}

func (c *connection) run() {
	defer c.Close()
	atomic.AddInt32(c.conns, 1)
	if c.handle.OnOpened(c) != nil {
		return
	}
	var cnt int
	for !c.IsClosed() && atomic.LoadInt32(c.flag) == 1 {
		if cnt, c.err = c.read(); c.err != nil {
			break
		} else if cnt > 0 {
			if c.err = c.handle.OnActivate(c); c.err != nil {
				break
			}
		}
	}
}

func newConn(handle Handler, conns *int32, flag *int32, conn net.Conn) {
	c := &connection{
		conn:    conn,
		closed:  0,
		reading: 0,
		writMu:  sync.Mutex{},
		flag:    flag,
		buffer:  new(bytes.Buffer),
		err:     nil,
		handle:  handle,
		conns:   conns,
	}
	go c.run()
}
