package frame

import (
	"encoding/binary"
	"hash/crc32"
	"testing"
)

func put(seq uint16, payload []byte, bad bool) []byte {
	buf := make([]byte, 5+len(payload)+4)
	buf[0] = 0xA5
	binary.LittleEndian.PutUint16(buf[1:3], seq)
	binary.LittleEndian.PutUint16(buf[3:5], uint16(len(payload)))
	copy(buf[5:], payload)
	sum := crc32.ChecksumIEEE(buf[1 : 5+len(payload)])
	if bad {
		sum ^= 0xFF
	}
	binary.LittleEndian.PutUint32(buf[5+len(payload):], sum)
	return buf
}

func TestTwoFramesKeepOrder(t *testing.T) {
	buf := append(put(1, []byte("甲"), false), put(2, []byte("乙"), false)...)
	frames, rest := Split(buf)
	if len(rest) != 0 || len(frames) != 2 {
		t.Fatalf("frames %d rest %d", len(frames), len(rest))
	}
	if frames[0].Seq != 1 || string(frames[0].Payload) != "甲" {
		t.Fatalf("first %+v", frames[0])
	}
	if frames[1].Seq != 2 || string(frames[1].Payload) != "乙" {
		t.Fatalf("second %+v", frames[1])
	}
}

func TestShortTailDoesNotSwallowNext(t *testing.T) {
	good := put(3, []byte("丙"), false)
	partial := []byte{0xA5, 0x04, 0x00, 0x14, 0x00}
	buf := append(append([]byte{}, partial...), good...)
	frames, rest := Split(buf)
	if len(frames) != 0 {
		t.Fatalf("半截不应切成帧 %+v", frames)
	}
	if len(rest) < len(good) || string(rest[len(rest)-len(good):]) != string(good) {
		t.Fatalf("下一条被吞了 rest %x", rest)
	}
}

func TestBadChecksumDoesNotShiftLaterSeq(t *testing.T) {
	buf := append(put(1, []byte("坏"), true), put(2, []byte("好"), false)...)
	frames, rest := Split(buf)
	if len(rest) != 0 || len(frames) != 1 {
		t.Fatalf("frames %d rest %d", len(frames), len(rest))
	}
	if frames[0].Seq != 2 || string(frames[0].Payload) != "好" {
		t.Fatalf("got %+v", frames[0])
	}
}
