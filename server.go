package nets

import (
	"net"
	"sync/atomic"
	"time"
)

type SvrHandler interface {
	OnBoot(svr Server) (err error) // 启动服务端触发
	OnShutdown(svr Server)         // 关闭服务端触发
	OnTick() (dur time.Duration)   // 定时触发
	ConnHandler
}

type Server interface {
	Serve() error
	Shutdown() error
	Conns() int32
}

type server struct {
	running   int32
	closed    chan struct{}
	protoAddr string
	conf      *net.ListenConfig
	listener  net.Listener
	handler   SvrHandler
	tick      bool
	conns     int32
	buffs     *Buffs
}

func (s *server) ticker() {
	if !s.tick {
		return
	}
	tk := time.NewTicker(s.handler.OnTick())
	defer tk.Stop()
BREAK:
	for {
		select {
		case <-s.closed:
			break BREAK
		case <-tk.C:
			tk.Reset(s.handler.OnTick())
		}
	}
}

func (s *server) Serve() error {
	if atomic.SwapInt32(&s.running, 1) != 0 {
		return ErrRunning
	}
	if listener, err := listenAddr(s.protoAddr, s.conf); err != nil {
		atomic.StoreInt32(&s.running, 0)
		return err
	} else {
		s.listener = listener
	}
	defer s.Shutdown()
	if err := s.handler.OnBoot(s); err != nil {
		return err
	}
	s.closed = make(chan struct{})
	defer close(s.closed)
	go s.ticker()
	var (
		conn net.Conn
		err  error
	)
BREAK:
	for atomic.LoadInt32(&s.running) == 1 {
		conn, err = s.listener.Accept()
		if err != nil {
			break BREAK
		}
		_ = newConn(s.buffs, s.handler, &s.conns, conn)
	}
	return err
}

func (s *server) Shutdown() error {
	if atomic.SwapInt32(&s.running, 0) != 1 {
		return ErrStopped
	}
	defer atomic.StoreInt32(&s.conns, 0)
	err := s.listener.Close()
	s.handler.OnShutdown(s)
	return err
}

func (s *server) Conns() int32 {
	return atomic.LoadInt32(&s.running)
}

// NewServer 返回一个新的服务对象
func NewServer(protoAddr string, handler SvrHandler, opts ...SvrOption) Server {
	svr := &server{
		protoAddr: protoAddr,
		conf:      new(net.ListenConfig),
		handler:   handler,
		buffs:     NewBuffs(4 << 10),
	}
	for _, option := range opts {
		option(svr)
	}
	return svr
}
