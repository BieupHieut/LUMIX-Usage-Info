//go:build windows

package main

import (
	"encoding/binary"
	"fmt"
)

func checkedServiceFrames(raw []byte) ([]TLV, error) {
	var out []TLV
	replies := 0
	for off := 0; off < len(raw); {
		if len(raw)-off < 8 {
			return nil, fmt.Errorf("truncated service reply header")
		}
		tag := binary.LittleEndian.Uint32(raw[off : off+4])
		n := binary.LittleEndian.Uint32(raw[off+4 : off+8])
		if uint64(n) > uint64(len(raw)-off-8) {
			return nil, fmt.Errorf("truncated service reply payload")
		}
		if tag == SETUP_INFO_REPLY {
			replies++
			if replies > 1 {
				return nil, fmt.Errorf("ambiguous duplicate service replies")
			}
			if n < 146 {
				return nil, fmt.Errorf("service counter payload too short")
			}
		}
		out = append(out, TLV{tag, raw[off+8 : off+8+int(n)]})
		off += 8 + int(n)
	}
	return out, nil
}
