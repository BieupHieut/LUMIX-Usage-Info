//go:build windows

package main

import (
	"encoding/binary"
	"fmt"
	"time"
)

// Numeric fields only: never retain card labels, paths, filenames or unknown
// payloads in diagnostics. Storage order is NOT a physical slot mapping.
type boardCameraStatus struct {
	ReadAt      time.Time              `json:"read_at"`
	State       *uint16                `json:"camera_state_code,omitempty"`
	Images      *uint32                `json:"remaining_images,omitempty"`
	Seconds     *uint32                `json:"remaining_seconds,omitempty"`
	CardsRead   bool                   `json:"cards_read"`
	Cards       []uint16               `json:"card_state_codes,omitempty"`
	StorageRead bool                   `json:"storage_list_read"`
	Stores      []boardCameraStorage   `json:"storage,omitempty"`
	Detail      *boardCameraFreeDetail `json:"detailed_free_spaces,omitempty"`
}

type boardCameraStorage struct {
	InfoSkipped bool   `json:"info_skipped,omitempty"`
	ID          uint32 `json:"id"`
	Read        bool   `json:"read"`
	Type        uint16 `json:"storage_type"`
	Filesystem  uint16 `json:"filesystem_type"`
	Access      uint16 `json:"access_capability"`
	Capacity    uint64 `json:"capacity_bytes"`
	Free        uint64 `json:"free_bytes"`
	Images      uint32 `json:"remaining_images"`
}

type boardCameraFreeDetail struct {
	Count     uint16      `json:"entry_count"`
	Images    uint32      `json:"remaining_images"`
	Images4K  uint32      `json:"remaining_4k_images"`
	Seconds   uint32      `json:"remaining_seconds"`
	Seconds4K uint32      `json:"remaining_4k_seconds"`
	Arrays    [4][]uint32 `json:"per_entry_arrays_unverified"`
	Tail      [4]uint16   `json:"trailing_fields_unverified"`
}

func boardStorageIDs(b []byte) ([]uint32, error) {
	if len(b) < 4 {
		return nil, fmt.Errorf("unsupported_structure")
	}
	n := int(binary.LittleEndian.Uint32(b))
	if n > 16 || len(b) != 4+4*n {
		return nil, fmt.Errorf("unsupported_structure")
	}
	ids := make([]uint32, 0, n)
	seen := map[uint32]bool{}
	for i := 0; i < n; i++ {
		id := binary.LittleEndian.Uint32(b[4+i*4:])
		if id == 0 || id == 0xffffffff || seen[id] {
			return nil, fmt.Errorf("unsupported_structure")
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids, nil
}

func boardStorageInfo(id uint32, b []byte) (boardCameraStorage, error) {
	s := boardCameraStorage{ID: id}
	if len(b) < 28 {
		return s, fmt.Errorf("unsupported_structure")
	}
	// Validate both PTP strings, then discard them for privacy.
	off := 26
	for i := 0; i < 2; i++ {
		if _, ok := readPTPString(b, &off); !ok {
			return s, fmt.Errorf("unsupported_structure")
		}
	}
	if off != len(b) {
		return s, fmt.Errorf("unsupported_structure")
	}
	s.Type = binary.LittleEndian.Uint16(b)
	s.Filesystem = binary.LittleEndian.Uint16(b[2:])
	s.Access = binary.LittleEndian.Uint16(b[4:])
	s.Capacity = binary.LittleEndian.Uint64(b[6:])
	s.Free = binary.LittleEndian.Uint64(b[14:])
	s.Images = binary.LittleEndian.Uint32(b[22:])
	s.Read = true
	return s, nil
}

func (s boardCameraStorage) capacityKnown() bool {
	return s.Read && s.Capacity != 0 && s.Capacity != ^uint64(0)
}
func (s boardCameraStorage) spaceKnown() bool {
	return s.capacityKnown() && s.Free != ^uint64(0) && s.Free <= s.Capacity
}
func boardKnown32(v uint32) *uint32 {
	if v == ^uint32(0) {
		return nil
	}
	return &v
}

// Tether 2.9 GetRecInfo: 12000011=u32 images, 12=u32 seconds,
// 13=u16 state. These are RESPONSE tags, not alternative query selectors.
func boardParseCameraStatus(raw []byte) (*uint16, *uint32, *uint32, error) {
	frames, e := checkedServiceFrames(raw)
	if e != nil {
		return nil, nil, nil, e
	}
	var state *uint16
	var images, seconds *uint32
	seen := map[uint32]bool{}
	found := false
	for _, f := range frames {
		if f.Tag < 0x12000011 || f.Tag > 0x12000013 {
			continue
		}
		if seen[f.Tag] {
			return nil, nil, nil, fmt.Errorf("unsupported_structure")
		}
		seen[f.Tag] = true
		n := 4
		if f.Tag == 0x12000013 {
			n = 2
		}
		if len(f.Data) != n {
			return nil, nil, nil, fmt.Errorf("unsupported_structure")
		}
		found = true
		switch f.Tag {
		case 0x12000011:
			images = boardKnown32(binary.LittleEndian.Uint32(f.Data))
		case 0x12000012:
			seconds = boardKnown32(binary.LittleEndian.Uint32(f.Data))
		case 0x12000013:
			v := binary.LittleEndian.Uint16(f.Data)
			if v != 0xffff {
				state = &v
			}
		}
	}
	if !found {
		return nil, nil, nil, fmt.Errorf("unsupported_structure")
	}
	return state, images, seconds, nil
}

func boardSingleFrame(raw []byte, tag uint32) ([]byte, error) {
	fs, e := checkedServiceFrames(raw)
	if e != nil {
		return nil, e
	}
	var data []byte
	found := false
	for _, f := range fs {
		if f.Tag == tag {
			if found {
				return nil, fmt.Errorf("unsupported_structure")
			}
			found = true
			data = f.Data
		}
	}
	if !found {
		return nil, fmt.Errorf("unsupported_structure")
	}
	return data, nil
}

func boardParseCards(raw []byte) ([]uint16, error) {
	b, e := boardSingleFrame(raw, 0x12000015)
	if e != nil {
		return nil, e
	}
	if len(b) < 2 {
		return nil, fmt.Errorf("unsupported_structure")
	}
	n := int(binary.LittleEndian.Uint16(b))
	if n > 16 || len(b) != 2+2*n {
		return nil, fmt.Errorf("unsupported_structure")
	}
	values := make([]uint16, n)
	for i := range values {
		values[i] = binary.LittleEndian.Uint16(b[2+i*2:])
	}
	return values, nil
}

func boardParseFreeDetail(raw []byte) (*boardCameraFreeDetail, error) {
	b, e := boardSingleFrame(raw, 0x12000016)
	if e != nil {
		return nil, e
	}
	if len(b) < 26 {
		return nil, fmt.Errorf("unsupported_structure")
	}
	n := int(binary.LittleEndian.Uint16(b))
	if n > 16 || len(b) != 26+16*n {
		return nil, fmt.Errorf("unsupported_structure")
	}
	d := &boardCameraFreeDetail{Count: uint16(n), Images: binary.LittleEndian.Uint32(b[2:]), Images4K: binary.LittleEndian.Uint32(b[6:])}
	off := 10
	for group := 0; group < 4; group++ {
		if group == 2 {
			d.Seconds = binary.LittleEndian.Uint32(b[off:])
			d.Seconds4K = binary.LittleEndian.Uint32(b[off+4:])
			off += 8
		}
		d.Arrays[group] = make([]uint32, n)
		for i := 0; i < n; i++ {
			d.Arrays[group][i] = binary.LittleEndian.Uint32(b[off:])
			off += 4
		}
	}
	for i := range d.Tail {
		d.Tail[i] = binary.LittleEndian.Uint16(b[off:])
		off += 2
	}
	return d, nil
}

// New optional queries are deliberately bounded and never retry. Publish each
// result so an eventual stalled driver does not erase prior successful fields.
func boardReadCameraStatus(dev boardTransport, out *boardReply, send func(string)) {
	c := &boardCameraStatus{ReadAt: time.Now()}
	out.Camera = c
	out.Diagnostic.Camera = c
	query := func(stage string, op uint32, args []uint32, parse func([]byte) error) {
		out.Diagnostic.ActiveStage = stage
		send("reading_extra")
		start := time.Now()
		raw, rc, e := dev.SendReadPTP(op, args)
		if e == nil && rc == PTP_RC_OK {
			e = parse(raw)
		}
		status := "ok"
		if e != nil || rc != PTP_RC_OK {
			status = "failed"
		}
		out.Diagnostic.step(stage, status, rc, e, start)
		s := &out.Diagnostic.Steps[len(out.Diagnostic.Steps)-1]
		s.Opcode = op
		s.Parameters = append([]uint32(nil), args...)
		if t, ok := dev.(*WPDDevice); ok {
			tr := t.trace
			s.Transfer = &tr
		}
		send("reading_extra")
	}
	query("card_status", 0x9414, []uint32{0x12000015}, func(raw []byte) error {
		v, e := boardParseCards(raw)
		if e == nil {
			c.Cards = v
			c.CardsRead = true
		}
		return e
	})
	query("storage_ids", 0x1004, nil, func(raw []byte) error {
		ids, e := boardStorageIDs(raw)
		if e != nil {
			return e
		}
		c.StorageRead = true
		for _, id := range ids {
			c.Stores = append(c.Stores, boardCameraStorage{ID: id})
		}
		return nil
	})
	for i := range c.Stores {
		// A zero logical-store component is not addressable by GetStorageInfo.
		// Preserve the returned ID for diagnostics; do not fabricate a replacement.
		if c.Stores[i].ID&0xffff == 0 {
			c.Stores[i].InfoSkipped = true
			out.Diagnostic.step("storage_info", "skipped", 0, fmt.Errorf("storage_not_addressable"), time.Now())
			s := &out.Diagnostic.Steps[len(out.Diagnostic.Steps)-1]
			s.Opcode = 0x1005
			s.Parameters = []uint32{c.Stores[i].ID}
			send("reading_extra")
			continue
		}
		query("storage_info", 0x1005, []uint32{c.Stores[i].ID}, func(raw []byte) error {
			s, e := boardStorageInfo(c.Stores[i].ID, raw)
			if e == nil {
				c.Stores[i] = s
			}
			return e
		})
	}
	out.Diagnostic.ActiveStage = ""
	send("reading_extra")
}

func boardSanitizeCamera(c *boardCameraStatus) *boardCameraStatus {
	if c == nil {
		return nil
	}
	v := *c
	if len(v.Cards) > 16 {
		v.Cards = nil
		v.CardsRead = false
	}
	if len(v.Stores) > 16 {
		v.Stores = nil
		v.StorageRead = false
	}
	if v.Detail != nil {
		d := *v.Detail
		valid := d.Count <= 16
		for _, a := range d.Arrays {
			valid = valid && len(a) == int(d.Count)
		}
		if !valid {
			v.Detail = nil
		} else {
			v.Detail = &d
		}
	}
	return &v
}

func boardCameraStateKey(v uint16) string {
	keys := []string{"state_standby", "state_movie", "state_movie", "state_playback", "state_processing", "state_nr", "state_menu", "state_stream", "state_remote", "state_movie", "state_stream_rec", "state_remote"}
	if int(v) < len(keys) {
		return keys[v]
	}
	return "state_unknown"
}
func boardCardStateKey(v uint16) string {
	switch v {
	case 0:
		return "card_absent"
	case 1:
		return "card_ready"
	case 2:
		return "card_protected"
	case 3:
		return "card_error"
	}
	return "state_unknown"
}
