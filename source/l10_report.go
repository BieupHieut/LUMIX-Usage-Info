//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type l10Report struct {
	Schema       string        `json:"schema_version"`
	TestID       string        `json:"test_id"`
	App          string        `json:"app_version"`
	Generated    string        `json:"generated_at"`
	Fixture      bool          `json:"synthetic_fixture_not_hardware"`
	Model        string        `json:"camera_model_reported_by_user"`
	Firmware     string        `json:"firmware_reported_by_user"`
	Mode         string        `json:"usb_mode_reported_by_user"`
	ModeVerified bool          `json:"usb_mode_verified_by_os"`
	OS           string        `json:"os"`
	Baseline     string        `json:"baseline_status"`
	USBStatus    string        `json:"usb_status"`
	USBReason    string        `json:"usb_reason"`
	Candidates   []l10USBType  `json:"usb_candidates"`
	Selected     *l10USBType   `json:"usb_selected"`
	PTP          l10PTPResult  `json:"ptp"`
	Read         l10ReadResult `json:"panasonic_read"`
	BConsent     bool          `json:"standard_ptp_user_consent"`
	CConsent     bool          `json:"panasonic_read_user_consent"`
	Removed      bool          `json:"personal_identifiers_removed"`
}

func l10Hex(text string, length int) bool {
	if len(text) != length {
		return false
	}
	for _, ch := range strings.ToUpper(text) {
		if !(ch >= '0' && ch <= '9' || ch >= 'A' && ch <= 'F') {
			return false
		}
	}
	return true
}

func l10PublicUSB(in l10USBType) l10USBType {
	out := in
	out.Name = l10ModelPattern.FindString(strings.ToUpper(in.Name))
	if out.Name == "" {
		out.Name = "Panasonic USB device"
	}
	for _, field := range []*string{&out.VID, &out.PID} {
		if !l10Hex(*field, 4) {
			*field = "not available"
		}
	}
	known := map[string]bool{"usbccgp": true, "wudfrd": true, "winusb": true, "usbstor": true, "usbvideo": true, "usb": true, "umpass": true, "fixture": true}
	if !known[strings.ToLower(out.Driver)] {
		out.Driver = "not available"
	}
	out.Interfaces = append([]l10Interface(nil), in.Interfaces...)
	for i := range out.Interfaces {
		v := &out.Interfaces[i]
		for _, field := range []*string{&v.Number, &v.Class, &v.Subclass, &v.Protocol} {
			if !l10Hex(*field, 2) {
				*field = "not available"
			}
		}
		v.Endpoints = "not available"
	}
	return out
}

func l10BuildReport(s l10Session) l10Report {
	fw := s.Firmware
	if !l10VersionPattern.MatchString(fw) {
		fw = "not reported"
	}
	out := l10Report{Schema: "1.1", TestID: "L10_LAB_USB", App: appVersion, Generated: time.Now().Format(time.RFC3339), Fixture: s.Fixture, Model: "DC-L10", Firmware: fw, Mode: "LUMIX Lab", OS: "Windows", Baseline: s.BaselineStatus, USBStatus: s.USBStatus, USBReason: s.USBReason, Candidates: []l10USBType{}, PTP: s.PTP, Read: s.Read, BConsent: s.BConsent, CConsent: s.CConsent, Removed: true}
	for _, c := range s.Candidates {
		out.Candidates = append(out.Candidates, l10PublicUSB(c.USB))
	}
	if s.Selected >= 0 && s.Selected < len(s.Candidates) {
		copy := l10PublicUSB(s.Candidates[s.Selected].USB)
		out.Selected = &copy
	}
	// Validation is never inheritable from main-camera or fixture status.
	out.Read.Validated = false
	if out.PTP.Operations == nil {
		out.PTP.Operations = []uint16{}
	}
	return out
}
func l10ReportText(r l10Report) string {
	var b strings.Builder
	b.WriteString("L10 USB Connection Test — UNVERIFIED\n")
	if r.Fixture {
		b.WriteString("SYNTHETIC FIXTURE — NOT A HARDWARE OBSERVATION\n")
	}
	fmt.Fprintf(&b, "Schema: %s / Test: %s\nTool: %s\nGenerated: %s\nSynthetic fixture: %t\nModel (user): %s\nFirmware (user): %s\nUSB mode (user): %s — NOT confirmed by Windows\nOS: %s\nPersonal identifiers removed: %t\n\n", r.Schema, r.TestID, r.App, r.Generated, r.Fixture, r.Model, r.Firmware, r.Mode, r.OS, r.Removed)
	fmt.Fprintf(&b, "A USB: %s / %s\nBaseline: %s\nCandidates: %d\n", r.USBStatus, r.USBReason, r.Baseline, len(r.Candidates))
	for i, c := range r.Candidates {
		fmt.Fprintf(&b, "  %d. %s | VID %s PID %s | Driver %s | WPD %t\n", i+1, c.Name, c.VID, c.PID, c.Driver, c.WPDExposed)
		fmt.Fprintf(&b, "WPD enumeration: %s\n", c.WPDStatus)
		if c.NewSinceBaseline != nil {
			fmt.Fprintf(&b, "New since baseline: %t\n", *c.NewSinceBaseline)
		} else {
			b.WriteString("New since baseline: not compared\n")
		}
		for _, v := range c.Interfaces {
			fmt.Fprintf(&b, "Interface %s | Class %s Subclass %s Protocol %s\nEndpoints: %s\n", v.Number, v.Class, v.Subclass, v.Protocol, v.Endpoints)
		}
	}
	if r.Selected != nil {
		fmt.Fprintf(&b, "Selected: %s / VID %s PID %s\n", r.Selected.Name, r.Selected.VID, r.Selected.PID)
	}
	fmt.Fprintf(&b, "\nB Standard PTP: %s / %s\nAttempted: %t / Execution uncertain: %t / User consent: %t\nConfirmed model: %s / Observed firmware: %s\nOperations:", r.PTP.Status, r.PTP.Reason, r.PTP.Attempted, r.PTP.Uncertain, r.BConsent, r.PTP.Model, r.PTP.Firmware)
	for _, op := range r.PTP.Operations {
		fmt.Fprintf(&b, " %04X", op)
	}
	b.WriteString("\n")
	if r.PTP.Response != nil {
		fmt.Fprintf(&b, "PTP response: %04X\n", *r.PTP.Response)
	}
	fmt.Fprintf(&b, "PTP standard version: %d / Vendor extension ID: %08X\n", r.PTP.StandardVersion, r.PTP.VendorExtension)
	fmt.Fprintf(&b, "\nC Existing read: %s / %s\nAttempted: %t / Execution uncertain: %t / User consent: %t\n", r.Read.Status, r.Read.Reason, r.Read.Attempted, r.Read.Uncertain, r.CConsent)
	if r.Read.Response != nil {
		fmt.Fprintf(&b, "PTP response: %04X\n", *r.Read.Response)
	}
	if r.Read.ExpectedTag != nil {
		fmt.Fprintf(&b, "Expected tag found: %t\n", *r.Read.ExpectedTag)
	}
	if r.Read.Length != nil {
		fmt.Fprintf(&b, "SetupInfo payload length: %d\n", *r.Read.Length)
	}
	for i, v := range r.Read.Counters {
		fmt.Fprintf(&b, "Raw Counter %d: %d — UNVERIFIED\n", i+1, v)
	}
	fmt.Fprintf(&b, "Counter meanings validated: %t\n", r.Read.Validated)
	b.WriteString("\nUSB discovery does not prove Lab communication or counter meanings.\nNo L10 shutter/power mapping is validated.\nNo serial, unique device path, pairing key, raw history or raw reply is included.\nMissing OS metadata is recorded as not available.\nNo automatic upload. Share this report only after reviewing it.\n")
	return b.String()
}
func l10ReportJSON(r l10Report) ([]byte, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	return append(data, '\n'), err
}

// Commit JSON/TXT together; restore previous files if either rename fails.
// Both existing destinations must already have been approved by the user.
func l10SavePair(path string, r l10Report) error {
	if !strings.EqualFold(filepath.Ext(path), ".json") {
		path += ".json"
	}
	txt := strings.TrimSuffix(path, filepath.Ext(path)) + ".txt"
	js, err := l10ReportJSON(r)
	if err != nil {
		return err
	}
	paths := []string{path, txt}
	data := [][]byte{js, []byte(l10ReportText(r))}
	temps := []string{"", ""}
	backups := []string{"", ""}
	committed := []bool{false, false}
	rollback := func() {
		for i := range paths {
			if committed[i] {
				os.Remove(paths[i])
			}
			if backups[i] != "" {
				os.Rename(backups[i], paths[i])
			}
			if temps[i] != "" {
				os.Remove(temps[i])
			}
		}
	}
	for i := range paths {
		f, e := os.CreateTemp(filepath.Dir(path), ".l10-report-*")
		if e != nil {
			rollback()
			return e
		}
		temps[i] = f.Name()
		if e = f.Chmod(0600); e == nil {
			_, e = f.Write(data[i])
		}
		if ce := f.Close(); e == nil {
			e = ce
		}
		if e != nil {
			rollback()
			return e
		}
	}
	for i := range paths {
		st, e := os.Stat(paths[i])
		if e != nil && !os.IsNotExist(e) {
			rollback()
			return e
		}
		if e == nil {
			if st.IsDir() {
				rollback()
				return fmt.Errorf("destination is a directory")
			}
			f, e := os.CreateTemp(filepath.Dir(path), ".l10-backup-*")
			if e != nil {
				rollback()
				return e
			}
			name := f.Name()
			f.Close()
			os.Remove(name)
			if e = os.Rename(paths[i], name); e != nil {
				rollback()
				return e
			}
			backups[i] = name
		}
	}
	for i := range paths {
		if e := os.Rename(temps[i], paths[i]); e != nil {
			rollback()
			return e
		}
		temps[i] = ""
		committed[i] = true
	}
	for _, p := range backups {
		if p != "" {
			os.Remove(p)
		}
	}
	return nil
}
