//go:build windows

package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

//go:embed camera_profiles.json
var bundledCameraData []byte

type cameraProfile struct {
	Model    string `json:"model"`
	Firmware string `json:"firmware"`
	By       string `json:"validated_by"`
	Date     string `json:"validation_date,omitempty"`
}
type cameraRegistry struct {
	Schema   int             `json:"schema"`
	Revision string          `json:"revision"`
	Profiles []cameraProfile `json:"profiles"`
}

var cameraData = mustBundledCameraData()
var cameraDataWarning bool
var cameraPage int
var cameraPreviousRect = Rect{730, 210, 785, 239}
var cameraNextRect = Rect{795, 210, 850, 239}

func decodeCameraData(data []byte) (cameraRegistry, error) {
	var r cameraRegistry
	if len(data) > 65536 {
		return r, fmt.Errorf("camera data too large")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&r); err != nil {
		return r, err
	}
	var extra interface{}
	if d.Decode(&extra) != io.EOF {
		return r, fmt.Errorf("trailing camera data")
	}
	if r.Schema != 1 || len(r.Profiles) == 0 || len(r.Profiles) > 100 {
		return r, fmt.Errorf("unsupported camera data schema or count")
	}
	if _, err := time.Parse("2006-01-02", r.Revision); err != nil {
		return r, fmt.Errorf("invalid revision date")
	}
	seen := map[string]bool{}
	for _, p := range r.Profiles {
		if !regexp.MustCompile(`^DC-[A-Z0-9]{2,16}$`).MatchString(p.Model) || !regexp.MustCompile(`^[0-9]{1,2}\.[0-9]{1,2}$`).MatchString(p.Firmware) {
			return r, fmt.Errorf("invalid model or firmware")
		}
		if len([]rune(p.By)) == 0 || len([]rune(p.By)) > 32 || strings.ContainsAny(p.By, "\r\n\t\x00") {
			return r, fmt.Errorf("invalid validator")
		}
		if p.Date != "" {
			if _, err := time.Parse("2006-01-02", p.Date); err != nil {
				return r, fmt.Errorf("invalid validation date")
			}
		}
		key := modelKey(p.Model) + "/" + firmwareKey(p.Firmware)
		if seen[key] {
			return r, fmt.Errorf("duplicate model / firmware")
		}
		seen[key] = true
	}
	return r, nil
}
func mustBundledCameraData() cameraRegistry {
	r, err := decodeCameraData(bundledCameraData)
	if err != nil {
		panic(err)
	}
	return r
}
func loadCameraData() {
	// Resolve next to the executable, never from the shell's working directory.
	exe, err := os.Executable()
	if err != nil {
		return
	}
	f, err := os.Open(filepath.Join(filepath.Dir(exe), "camera_profiles.json"))
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		cameraDataWarning = true
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 65537))
	if err != nil {
		cameraDataWarning = true
		return
	}
	r, err := decodeCameraData(data)
	if err != nil {
		cameraDataWarning = true
		return
	}
	cameraData = r
}
func modelKey(model string) string {
	return strings.TrimPrefix(strings.ToUpper(strings.NewReplacer("-", "", "_", "", " ", "").Replace(model)), "DC")
}
func firmwareKey(fw string) string {
	fw = strings.TrimPrefix(strings.TrimPrefix(strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(fw), " ", "")), "VER."), "V")
	// A trailing zero in the minor part is how the existing reader formats firmware.
	if dot := strings.IndexByte(fw, '.'); dot >= 0 && len(fw[dot+1:]) == 2 && strings.HasSuffix(fw, "0") {
		fw = strings.TrimSuffix(fw, "0")
	}
	return fw
}
func findCameraProfile(r UsageResult) (cameraProfile, bool) {
	for _, p := range cameraData.Profiles {
		if modelKey(r.Model) == modelKey(p.Model) && firmwareKey(r.Firmware) == firmwareKey(p.Firmware) {
			return p, true
		}
	}
	return cameraProfile{}, false
}
func applyValidationStatus(r *UsageResult) {
	r.VerifiedModel = false
	for _, p := range cameraData.Profiles {
		if modelKey(r.Model) == modelKey(p.Model) {
			r.VerifiedModel = true
			break
		}
	}
	_, r.VerifiedFirmware = findCameraProfile(*r)
	r.SemanticsVerified = r.VerifiedFirmware
}
func validationDetails(r UsageResult) (string, string) {
	if p, ok := findCameraProfile(r); ok && r.SemanticsVerified {
		if p.Date == "" {
			p.Date = "Not supplied"
		}
		return p.By, p.Date
	}
	return "Not verified", "Not available"
}
func counterName(index int, r UsageResult) string {
	switch index {
	case 1:
		if r.SemanticsVerified {
			return uiText("Power / Wake Activations", "전원 / 깨우기 횟수")
		}
		return uiText("Power / Wake Activations (estimated)", "전원 / 깨우기 횟수 (추정됨)")
	case 2:
		if r.SemanticsVerified {
			return uiText("Shutter Actuations", "셔터 작동 횟수")
		}
		return uiText("Shutter Actuations (estimated)", "셔터 작동 횟수 (추정됨)")
	case 3:
		return uiText("Flash Activations (estimated)", "플래시 작동 횟수 (추정됨)")
	case 4:
		return uiText("Power-saving Events (estimated)", "절전 작동 횟수 (추정됨)")
	default:
		return uiText("Unidentified", "의미 미확인")
	}
}
