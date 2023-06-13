package nets

import (
	"bytes"
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
	conn    net.Conn
	closed  int32
	reading int32
	writMu  sync.Mutex
	flag    *int32
	buffer  *bytes.Buffer
	err     error
	handle  ConnHandler
	cnt     *int32
	ctx     interface{}
	buffs   *buffs
}

func (c *connection) read() (int, error) {
	if atomic.SwapInt32(&c.reading, 1) != 0 {
		return 0, nil
	}
	if c.buffer == nil {
		c.buffer = c.buffs.GetCache()
	}
	buf := c.buffs.Get()
	defer c.buffs.Put(buf)
	var (
		cnt int
		err error
	)
	for {
		n, e := c.conn.Read(buf[:])
		if n > 0 {
			cnt += n
			c.buffer.Write(buf[:n])
		} else if n <= 0 || e != nil {
			err = e
			break
		}
	}
	atomic.StoreInt32(&c.reading, 0)
	return cnt, err
}

// Read 从缓冲中读出数据(由于为非阻塞，以io.EOF为读取结束标记)
func (c *connection) Read(b []byte) (n int, err error) {
	if c.IsClosed() {
		return 0, net.ErrClosed
	} else if atomic.SwapInt32(&c.reading, 1) != 0 {
		return 0, nil
	}
	if c.buffer != nil {
		if n, err = c.buffer.Read(b); err == io.EOF {
			c.buffs.PutCache(c.buffer)
			c.buffer = nil
		}
	}
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
	atomic.AddInt32(c.cnt, -1)
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

func (c *connection) run() {
	defer c.Close()
	atomic.AddInt32(c.cnt, 1)
	if c.handle.OnOpened(c) != nil {
		return
	}
	var cnt int
	var tk = time.NewTimer(time.Millisecond * 15)
	defer tk.Stop()
	for !c.IsClosed() && atomic.LoadInt32(c.flag) == 1 {
		if cnt, c.err = c.read(); c.err != nil {
			break
		} else if cnt > 0 {
			if c.err = c.handle.OnActivate(c); c.err != nil {
				break
			}
		} else {
			<-tk.C
		}
	}
}

func newSvrConn(svr *server, conn net.Conn) {
	c := &connection{
		conn:    conn,
		closed:  0,
		reading: 0,
		writMu:  sync.Mutex{},
		flag:    &svr.running,
		buffer:  new(bytes.Buffer),
		err:     nil,
		handle:  svr.handle,
		cnt:     &svr.conns,
		ctx:     nil,
		buffs:   svr.buffs,
	}
	go c.run()
}

func newCliConn(cli *client, conn net.Conn) *connection {
	c := &connection{
		conn:    conn,
		closed:  0,
		reading: 0,
		writMu:  sync.Mutex{},
		flag:    &cli.running,
		buffer:  new(bytes.Buffer),
		err:     nil,
		handle:  cli.handle,
		cnt:     &cli.conns,
		ctx:     nil,
		buffs:   cli.buffs,
	}
	go c.run()
	return c
}
