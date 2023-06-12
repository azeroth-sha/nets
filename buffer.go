package nets

import "sync"

var bufPool = sync.Pool{New: newBuf}

func newBuf() interface{} {
	return make([]byte, 256, 256)
}

func getBuf() []byte {
	return bufPool.Get().([]byte)
}

func putBuf(b []byte) {
	bufPool.Put(b)
}
