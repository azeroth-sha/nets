package main

import (
	"github.com/azeroth-sha/nets"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

var (
	protoAddr = "tcp://localhost:10000"
)

type handle struct {
	cli   nets.Client
	conns *sync.Map
}

func (h *handle) OnBoot(cli nets.Client) (err error) {
	h.cli = cli
	return nil
}

func (h *handle) OnShutdown(cli nets.Client) {
	h.conns.Range(func(key, value interface{}) bool {
		log.Printf("%v %v\r\n", key, value.(nets.Conn).Close())
		return true
	})
}

func (h *handle) OnTick() (dur time.Duration) {
	var count int
	var b = []byte("hello world!")
	h.conns.Range(func(_, value interface{}) bool {
		if _, err := value.(nets.Conn).Write(b); err != nil {
			log.Println(err)
		}
		count++
		return true
	})
	log.Printf("connections: %d", count)
	if count == 0 {
		if _, err := h.cli.NewConn(); err != nil {
			log.Println(err)
		}
	}
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
	go func() {
		time.Sleep(time.Millisecond * 15)
		buf := make([]byte, 4096, 4096)
		if n, err := conn.Read(buf); err != nil {
			log.Panicln(err)
		} else {
			if n == 0 {
				log.Printf("count: %d [%s]\r\n%s", n, buf[:n], debug.Stack())
			} else {
				log.Printf("count: %d [%s]", n, buf[:n])
			}

		}
	}()
	return nil
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	svr := nets.NewClient(
		protoAddr,
		&handle{conns: new(sync.Map)},
		nets.WithCliTick(true),
	)
	if err := svr.Serve(); err != nil {
		log.Println(err)
	}
	time.Sleep(time.Second * 15)
	if err := svr.Shutdown(); err != nil {
		log.Println(err)
	}
}
