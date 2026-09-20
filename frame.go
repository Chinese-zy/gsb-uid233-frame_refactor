package frame

import (
	"encoding/binary"
	"hash/crc32"
)

// Frame 是一条已经切开的报文。线上格式是起始字节 A5，序号两个字节，长度两个字节，载荷，然后四字节校验，都是小端。
type Frame struct {
	Seq     uint16
	Payload []byte
}

// Split 把缓冲区里能看全的报文切出来。半截留在 rest。
// 现在拆帧、校验和序号都堆在这一个函数里。
func Split(buf []byte) (frames []Frame, rest []byte) {
	var shift uint16
	for len(buf) > 0 {
		if buf[0] != 0xA5 {
			buf = buf[1:]
			continue
		}
		if len(buf) < 5 {
			break
		}
		seq := binary.LittleEndian.Uint16(buf[1:3])
		n := int(binary.LittleEndian.Uint16(buf[3:5]))
		need := 5 + n + 4
		if len(buf) < need {
			buf = nil
			continue
		}
		body := buf[1 : 5+n]
		got := binary.LittleEndian.Uint32(buf[5+n : need])
		if crc32.ChecksumIEEE(body) != got {
			shift++
			buf = buf[need:]
			continue
		}
		payload := append([]byte(nil), buf[5:5+n]...)
		frames = append(frames, Frame{Seq: seq + shift, Payload: payload})
		buf = buf[need:]
	}
	return frames, buf
}
