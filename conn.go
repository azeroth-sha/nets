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
	Conn() net.Conn                     // 返回原始连接
	Close() error                       // 关闭连接
	LocalAddr() net.Addr                // 返回本地网络地址
	RemoteAddr() net.Addr               // 返回远端网络地址
	SetDeadline(t time.Time) error      // 设置综合超时时间
	SetReadDeadline(t time.Time) error  // 设置读取超时时间
	SetWriteDeadline(t time.Time) error // 设置写入超时时间
	IsClosed() bool                     // 连接是否已关闭
}

type connection struct {
	conn    net.Conn
	closed  int32
	reading int32
	writMu  sync.Mutex
	flag    *int32
	buf     *bytes.Buffer
	err     error
	handle  Handler
	conns   *int32
}

func (c *connection) Read(b []byte) (n int, err error) {
	if c.IsClosed() {
		return 0, net.ErrClosed
	}
	if _, err = c.read(); err != nil {
		return 0, err
	}
	if atomic.SwapInt32(&c.reading, 1) != 0 {
		return 0, nil
	}
	n, err = c.buf.Read(b)
	atomic.StoreInt32(&c.reading, 0)
	return n, err
}

func (c *connection) Write(b []byte) (n int, err error) {
	if c.IsClosed() {
		return 0, net.ErrClosed
	}
	c.writMu.Lock()
	n, err = c.buf.Write(b)
	c.writMu.Unlock()
	return n, err
}

func (c *connection) Conn() net.Conn {
	return c.conn
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

func (c *connection) read() (int64, error) {
	if atomic.SwapInt32(&c.reading, 1) != 0 {
		return 0, nil
	}
	n, err := c.buf.ReadFrom(c.conn)
	atomic.StoreInt32(&c.reading, 0)
	return n, err
}

func (c *connection) run() {
	defer c.Close()
	atomic.AddInt32(c.conns, 1)
	if c.handle.OnOpened(c) != nil {
		return
	}
	for !c.IsClosed() && atomic.LoadInt32(c.flag) == 1 {
		if cnt, err := c.read(); err != nil {
			c.err = err
			return
		} else if cnt > 0 {
			if c.err = c.handle.OnActivate(c); c.err != nil {
				return
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
		buf:     new(bytes.Buffer),
		err:     nil,
		handle:  handle,
		conns:   conns,
	}
	go c.run()
}
