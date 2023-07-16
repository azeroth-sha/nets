package main

import (
	"github.com/azeroth-sha/nets"
	"log"
	"sync"
	"time"
)

var (
	protoAddr = "tcp://:10000"
)

type handle struct {
	svr   nets.Server
	conns *sync.Map
}

func (h *handle) OnBoot(svr nets.Server) (err error) {
	h.svr = svr
	return nil
}

func (h *handle) OnShutdown(_ nets.Server) {
	h.conns.Range(func(key, value interface{}) bool {
		log.Printf("%v %v\r\n", key, value.(nets.Conn).Close())
		return true
	})
}

func (h *handle) OnTick() (dur time.Duration) {
	var count int
	h.conns.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	log.Printf("connections: %d", count)
	return time.Second
}

func (h *handle) OnOpened(conn nets.Conn) (err error) {
	log.Printf("opened: %s", conn.RemoteAddr().String())
	h.conns.Store(conn.RemoteAddr().String(), conn)
	return nil
}

func (h *handle) OnClosed(conn nets.Conn, err error) {
	log.Printf("closed: %s %v", conn.RemoteAddr().String(), err)
	h.conns.Delete(conn.RemoteAddr().String())
}

func (h *handle) OnActivate(conn nets.Conn) (err error) {
	buf := make([]byte, 4096, 4096)
	if n, err := conn.Read(buf); err != nil {
		log.Panicln(err)
	} else {
		if _, err2 := conn.Write(buf[:n]); err2 != nil {
			log.Println(err2)
		}
		log.Printf("count: %d [%s]", n, buf[:n])
	}
	return nil
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	wg := new(sync.WaitGroup)
	svr := nets.NewServer(
		protoAddr,
		&handle{conns: new(sync.Map)},
		nets.WithSvrTick(true),
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := svr.Serve(); err != nil {
			log.Println(err)
		}
	}()
	time.Sleep(time.Second * 15)
	if err := svr.Shutdown(); err != nil {
		log.Println(err)
	}
}
