//go:build windows

package main

import (
	"archive/zip"
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type boardRecord struct {
	BreakBefore  bool      `json:"break_before,omitempty"`
	ID           string    `json:"id"`
	Body         string    `json:"body"`
	Model        string    `json:"model"`
	Firmware     string    `json:"firmware"`
	SerialSuffix string    `json:"serial_suffix,omitempty"`
	Time         time.Time `json:"time"`
	Shutter      uint32    `json:"shutter"`
	Power        uint32    `json:"power_wake"`
	Verified     bool      `json:"verified"`
}
type boardStore struct {
	Schema       int           `json:"schema"`
	Key          string        `json:"local_key"`
	Auto         bool          `json:"auto_history"`
	External     string        `json:"external_file,omitempty"`
	ExternalHash string        `json:"external_hash,omitempty"`
	ExternalAuto bool          `json:"external_auto"`
	Records      []boardRecord `json:"records"`
}

func boardStorePath() string { return filepath.Join(preferredReportDir(), "..", "usage-board.json") }
func boardAtomicWrite(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".lumerian-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
func boardLoad(path string) (boardStore, error) {
	var s boardStore
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		key := make([]byte, 32)
		if _, err = rand.Read(key); err != nil {
			return s, err
		}
		s = boardStore{Schema: 1, Key: hex.EncodeToString(key), Auto: true, Records: []boardRecord{}}
		return s, boardPersist(path, s)
	}
	if err != nil {
		return s, err
	}
	if err = json.Unmarshal(b, &s); err != nil {
		return s, err
	}
	key, e := hex.DecodeString(s.Key)
	if (s.Schema != 1 && s.Schema != 2) || e != nil || len(key) != 32 {
		return s, errors.New("invalid history store")
	}
	seen := map[string]bool{}
	for _, r := range s.Records {
		if r.ID == "" || r.Body == "" || r.Time.IsZero() || seen[r.ID] {
			return s, errors.New("invalid history record")
		}
		seen[r.ID] = true
	}
	return s, nil
}
func boardPersist(path string, s boardStore) error {
	b, e := json.MarshalIndent(s, "", "  ")
	if e != nil {
		return e
	}
	return boardAtomicWrite(path, b)
}
func boardAppend(s *boardStore, r UsageResult) error {
	if !boardReliableSerial(r.Serial) {
		return errors.New("serial_unavailable")
	}
	key, e := hex.DecodeString(s.Key)
	if e != nil || len(key) != 32 {
		return errors.New("invalid_key")
	}
	h := hmac.New(sha256.New, key)
	h.Write([]byte(modelKey(r.Model) + "\x00" + strings.TrimSpace(r.Serial)))
	nonce := make([]byte, 8)
	if _, e = rand.Read(nonce); e != nil {
		return e
	}
	suffix := r.Serial
	if len(suffix) > 4 {
		suffix = suffix[len(suffix)-4:]
	}
	s.Records = append(s.Records, boardRecord{ID: r.Timestamp.Format("20060102_150405") + "_" + hex.EncodeToString(nonce), Body: hex.EncodeToString(h.Sum(nil)), Model: r.Model, Firmware: r.Firmware, SerialSuffix: suffix, Time: r.Timestamp, Shutter: r.Shutter, Power: r.PowerWake, Verified: r.SemanticsVerified})
	return nil
}
func boardReliableSerial(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 4 || strings.ContainsAny(s, "*\r\n\t\x00") {
		return false
	}
	switch strings.ToLower(s) {
	case "unknown", "not available", "n/a", "none", "0000", "00000000":
		return false
	}
	return true
}
func boardDifference(a, b boardRecord) (uint32, uint32, bool) {
	if b.BreakBefore || a.Verified != b.Verified || modelKey(a.Model) != modelKey(b.Model) || a.Body != b.Body || firmwareKey(a.Firmware) != firmwareKey(b.Firmware) || b.Time.Before(a.Time) || b.Shutter < a.Shutter || b.Power < a.Power {
		return 0, 0, false
	}
	return b.Shutter - a.Shutter, b.Power - a.Power, true
}
func boardExportBytes(s boardStore) ([]byte, error) {
	records := append([]boardRecord(nil), s.Records...)
	aliases := map[string]string{}
	for i := range records {
		body := records[i].Body
		if aliases[body] == "" {
			aliases[body] = fmt.Sprintf("camera-%03d", len(aliases)+1)
		}
		records[i].Body = aliases[body]
	}
	return json.MarshalIndent(struct {
		Schema  int           `json:"schema"`
		Records []boardRecord `json:"records"`
	}{s.Schema, records}, "", "  ")
}
func boardFileHash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func boardExternalSave(s *boardStore, path string, replace bool) error {
	existing, e := os.ReadFile(path)
	if e != nil && !os.IsNotExist(e) {
		return e
	}
	if !replace && (e != nil || s.ExternalHash == "" || boardFileHash(existing) != s.ExternalHash) {
		return errors.New("external_file_changed")
	}
	b, e := boardExportBytes(*s)
	if e != nil {
		return e
	}
	if e = boardAtomicWrite(path, b); e != nil {
		return e
	}
	s.External = path
	s.ExternalHash = boardFileHash(b)
	return nil
}
func boardDiagnosticZIP(d boardDiagnostic) ([]byte, error) {
	if b, e := hex.DecodeString(d.RunID); e != nil || len(b) != 16 {
		d.RunID = ""
	}
	// Diagnostics are constructed from a whitelist; camera payloads and identifiers never enter it.
	if !boardModelPattern.MatchString(d.Model) {
		d.Model = ""
	}
	if !l10VersionPattern.MatchString(d.Firmware) {
		d.Firmware = ""
	}
	d.Camera = boardSanitizeCamera(d.Camera)
	safe := map[string]bool{"identity": true, "parse_extra": true, "settle": true, "initialize": true, "enumerate": true, "open_device": true, "device_info": true, "session_policy": true, "open_session": true, "usage": true, "parse_usage": true, "close_session": true, "release_device": true, "worker": true}
	for _, stage := range []string{"card_status", "storage_ids", "storage_info"} {
		safe[stage] = true
	}
	steps := []boardStep{}
	for _, s := range d.Steps {
		if !safe[s.Stage] {
			continue
		}
		if s.Status != "ok" && s.Status != "failed" && s.Status != "skipped" && s.Status != "unconfirmed" {
			continue
		}
		if !boardResponsePattern.MatchString(s.PTP) {
			s.PTP = ""
		}
		if !boardResponsePattern.MatchString(s.Code) && !boardSafeCode[s.Code] {
			s.Code = ""
		}
		if !boardWPDOrigins[s.ErrorSource] {
			s.ErrorSource = ""
		}
		steps = append(steps, boardSanitizeStep(s))
	}
	d.Steps = steps
	if !safe[d.FailedStage] {
		d.FailedStage = ""
	}
	if !safe[d.ActiveStage] {
		d.ActiveStage = ""
	}
	if len(d.Operations) > 1024 {
		d.Operations = nil
	}
	if d.Outcome != "completed" && d.Outcome != "stopped" && d.Outcome != "interrupted" {
		d.Outcome = ""
	}
	d.Version = appVersion
	if d.DeviceAccess != "GENERIC_READ" && d.DeviceAccess != "GENERIC_READ|GENERIC_WRITE (session transport only)" && d.DeviceAccess != "GENERIC_READ then GENERIC_READ|GENERIC_WRITE (session transport only)" {
		d.DeviceAccess = ""
	}
	if d.Policy != "legacy" && d.Policy != "diagnostic-session" && d.Policy != "s1-comprehensive-v1" {
		d.Policy = "unknown"
	}
	b, e := json.MarshalIndent(d, "", "  ")
	if e != nil {
		return nil, e
	}
	var out bytes.Buffer
	z := zip.NewWriter(&out)
	w, e := z.Create("diagnostic.json")
	if e != nil {
		return nil, e
	}
	if _, e = w.Write(b); e != nil {
		return nil, e
	}
	w, e = z.Create("README.txt")
	if e != nil {
		return nil, e
	}
	_, e = fmt.Fprintf(w, "%s\r\n%s\r\n\r\nhttps://github.com/BieupHieut/LUMIX-Usage-Info/issues\r\nInstagram DM: @bieup_hieut\r\n\r\n%s / %s\r\n", boardText("diagnostic"), boardText("privacy"), d.Model, d.Firmware)
	if e != nil {
		return nil, e
	}
	if d.Policy == "s1-comprehensive-v1" {
		w, e = z.Create("SUMMARY.txt")
		if e != nil {
			return nil, e
		}
		if _, e = w.Write([]byte(boardSuiteSummary(d))); e != nil {
			return nil, e
		}
	}
	if e = z.Close(); e != nil {
		return nil, e
	}
	return out.Bytes(), nil
}
