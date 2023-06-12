package nets

import (
	"net"
	"sync/atomic"
	"time"
)

type SvrHandler interface {
	ConnHandler
	OnBoot(svr Server) (err error) // 启动服务端触发
	OnShutdown(svr Server)         // 关闭服务端触发
	OnTick() (dur time.Duration)   // 定时触发
}

type Server interface {
	Serve() error
	Shutdown() error
	Conns() int32
}

// server 服务实现
type server struct {
	running   int32
	closedCh  chan struct{}
	protoAddr string
	conf      *net.ListenConfig
	listen    net.Listener
	handle    SvrHandler
	tick      bool
	conns     int32
	buffs     *buffs
}

func (s *server) ticker() {
	if !s.tick {
		return
	}
	var tm = time.NewTimer(s.handle.OnTick())
	defer tm.Stop()
	for atomic.LoadInt32(&s.running) == 1 {
		select {
		case <-s.closedCh:
			return
		case <-tm.C:
			tm.Reset(s.handle.OnTick())
		}
	}
}

// Serve 启动服务
func (s *server) Serve() (err error) {
	if atomic.SwapInt32(&s.running, 1) != 0 {
		return ErrRunning
	}
	if listen, e := Listen(s.protoAddr, s.conf); e != nil {
		atomic.StoreInt32(&s.running, 0)
		return e
	} else {
		s.listen = listen
		s.closedCh = make(chan struct{}, 0)
	}
	defer func() {
		if e := s.Shutdown(); e != nil && err == nil {
			err = e
		}
	}()
	if err = s.handle.OnBoot(s); err != nil {
		return err
	}
	go s.ticker()
	for true {
		c, e := s.listen.Accept()
		if e != nil {
			return e
		}
		go newSvrConn(s, c)
	}
	return err
}

// Shutdown 关停服务
func (s *server) Shutdown() (err error) {
	if atomic.SwapInt32(&s.running, 0) != 1 {
		return ErrShutdown
	}
	close(s.closedCh)
	s.handle.OnShutdown(s)
	return s.listen.Close()
}

// Conns 获取当前连接数
func (s *server) Conns() int32 {
	return atomic.LoadInt32(&s.conns)
}

// NewServer 返回一个新的服务对象
func NewServer(protoAddr string, handle SvrHandler, opts ...SvrOption) Server {
	svr := &server{
		running:   0,
		closedCh:  nil,
		protoAddr: protoAddr,
		conf:      new(net.ListenConfig),
		listen:    nil,
		handle:    handle,
		tick:      false,
		conns:     0,
		buffs:     newBuffs(),
	}
	for _, opt := range opts {
		opt(svr)
	}
	return svr
}
