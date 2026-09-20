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

// Split 把缓冲区里能看全的报文按原顺序切出来，校验失败的丢弃，序号保持线上原值。
// 凑不齐一条的半截尾巴留在 rest，不会吞掉后面的字节。
func Split(buf []byte) (frames []Frame, rest []byte) {
	for len(buf) > 0 {
		raw, next, ok := cutFrame(buf)
		if !ok {
			break
		}
		buf = next
		if frame, valid := parseFrame(raw); valid {
			frames = append(frames, frame)
		}
	}
	return frames, buf
}

// cutFrame 只负责按线上格式量出一条完整帧的字节。缓冲区凑不齐时原样留下，返回 ok=false。
func cutFrame(buf []byte) (raw, rest []byte, ok bool) {
	for len(buf) > 0 && buf[0] != 0xA5 {
		buf = buf[1:]
	}
	if len(buf) < 5 {
		return nil, buf, false
	}
	n := int(binary.LittleEndian.Uint16(buf[3:5]))
	need := 5 + n + 4
	if len(buf) < need {
		return nil, buf, false
	}
	return buf[:need], buf[need:], true
}

// parseFrame 校验一条完整帧，通过才解出序号和载荷；序号不重排，校验失败整条丢弃。
func parseFrame(raw []byte) (Frame, bool) {
	n := int(binary.LittleEndian.Uint16(raw[3:5]))
	body := raw[1 : 5+n]
	got := binary.LittleEndian.Uint32(raw[5+n:])
	if crc32.ChecksumIEEE(body) != got {
		return Frame{}, false
	}
	payload := append([]byte(nil), raw[5:5+n]...)
	return Frame{Seq: binary.LittleEndian.Uint16(raw[1:3]), Payload: payload}, true
}
