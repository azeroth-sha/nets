package nets

import (
	"bytes"
	"sync"
)

type buffs struct {
	cap   int32
	pool  *sync.Pool
	cache *sync.Pool
}

func (b *buffs) newBuf() interface{} {
	return make([]byte, b.cap, b.cap)
}

func (b *buffs) Get() []byte {
	return b.pool.Get().([]byte)
}

func (b *buffs) Put(p []byte) {
	b.pool.Put(p)
}

func (b *buffs) newCache() interface{} {
	return bytes.NewBuffer(nil)
}

func (b *buffs) GetCache() *bytes.Buffer {
	return b.cache.Get().(*bytes.Buffer)
}

func (b *buffs) PutCache(p *bytes.Buffer) {
	b.cache.Put(p)
}

func (b *buffs) reset(size int32) {
	b.cap = size
}

func newBuffs() *buffs {
	b := &buffs{
		cap: 4 << 10,
	}
	b.pool = &sync.Pool{New: b.newBuf}
	b.cache = &sync.Pool{New: b.newCache}
	return b
}
