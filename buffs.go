package nets

import (
	"sync"
)

type buffs struct {
	cap  int32
	pool *sync.Pool
}

func (b *buffs) new() interface{} {
	return make([]byte, b.cap, b.cap)
}

func (b *buffs) Get() []byte {
	return b.pool.Get().([]byte)
}

func (b *buffs) Put(p []byte) {
	b.pool.Put(p)
}

func (b *buffs) reset(size int32) {
	b.cap = size
}

func newBuffs() *buffs {
	b := &buffs{
		cap: 4 << 10,
	}
	b.pool = &sync.Pool{New: b.new}
	return b
}
