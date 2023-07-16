package nets

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type ConnHandler interface {
	OnOpened(conn Conn) (err error)   // 打开连接触发
	OnClosed(conn Conn, err error)    // 关闭连接触发
	OnActivate(conn Conn) (err error) // 收到数据触发
}

type Conn interface {
	net.Conn
	IsClosed() bool             // 连接是否已关闭
	Context() interface{}       // 获取上下文信息
	SetContext(ctx interface{}) // 设置上下文信息
}

type connection struct {
	handler  ConnHandler
	closed   int32
	conn     net.Conn
	readMu   *sync.Mutex
	readBuff *bytes.Buffer
	writMu   *sync.Mutex
	ctx      *atomic.Value
	buffs    *Buffs
	count    *int32
}

func (c *connection) Read(b []byte) (n int, err error) {
	if c.IsClosed() {
		return 0, ErrClosed
	}
	c.readMu.Lock()
	defer c.readMu.Unlock()
	if c.readBuff == nil {
		return 0, nil
	}
	n, err = c.readBuff.Read(b)
	if err == io.EOF {
		c.buffs.PutBuff(c.readBuff)
		c.readBuff = nil
		return n, nil
	}
	return n, err
}

func (c *connection) Write(b []byte) (n int, err error) {
	if c.IsClosed() {
		return 0, ErrClosed
	}
	c.writMu.Lock()
	defer c.writMu.Unlock()
	return c.conn.Write(b)
}

func (c *connection) Close() error {
	if atomic.SwapInt32(&c.closed, 1) != 0 {
		return ErrClosed
	}
	return c.conn.Close()
}

func (c *connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *connection) SetDeadline(t time.Time) error {
	if c.IsClosed() {
		return ErrClosed
	}
	return c.conn.SetDeadline(t)
}

func (c *connection) SetReadDeadline(t time.Time) error {
	if c.IsClosed() {
		return ErrClosed
	}
	return c.conn.SetReadDeadline(t)
}

func (c *connection) SetWriteDeadline(t time.Time) error {
	if c.IsClosed() {
		return ErrClosed
	}
	return c.conn.SetWriteDeadline(t)
}

func (c *connection) IsClosed() bool {
	return atomic.LoadInt32(&c.closed) == 1
}

func (c *connection) Context() interface{} {
	return c.ctx.Load()
}

func (c *connection) SetContext(ctx interface{}) {
	c.ctx.Store(ctx)
}

func (c *connection) append(b []byte) error {
	if c.IsClosed() {
		return ErrClosed
	}
	c.readMu.Lock()
	defer c.readMu.Unlock()
	if len(b) == 0 {
		return nil
	}
	if c.readBuff == nil {
		c.readBuff = c.buffs.GetBuff()
	}
	_, err := c.readBuff.Write(b)
	return err
}

func (c *connection) read() error {
	if c.IsClosed() {
		return ErrClosed
	}
	buf := c.buffs.GetBuf()
	defer c.buffs.PutBuf(buf)
	n, err := c.conn.Read(buf)
	if n > 0 {
		if err2 := c.append(buf[:n]); err2 != nil && err == nil {
			err = err2
		}
	}
	if n > 0 && err == io.EOF {
		return nil
	} else {
		return err
	}
}

func (c *connection) run() {
	var err error
	atomic.AddInt32(c.count, 1)
	defer func() {
		_ = c.Close()
		_ = c.callback(c.handler.OnClosed, err)
		atomic.AddInt32(c.count, -1)
	}()
	if err = c.callback(c.handler.OnOpened, nil); err != nil {
		return
	}
	for err = c.read(); err == nil && !c.IsClosed(); err = c.read() {
		if err = c.callback(c.handler.OnActivate, nil); err != nil {
			break
		}
	}
}

func (c *connection) callback(f interface{}, e error) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("%v", rec)
		}
	}()
	switch call := f.(type) {
	case func(Conn) error:
		return call(c)
	case func(Conn, error):
		call(c, e)
	}
	return nil
}

func newConn(buffs *Buffs, handler ConnHandler, count *int32, conn net.Conn) Conn {
	c := &connection{
		handler: handler,
		conn:    conn,
		readMu:  new(sync.Mutex),
		writMu:  new(sync.Mutex),
		ctx:     new(atomic.Value),
		buffs:   buffs,
		count:   count,
	}
	go c.run()
	return c
}
