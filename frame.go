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

// raw 是刚切下来、还没过校验的一段完整报文。
type raw struct {
	seq     uint16
	payload []byte
	sum     uint32
}

// cut 只负责拆帧：跳过噪声字节，把完整的报文切出来。
// 不够一条的半截尾巴原样留在 rest，不丢也不吞后面的数据。
func cut(buf []byte) (raws []raw, rest []byte) {
	for len(buf) > 0 {
		if buf[0] != 0xA5 {
			buf = buf[1:]
			continue
		}
		if len(buf) < 5 {
			break
		}
		n := int(binary.LittleEndian.Uint16(buf[3:5]))
		need := 5 + n + 4
		if len(buf) < need {
			break
		}
		raws = append(raws, raw{
			seq:     binary.LittleEndian.Uint16(buf[1:3]),
			payload: append([]byte(nil), buf[5:5+n]...),
			sum:     binary.LittleEndian.Uint32(buf[5+n : need]),
		})
		buf = buf[need:]
	}
	return raws, buf
}

// check 只负责校验：对切下来的报文逐条核对 CRC，坏的丢掉。
func check(r raw) bool {
	body := make([]byte, 0, 4+len(r.payload))
	body = binary.LittleEndian.AppendUint16(body, r.seq)
	body = binary.LittleEndian.AppendUint16(body, uint16(len(r.payload)))
	body = append(body, r.payload...)
	return crc32.ChecksumIEEE(body) == r.sum
}

// reseq 只负责序号：按切出来的原顺序发，不改写线上序号。
func reseq(r raw) Frame {
	return Frame{Seq: r.seq, Payload: r.payload}
}

// Split 把缓冲区里能看全的报文切出来，按原顺序返回，半截留在 rest。
func Split(buf []byte) (frames []Frame, rest []byte) {
	raws, rest := cut(buf)
	for _, r := range raws {
		if !check(r) {
			continue
		}
		frames = append(frames, reseq(r))
	}
	return frames, rest
}
