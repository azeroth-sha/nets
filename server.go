package nets

import (
	"net"
	"sync/atomic"
)

type Server struct {
	running   int32
	protoAddr string
	listen    net.Listener
}

func (s *Server) Serve() error {
	if atomic.SwapInt32(&s.running, 1) != 0 {
		return ErrRunning
	}
	return nil
}
