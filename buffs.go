package nets

import (
	"bytes"
	"sync"
)

type Buffs struct {
	bufCap   uint32
	bufPool  *sync.Pool
	buffPool *sync.Pool
}

func (p *Buffs) newBuf() interface{} {
	return make([]byte, p.bufCap, p.bufCap)
}

func (p *Buffs) GetBuf() []byte {
	return p.bufPool.Get().([]byte)
}

func (p *Buffs) PutBuf(b []byte) {
	p.bufPool.Put(b)
}

func (p *Buffs) newBuff() interface{} {
	return bytes.NewBuffer(p.GetBuf()[:0])
}

func (p *Buffs) GetBuff() *bytes.Buffer {
	return p.buffPool.Get().(*bytes.Buffer)
}

func (p *Buffs) PutBuff(b *bytes.Buffer) {
	p.buffPool.Put(b)
}

func NewBuffs(c uint32) *Buffs {
	b := &Buffs{
		bufCap: c,
	}
	b.bufPool = &sync.Pool{New: b.newBuf}
	b.buffPool = &sync.Pool{New: b.newBuff}
	return b
}
