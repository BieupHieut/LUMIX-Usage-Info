package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRejectTruncatedOrAmbiguousServiceReply(t *testing.T) {
	raw := make([]byte, 154)
	binary.LittleEndian.PutUint32(raw, SETUP_INFO_REPLY)
	binary.LittleEndian.PutUint32(raw[4:], 146)
	binary.LittleEndian.PutUint32(raw[140:], 12345)
	if r, err := parseUsageRaw(raw); err != nil || r.Shutter != 12345 {
		t.Fatalf("normal reply rejected: %v", err)
	}
	for _, bad := range [][]byte{raw[:153], append(append([]byte{}, raw...), 1), append(append([]byte{}, raw...), raw...)} {
		if _, err := parseUsageRaw(bad); err == nil {
			t.Fatal("malformed reply accepted")
		}
	}
}
func TestPNGReplacementFailurePreservesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "share.png")
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	if err := writePNGFile(path, img); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(path)
	if !bytes.HasPrefix(first, []byte{137, 80, 78, 71}) {
		t.Fatal("PNG invalid")
	}
	if err := writePNGFile(path, img); err != nil {
		t.Fatal(err)
	}
	if err := writePNGFile(dir, img); err == nil {
		t.Fatal("directory target succeeded")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(first, after) {
		t.Fatal("unrelated existing PNG corrupted")
	}
	leftovers, _ := filepath.Glob(filepath.Join(dir, ".lumix-export-*"))
	if len(leftovers) != 0 {
		t.Fatal("temporary files leaked")
	}
}

func TestCameraDataRejectsInvalidMetadata(t *testing.T) {
	if _, err := decodeCameraData(bundledCameraData); err != nil {
		t.Fatal(err)
	}
	for _, data := range [][]byte{
		append(append([]byte{}, bundledCameraData...), []byte(` {}`)...),
		bytes.Replace(bundledCameraData, []byte(`"schema": 1`), []byte(`"schema": 2`), 1),
		bytes.Replace(bundledCameraData, []byte(`2026-10-02`), []byte(`2026-13-02`), 1),
		bytes.Replace(bundledCameraData, []byte(`"schema": 1`), []byte(`"schema": 1, "unknown": true`), 1),
		bytes.Replace(bundledCameraData, []byte(`DC-S5M2`), []byte(`../../unsafe`), 1),
		bytes.Repeat([]byte(" "), 65537),
	} {
		if _, err := decodeCameraData(data); err == nil {
			t.Fatal("accepted invalid camera data")
		}
	}
	r := mustBundledCameraData()
	r.Profiles = append(r.Profiles, r.Profiles[0])
	data, _ := json.Marshal(r)
	if _, err := decodeCameraData(data); err == nil {
		t.Fatal("accepted duplicate verified pair")
	}
}
func TestCameraAdditionDoesNotChangeAppVersion(t *testing.T) {
	prior := cameraData
	defer func() { cameraData = prior; cameraPage = 0 }()
	r := mustBundledCameraData()
	r.Revision = "2026-10-03"
	r.Profiles = append(r.Profiles, cameraProfile{Model: "DC-S9", Firmware: "2.0", By: "TEST FIXTURE"})
	data, _ := json.Marshal(r)
	loaded, err := decodeCameraData(data)
	if err != nil {
		t.Fatal(err)
	}
	cameraData = loaded
	sample := UsageResult{Model: "DC-S9", Firmware: "Ver. 2.0"}
	applyValidationStatus(&sample)
	if !sample.SemanticsVerified || appVersion != "1.0.0" {
		t.Fatal("data-only addition failed")
	}
	sample.Firmware = "Ver. 2.1"
	applyValidationStatus(&sample)
	if sample.SemanticsVerified {
		t.Fatal("firmware validation leaked")
	}
	stateMu.Lock()
	currentView = "validation"
	stateMu.Unlock()
	var pager bool
	for _, a := range visibleActions() {
		if a.ID == "cameraNext" && a.Enabled {
			pager = true
		}
	}
	if !pager {
		t.Fatal("future camera rows inaccessible")
	}
}
func TestEstimatedMeaningsForRequestedModels(t *testing.T) {
	defer atomic.StoreUint32(&languageMode, languageAuto)
	for _, mode := range []uint32{languageKorean, languageEnglish} {
		atomic.StoreUint32(&languageMode, mode)
		for _, pair := range [][2]string{{"DC-S9", "2.0"}, {"DC-S5M2", "3.8"}, {"DC-G9M2", "2.5"}, {"DC-GH7", "1.0"}} {
			r := UsageResult{Model: pair[0], Firmware: pair[1], Serial: "PRIVATE0230", Shutter: 500, PowerWake: 100, Raw3: 3, Raw4: 4}
			applyValidationStatus(&r)
			if r.SemanticsVerified {
				t.Fatal("unverified firmware became verified")
			}
			report := buildReport(r, "UNVERIFIED MODEL / FIRMWARE", "Not verified", "Not available")
			for _, want := range []string{counterName(3, r), counterName(4, r), "2026-10-02"} {
				if !strings.Contains(report, want) {
					t.Fatalf("missing %q in report", want)
				}
			}
			if strings.Contains(report, "PRIVATE") || strings.Contains(report, "Abnormal Shutdown") {
				t.Fatal("privacy/false diagnostic regression")
			}
			summary := buildSummary(r, "last4")
			for _, want := range []string{counterName(1, r), counterName(2, r)} {
				if !strings.Contains(summary, want) {
					t.Fatal("estimated summary absent")
				}
			}
		}
	}
}
func TestReportsStayOutsideDocumentsAndNeverOverwrite(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	dir := preferredReportDir()
	if dir != filepath.Join(root, "Lumerian", "LUMIX Usage Info", "Reports") {
		t.Fatal(dir)
	}
	now := time.Date(2026, 10, 2, 22, 0, 0, 0, time.Local)
	p1, err := writeUniqueReport(dir, []byte("first"), now)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := writeUniqueReport(dir, []byte("second"), now)
	if err != nil {
		t.Fatal(err)
	}
	if p1 == p2 {
		t.Fatal("same-second overwrite")
	}
	one, _ := os.ReadFile(p1)
	if string(one) != "first" {
		t.Fatal("first report overwritten")
	}
	// A real path creation failure must not claim success or silently use Temp.
	bad := filepath.Join(root, "blocked-file")
	if err := os.WriteFile(bad, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	if path, err := writeUniqueReport(filepath.Join(bad, "reports"), []byte("test"), now); err == nil || path != "" {
		t.Fatal("failed save claimed success")
	}
}
