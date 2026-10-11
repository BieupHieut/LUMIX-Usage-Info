//go:build windows

package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"testing"
)

func statusFrame(tag uint32, data []byte) []byte {
	b := make([]byte, 8+len(data))
	binary.LittleEndian.PutUint32(b, tag)
	binary.LittleEndian.PutUint32(b[4:], uint32(len(data)))
	copy(b[8:], data)
	return b
}
func statusWords(tag uint32, values ...uint16) []byte {
	b := make([]byte, len(values)*2)
	for i, v := range values {
		binary.LittleEndian.PutUint16(b[i*2:], v)
	}
	return statusFrame(tag, b)
}
func statusU32(tag, v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return statusFrame(tag, b)
}
func statusStorageFixture(capacity, free uint64) []byte {
	b := make([]byte, 26)
	binary.LittleEndian.PutUint16(b, 3)
	binary.LittleEndian.PutUint16(b[2:], 2)
	binary.LittleEndian.PutUint64(b[6:], capacity)
	binary.LittleEndian.PutUint64(b[14:], free)
	binary.LittleEndian.PutUint32(b[22:], 1234)
	// Private user-defined card labels must never appear in exports.
	for _, s := range []string{"PRIVATE_CARD_NAME", "PRIVATE_VOLUME_LABEL"} {
		b = append(b, byte(len(s)+1))
		for _, c := range s {
			b = append(b, byte(c), 0)
		}
		b = append(b, 0, 0)
	}
	return b
}
func statusDetailFixture(n int) []byte {
	b := make([]byte, 26+16*n)
	binary.LittleEndian.PutUint16(b, uint16(n))
	for off, v := range map[int]uint32{2: 1200, 6: 42, 10 + 8*n: 7200, 14 + 8*n: 300} {
		binary.LittleEndian.PutUint32(b[off:], v)
	}
	return statusFrame(0x12000016, b)
}
func TestCameraStatusResponseTagsAndMissingValues(t *testing.T) {
	raw := append(statusU32(0x12000011, 9745), statusU32(0x12000012, 9960)...)
	raw = append(raw, statusWords(0x12000013, 0)...)
	s, photos, seconds, e := boardParseCameraStatus(raw)
	if e != nil || s == nil || *s != 0 || *photos != 9745 || *seconds != 9960 {
		t.Fatal(s, photos, seconds, e)
	}
	if boardCameraStateKey(0) != "state_standby" || boardCameraStateKey(0x1234) != "state_unknown" {
		t.Fatal("unknown state guessed")
	}
	s, photos, seconds, e = boardParseCameraStatus(statusU32(0x12000011, 0xffffffff))
	if e != nil || s != nil || photos != nil || seconds != nil {
		t.Fatal("absent fields became zero")
	}
	_, photos, _, e = boardParseCameraStatus(statusU32(0x12000011, 0))
	if e != nil || photos == nil || *photos != 0 {
		t.Fatal("valid zero lost")
	}
	for _, bad := range [][]byte{nil, raw[:len(raw)-1], append(append([]byte{}, raw...), raw...), statusWords(0x12000011, 3)} {
		if _, _, _, e = boardParseCameraStatus(bad); e == nil {
			t.Fatal("bad response accepted")
		}
	}
}
func TestCameraCardsAndDetailedSpaceBounds(t *testing.T) {
	v, e := boardParseCards(statusWords(0x12000015, 2, 1, 0))
	if e != nil || len(v) != 2 || boardCardStateKey(v[0]) != "card_ready" || boardCardStateKey(v[1]) != "card_absent" {
		t.Fatal(v, e)
	}
	if _, e = boardParseCards(statusWords(0x12000015, 0)); e != nil {
		t.Fatal(e)
	}
	for _, bad := range [][]byte{statusWords(0x12000015, 2, 1), statusWords(0x12000015, 17), statusWords(0x12000016, 0)} {
		if _, e = boardParseCards(bad); e == nil {
			t.Fatal("bad card count/tag")
		}
	}
	for _, n := range []int{0, 1, 2, 16} {
		d, e := boardParseFreeDetail(statusDetailFixture(n))
		if e != nil || int(d.Count) != n || d.Images != 1200 || d.Seconds != 7200 || d.Seconds4K != 300 {
			t.Fatal(n, d, e)
		}
		for _, a := range d.Arrays {
			if len(a) != n {
				t.Fatal("array length")
			}
		}
		b := statusDetailFixture(n)
		if _, e = boardParseFreeDetail(b[:len(b)-1]); e == nil {
			t.Fatal("truncated detail")
		}
	}
	if _, e = boardParseFreeDetail(statusDetailFixture(17)); e == nil {
		t.Fatal("unbounded count")
	}
}
func TestCameraStorageCapacityAndPrivacy(t *testing.T) {
	for _, tc := range []struct {
		total, free uint64
		known       bool
	}{{128000000000, 96000000000, true}, {128, 0, true}, {128, 129, false}, {0, 0, false}, {^uint64(0), 0, false}, {128, ^uint64(0), false}} {
		s, e := boardStorageInfo(0x10001, statusStorageFixture(tc.total, tc.free))
		if e != nil || s.spaceKnown() != tc.known {
			t.Fatal(s, e)
		}
	}
	b := statusStorageFixture(128, 90)
	for i := 0; i < len(b); i++ {
		if _, e := boardStorageInfo(1, b[:i]); e == nil {
			t.Fatal("truncated storage", i)
		}
	}
	for _, bad := range [][]byte{nil, {2, 0, 0, 0, 1, 0, 0, 0}, {1, 0, 0, 0, 0, 0, 0, 0}, {17, 0, 0, 0}, {2, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0}} {
		if _, e := boardStorageIDs(bad); e == nil {
			t.Fatal("bad storage IDs accepted")
		}
	}
	s, _ := boardStorageInfo(1, b)
	d := newBoardDiagnostic(false)
	d.Camera = &boardCameraStatus{StorageRead: true, Stores: []boardCameraStorage{s}}
	payload, e := boardDiagnosticZIP(d)
	if e != nil {
		t.Fatal(e)
	}
	z, e := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if e != nil {
		t.Fatal(e)
	}
	for _, f := range z.File {
		r, _ := f.Open()
		v, _ := io.ReadAll(r)
		r.Close()
		if bytes.Contains(v, []byte("PRIVATE_")) {
			t.Fatal("private card label exported")
		}
	}
}

type statusFake struct {
	boardFake
	storageFail bool
}

func (f *statusFake) SendReadPTP(op uint32, args []uint32) ([]byte, uint32, error) {
	if op == 0x1004 {
		f.calls = append(f.calls, "1004")
		return []byte{2, 0, 0, 0, 1, 0, 1, 0, 2, 0, 1, 0}, 0x2001, nil
	}
	if op == 0x1005 {
		f.calls = append(f.calls, fmt.Sprintf("1005:%X", args))
		if f.storageFail && args[0] == 0x10001 {
			return nil, 0x2002, nil
		}
		return statusStorageFixture(128000000000, 96000000000), 0x2001, nil
	}
	if op == 0x9414 && len(args) == 1 {
		var b []byte
		switch args[0] {
		case 0x12000010:
			b = append(statusU32(0x12000011, 1234), statusWords(0x12000013, 0)...)
		case 0x12000015:
			b = statusWords(0x12000015, 2, 1, 2)
		case 0x12000016:
			b = statusDetailFixture(2)
		}
		if b != nil {
			f.calls = append(f.calls, fmt.Sprintf("9414:%X", args))
			return b, 0x2001, nil
		}
	}
	return f.boardFake.SendReadPTP(op, args)
}
func TestCameraOptionalQueriesIndependentAndExportable(t *testing.T) {
	for _, usageOK := range []bool{true, false} {
		f := &statusFake{boardFake: boardFake{model: "DC-S5", usageRC: 0x2001}, storageFail: true}
		if !usageOK {
			f.usageRC = 0x2002
		}
		out := boardQueryOptions(f, false, true, func(boardReply) {})
		if out.Valid != usageOK || out.Camera == nil || !out.Camera.CardsRead || len(out.Camera.Stores) != 2 || out.Camera.Stores[0].Read || !out.Camera.Stores[1].Read || f.closeCount != 1 {
			t.Fatal(out, f.calls)
		}
		if strings.Count(strings.Join(f.calls, ";"), "1004") != 1 || len(f.calls) != 6 {
			t.Fatal("unexpected retries or operations", f.calls)
		}
		payload, e := boardDiagnosticZIP(out.Diagnostic)
		if e != nil {
			t.Fatal(e)
		}
		z, _ := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
		found := false
		for _, file := range z.File {
			if file.Name == "diagnostic.json" {
				r, _ := file.Open()
				b, _ := io.ReadAll(r)
				r.Close()
				found = bytes.Contains(b, []byte("card_state_codes")) && bytes.Contains(b, []byte("storage_info")) && !bytes.Contains(b, []byte("PRIVATE_"))
			}
		}
		if !found {
			t.Fatal("missing status evidence")
		}
	}
}
