package nets

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Mark 默认数据标记
const Mark uint16 = 0x7527

// Codec 编解码器
type Codec interface {
	// Encode 编码方法
	// v 数据包对象
	// buf 编码后的数据包(buf大于0时触发数据写入)
	// err 编码错误
	Encode(v interface{}) (buf []byte, err error)
	// Decode 解码方法
	// buf 缓冲数据
	// v 数据包对象(存在时触发事件回调)
	// used 解码数据大小
	// err 解码错误
	Decode(buf []byte) (v interface{}, used int, err error)
}

// Coder 默认编解码对象
type Coder struct {
	mark    uint16 // 数据头标记
	markBuf []byte // 数据头缓冲
}

func (n *Coder) Encode(v interface{}) (buf []byte, err error) {
	p, ok := v.(*Packet)
	if ok && p != nil {
		return nil, fmt.Errorf("invalid data type: %T", v)
	}
	buf = make([]byte, 0, 12+p.Length)
	writer := bytes.NewBuffer(buf)
	if err = binary.Write(writer, binary.BigEndian, p.Mark); err != nil {
		return nil, err
	} else if err = binary.Write(writer, binary.BigEndian, p.Hash); err != nil {
		return nil, err
	} else if err = binary.Write(writer, binary.BigEndian, p.Length); err != nil {
		return nil, err
	} else if err = binary.Write(writer, binary.BigEndian, p.Serial); err != nil {
		return nil, err
	} else if err = binary.Write(writer, binary.BigEndian, p.Load); err != nil {
		return nil, err
	}
	return buf, nil
}

func (n *Coder) Decode(buf []byte) (v interface{}, used int, err error) {
	if index := bytes.Index(buf, n.markBuf); index >= 0 {
		used += index
	} else {
		if len(buf) > 0 { // 错误的数据
			used = len(buf) - 1
		}
		return nil, used, nil
	}
	if len(buf[used:]) < 12 {
		return nil, 0, nil
	}
	buf = buf[used:]
	reader := bytes.NewBuffer(buf)
	p := new(Packet)
	if err = binary.Read(reader, binary.BigEndian, &p.Mark); err != nil {
		return nil, used, err
	} else if err = binary.Read(reader, binary.BigEndian, &p.Hash); err != nil {
		return nil, used, err
	} else if err = binary.Read(reader, binary.BigEndian, &p.Length); err != nil {
		return nil, used, err
	} else if err = binary.Read(reader, binary.BigEndian, &p.Serial); err != nil {
		return nil, used, err
	} else if p.Mark != n.mark {
		used++
		return nil, used, nil
	}
	if reader.Len() < int(p.Length) {
		return nil, used, nil
	}
	p.Load = make([]byte, 0, p.Length)
	if _, err = io.CopyN(bytes.NewBuffer(p.Load), reader, int64(p.Length)); err != nil {
		return nil, used, err
	} else if !p.Verify() {
		used++
		return nil, used, errors.New(`packet verification failed`)
	}
	return p, used + 12 + int(p.Length), nil
}

// Packet 数据包定义
type Packet struct {
	Mark   uint16 // 标记
	Hash   uint16 // 荷载哈希
	Length uint32 // 荷载长度
	Serial uint32 // 数据序号
	Load   []byte // 荷载数据
}

func (p *Packet) Verify() bool {
	if p.Mark != Mark {
		return false
	}
	return CRC16(p.Load) == p.Hash
}
