//go:build windows

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
)

func l10TestPTP(model, serial string, ops []uint16) []byte {
	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, uint16(100))
	binary.Write(&b, binary.LittleEndian, uint32(1))
	binary.Write(&b, binary.LittleEndian, uint16(100))
	b.WriteByte(0)
	binary.Write(&b, binary.LittleEndian, uint16(0))
	binary.Write(&b, binary.LittleEndian, uint32(len(ops)))
	for _, op := range ops {
		binary.Write(&b, binary.LittleEndian, op)
	}
	for i := 0; i < 4; i++ {
		binary.Write(&b, binary.LittleEndian, uint32(0))
	}
	for _, s := range []string{"Panasonic", model, "1.2", serial} {
		u := syscall.StringToUTF16(s)
		b.WriteByte(byte(len(u)))
		for _, v := range u {
			binary.Write(&b, binary.LittleEndian, v)
		}
	}
	return b.Bytes()
}
func l10TestReply(length int, tag uint32) []byte {
	data := make([]byte, 8+length)
	binary.LittleEndian.PutUint32(data, tag)
	binary.LittleEndian.PutUint32(data[4:], uint32(length))
	if length >= 146 {
		binary.LittleEndian.PutUint32(data[8+128:], 42)
		binary.LittleEndian.PutUint32(data[8+132:], 100)
	}
	return data
}

type l10FakeTransport struct {
	info, raw []byte
	ops       []uint32
	args      [][]uint32
	rc        uint32
	err       error
}

func (d *l10FakeTransport) SendReadPTP(op uint32, args []uint32) ([]byte, uint32, error) {
	d.ops = append(d.ops, op)
	d.args = append(d.args, append([]uint32(nil), args...))
	if d.err != nil {
		return nil, 0, d.err
	}
	if op == PTP_GET_DEVICE_INFO {
		return d.info, d.rc, nil
	}
	return d.raw, d.rc, nil
}
func l10TestWorkerReply() l10WorkerReply {
	return l10WorkerReply{Status: "ERROR", PTP: l10PTPResult{Status: "NOT_RUN"}, Read: l10ReadResult{Status: "NOT_RUN"}}
}

func TestL10TransportConsentModelAndOperationGates(t *testing.T) {
	good := l10TestPTP("DC-L10", "PRIVATE_SERIAL", []uint16{0x1001, 0x9414})
	_, identity, err := l10ParsePTP(good)
	if err != nil {
		t.Fatal(err)
	}
	candidate := l10Fixture("A").Candidates[0]
	cases := []struct {
		name, stage, model string
		consent            bool
		identity           string
		ops                []uint16
		want               []uint32
		pass               bool
	}{
		{"no consent", "B", "DC-L10", false, "", []uint16{0x9414}, nil, false},
		{"B standard only", "B", "DC-L10", true, "", []uint16{0x1001, 0x9414}, []uint32{0x1001}, true},
		{"C identity missing", "C", "DC-L10", true, "", []uint16{0x9414}, nil, false},
		{"other model", "C", "DC-S9", true, identity, []uint16{0x1001, 0x9414}, []uint32{0x1001}, false},
		{"old DMC-L10", "B", "DMC-L10", true, "", []uint16{0x1001, 0x9414}, []uint32{0x1001}, false},
		{"no advertised read", "C", "DC-L10", true, identity, []uint16{0x1001}, []uint32{0x1001}, false},
		{"different camera identity", "C", "DC-L10", true, strings.Repeat("0", 64), []uint16{0x1001, 0x9414}, []uint32{0x1001}, false},
		{"C permitted once", "C", "DC-L10", true, identity, []uint16{0x1001, 0x9414}, []uint32{0x1001, 0x9414}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dev := &l10FakeTransport{info: l10TestPTP(tc.model, "PRIVATE_SERIAL", tc.ops), raw: l10TestReply(146, SETUP_INFO_REPLY), rc: PTP_RC_OK}
			r := l10ProbeTransport(l10WorkerRequest{Stage: tc.stage, Target: candidate, Consent: tc.consent, Identity: tc.identity}, dev, l10TestWorkerReply())
			if !reflect.DeepEqual(dev.ops, tc.want) {
				t.Fatalf("commands %v, want %v", dev.ops, tc.want)
			}
			if (r.Status == "PASS") != tc.pass {
				t.Fatalf("result %+v", r)
			}
			if len(dev.ops) == 2 && (!reflect.DeepEqual(dev.args[1], []uint32{SETUP_INFO_TAG}) || r.Read.Validated) {
				t.Fatal("bad Panasonic parameter or validation leak")
			}
		})
	}
}
func TestL10StrictReplyValidation(t *testing.T) {
	cases := []struct {
		name string
		raw  []byte
		rc   uint32
		pass bool
	}{{"valid", l10TestReply(146, SETUP_INFO_REPLY), 0x2001, true}, {"too short", l10TestReply(145, SETUP_INFO_REPLY), 0x2001, false}, {"too long", l10TestReply(147, SETUP_INFO_REPLY), 0x2001, false}, {"wrong tag", l10TestReply(146, 0x12345678), 0x2001, false}, {"bad status", l10TestReply(146, SETUP_INFO_REPLY), 0x2005, false}, {"truncated tail", append(l10TestReply(146, SETUP_INFO_REPLY), 1), 0x2001, false}, {"duplicate tag", append(l10TestReply(146, SETUP_INFO_REPLY), l10TestReply(146, SETUP_INFO_REPLY)...), 0x2001, false}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := l10ParseCounters(tc.raw, tc.rc)
			if (r.Status == "PASS") != tc.pass || r.Validated {
				t.Fatalf("wrong strict result: %+v", r)
			}
			if !tc.pass && len(r.Counters) != 0 {
				t.Fatal("invalid response leaked counters")
			}
			if tc.pass && r.Counters[1] != 100 {
				t.Fatal("wrong raw counter")
			}
		})
	}
	// Exercise parser bounds without native USB calls.
	for n := 0; n < 160; n++ {
		l10ParseCounters(make([]byte, n), 0x2001)
		l10ParsePTP(make([]byte, n))
	}
}
func TestL10MaskedReportsAndAtomicPair(t *testing.T) {
	s := l10NewSession()
	s.Firmware = "1.2"
	s.Fixture = true
	s.USBStatus = "PASS"
	s.Candidates = l10Fixture("A").Candidates
	s.Selected = 0
	s.Candidates[0].USBID = `USB\PRIVATE_SERIAL_AND_PATH`
	s.Candidates[0].WPDIDs = []string{`C:\Users\PRIVATE_USERNAME\pairing_key`}
	s.Identity = "PRIVATE_IDENTITY"
	s.Candidates[0].USB.Name = "my DC-L10 PRIVATE_SERIAL"
	s.Candidates[0].USB.Driver = "PRIVATE_USERNAME"
	s.Candidates[0].USB.Interfaces[0].Endpoints = "PRIVATE_PAIRING_KEY"
	s.PTP = l10Fixture("B").PTP
	s.Read = l10Fixture("C").Read
	s.Read.Validated = true
	report := l10BuildReport(s)
	j, err := l10ReportJSON(report)
	if err != nil {
		t.Fatal(err)
	}
	txt := l10ReportText(report)
	for _, needle := range []string{"PRIVATE_SERIAL", "PRIVATE_USERNAME", "pairing_key", "PRIVATE_IDENTITY", "PRIVATE_PAIRING_KEY"} {
		if strings.Contains(string(j)+txt, needle) {
			t.Fatalf("private field exported: %s", needle)
		}
	}
	if report.Read.Validated || !report.Fixture || report.ModeVerified || !strings.Contains(txt, "Raw Counter 2: 100 — UNVERIFIED") {
		t.Fatal("fixture or counter meanings were misrepresented")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "report.json")
	if err := l10SavePair(path, report); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.ReadFile(filepath.Join(dir, "report.txt")); err != nil {
		t.Fatal(err)
	}
	if err = l10SavePair(path, report); err != nil {
		t.Fatal("overwrite", err)
	}
	// Failure on the second destination must retain the first destination intact.
	os.Remove(filepath.Join(dir, "report.txt"))
	os.Mkdir(filepath.Join(dir, "report.txt"), 0700)
	report.Firmware = "1.3"
	if err = l10SavePair(path, report); err == nil {
		t.Fatal("directory destination accepted")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("JSON was not restored after failed pair save")
	}
}
func TestL10OfflineGateAndExistingCameraRegression(t *testing.T) {
	s := l10NewSession()
	s.Firmware = "1.2"
	if l10CanB(s) || l10CanC(s) {
		t.Fatal("new session sends commands")
	}
	s.Candidates = l10Fixture("A").Candidates
	s.Selected = 0
	s.USBStatus = "PASS"
	if !l10CanB(s) || l10CanC(s) {
		t.Fatal("A incorrectly enables C")
	}
	s.PTP = l10Fixture("B").PTP
	s.BConsent = true
	s.Fixture = true
	s.Identity = "FIXTURE_ONLY"
	if !l10CanC(s) {
		t.Fatal("B did not unlock conditional C")
	}
	s.Read.Status = "ERROR"
	if l10CanC(s) {
		t.Fatal("automatic repeat of C permitted")
	}
	s.Read.Status = "NOT_RUN"
	s.Busy = true
	if l10CanC(s) || l10CanB(s) {
		t.Fatal("busy session accepts another request")
	}
	for _, model := range []string{"DC-S1RM2", "DC-S5M2", "DC-S5", "DC-S9"} {
		if l10Recognized(DeviceInfo{FriendlyName: model}) {
			t.Fatal("other camera blocked from main path")
		}
	}
	if l10AllowedCandidate(l10Candidate{ModelHint: "DC-S9", USB: l10USBType{VID: "04DA"}}) {
		t.Fatal("S9 selectable in L10 diagnostic")
	}
	for _, tc := range []struct {
		err    error
		reason string
	}{{errors.New("80070005"), "ACCESS_DENIED"}, {errors.New("800700AA"), "DEVICE_BUSY"}, {errors.New("8007048F"), "DEVICE_DISCONNECTED"}} {
		_, reason := l10Failure(tc.err)
		if reason != tc.reason {
			t.Fatal(reason)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := l10Worker(ctx, l10WorkerRequest{Stage: "A"})
	if r.Status != "CANCELLED" {
		t.Fatal("cancelled worker was not stopped")
	}
}

func TestL10PTPFailuresDoNotReadCounters(t *testing.T) {
	for _, tc := range []struct {
		rc             uint32
		status, reason string
	}{
		{0x2005, "UNSUPPORTED", "GET_DEVICE_INFO_UNSUPPORTED"},
		{0x200F, "INACCESSIBLE", "ACCESS_DENIED"},
		{0x2019, "INACCESSIBLE", "DEVICE_BUSY"},
		{0x201F, "CANCELLED", "TRANSACTION_CANCELLED"},
	} {
		dev := &l10FakeTransport{rc: tc.rc}
		r := l10ProbeTransport(l10WorkerRequest{Stage: "B", Target: l10Fixture("A").Candidates[0], Consent: true}, dev, l10TestWorkerReply())
		if r.Status != tc.status || r.Reason != tc.reason || len(dev.ops) != 1 || dev.ops[0] != 0x1001 || r.Read.Attempted {
			t.Fatalf("bad failure gate: %+v", r)
		}
		read := l10ParseCounters(l10TestReply(146, SETUP_INFO_REPLY), tc.rc)
		if read.Status != tc.status || len(read.Counters) != 0 {
			t.Fatalf("failure released counters: %+v", read)
		}
	}
}

func TestL10AbsentIdentityCannotUnlockC(t *testing.T) {
	dev := &l10FakeTransport{info: l10TestPTP("DC-L10", "", []uint16{0x1001, 0x9414}), rc: PTP_RC_OK}
	r := l10ProbeTransport(l10WorkerRequest{Stage: "B", Target: l10Fixture("A").Candidates[0], Consent: true}, dev, l10TestWorkerReply())
	if r.Status != "PASS" || r.PTP.Model != "DC-L10" || r.Identity != "" || r.PTP.Reason != "DEVICE_IDENTITY_UNAVAILABLE" {
		t.Fatalf("bad absent identity handling: %+v", r)
	}
	s := l10NewSession()
	s.Candidates = l10Fixture("A").Candidates
	s.Selected = 0
	s.BConsent = true
	s.PTP = r.PTP
	s.Identity = r.Identity
	if l10CanC(s) {
		t.Fatal("absent serial identity unlocked C")
	}
}

func TestL10ResultNoticesFollowLanguageChanges(t *testing.T) {
	old := atomic.LoadUint32(&languageMode)
	defer atomic.StoreUint32(&languageMode, old)
	atomic.StoreUint32(&languageMode, languageEnglish)
	notice := l10ResultNotice("A", l10WorkerReply{Status: "DEVICE_NOT_FOUND"})
	atomic.StoreUint32(&languageMode, languageKorean)
	ko := l10NoticeText(notice)
	if !strings.Contains(ko, "장치 후보가 없습니다") {
		t.Fatal("old English notice survived Korean switch")
	}
	failed := l10ResultNotice("B", l10WorkerReply{Status: "INACCESSIBLE", Reason: "DEVICE_BUSY"})
	atomic.StoreUint32(&languageMode, languageEnglish)
	if l10NoticeText(ko) != notice || l10NoticeText(failed) != "Test stopped: INACCESSIBLE · DEVICE_BUSY" {
		t.Fatal("Korean notice survived English switch")
	}
	if l10NoticeText("JSON + TXT 저장됨: example.json") != "JSON + TXT saved: example.json" {
		t.Fatal("saved notice lost filename or language")
	}
}
