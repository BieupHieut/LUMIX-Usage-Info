//go:build windows

package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// Only session ownership commands have a no-data transport entry point.
func (d *WPDDevice) SendSession(op uint32) (uint32, error) {
	d.trace = boardTransferTrace{Phase: "session_request"}
	if !d.sessionAccess {
		return 0, fmt.Errorf("session_access_required")
	}
	p, e := newSessionCommand(op)
	if e != nil {
		return 0, &boardWPDError{Origin: "request_build", Err: e}
	}
	defer release(p)
	r, e := d.send(p)
	if e != nil {
		return 0, e
	}
	defer release(r)
	d.trace.ResponseParams = boardResponseParameters(r)
	rc, e := valuesGetUI4(r, &PK_MTP_RESPONSE_CODE)
	if e != nil {
		return 0, &boardWPDError{Origin: "response_code", Err: e}
	}
	d.trace.Phase = "complete"
	return rc, nil
}

const boardSessionIOCTLAccess uint32 = 3 // FILE_READ_ACCESS | FILE_WRITE_ACCESS

// WPD access is a transport permission. Only 9102/9103 are permitted here;
// no camera setting, file write, or service-mode command is exposed.
func newSessionCommand(op uint32) (p unsafe.Pointer, e error) {
	if op != 0x9102 && op != 0x9103 {
		return nil, fmt.Errorf("session opcode not allowed")
	}
	p, e = setupCommandValues(12)
	if e != nil {
		return nil, e
	}
	defer func() {
		if e != nil {
			release(p)
			p = nil
		}
	}()
	if e = valuesSetUI4(p, &PK_IOCTL_ACCESS, boardSessionIOCTLAccess); e != nil {
		return p, e
	}
	if e = valuesSetUI4(p, &PK_MTP_OPERATION_CODE, op); e != nil {
		return p, e
	}
	var args []uint32
	if op == 0x9102 {
		args = []uint32{0x00010001}
	}
	c, e := newUI4Collection(args)
	if e != nil {
		return p, e
	}
	defer release(c)
	if e = valuesSetPVC(p, &PK_MTP_OPERATION_PARAMS, c); e != nil {
		return p, e
	}
	return p, nil
}

type boardWPDError struct {
	Origin string
	Err    error
	PTP    uint32
}

func (e *boardWPDError) Error() string { return e.Err.Error() }
func (e *boardWPDError) Unwrap() error { return e.Err }

var boardWPDOrigins = map[string]bool{"request_build": true, "send_command": true, "driver_hresult": true, "response_code": true}

type boardTransport interface {
	SendReadPTP(uint32, []uint32) ([]byte, uint32, error)
	SendSession(uint32) (uint32, error)
	Close()
}
type boardStep struct {
	Stage        string              `json:"stage"`
	Status       string              `json:"status"`
	PTP          string              `json:"ptp,omitempty"`
	Code         string              `json:"code,omitempty"`
	ErrorSource  string              `json:"error_source,omitempty"`
	Milliseconds int64               `json:"milliseconds"`
	Scenario     string              `json:"scenario,omitempty"`
	Opcode       uint32              `json:"opcode,omitempty"`
	Parameters   []uint32            `json:"parameters,omitempty"`
	Counters     []uint32            `json:"counters_1_to_7,omitempty"`
	Transfer     *boardTransferTrace `json:"transfer,omitempty"`
}
type boardDiagnostic struct {
	RunID        string             `json:"run_id,omitempty"`
	Schema       int                `json:"schema"`
	Outcome      string             `json:"suite_outcome,omitempty"`
	Operations   []uint16           `json:"advertised_operations,omitempty"`
	ActiveStage  string             `json:"active_stage,omitempty"`
	Version      string             `json:"app_version"`
	Started      time.Time          `json:"started"`
	Model        string             `json:"model,omitempty"`
	Firmware     string             `json:"firmware,omitempty"`
	Policy       string             `json:"connection_policy"`
	DeviceAccess string             `json:"device_access,omitempty"`
	Steps        []boardStep        `json:"steps"`
	FailedStage  string             `json:"failed_stage,omitempty"`
	Camera       *boardCameraStatus `json:"camera_status,omitempty"`
}
type boardReply struct {
	Kind       string             `json:"kind"`
	Result     *UsageResult       `json:"result,omitempty"` // Private parent/child pipe, never diagnostic export.
	Valid      bool               `json:"valid"`
	Diagnostic boardDiagnostic    `json:"diagnostic"`
	Devices    []DeviceInfo       `json:"devices,omitempty"` // Private selection only.
	Camera     *boardCameraStatus `json:"camera_status,omitempty"`
}
type boardRequest struct {
	Device  string
	Session bool
	Extras  bool
	Suite   bool
}

var boardResponsePattern = regexp.MustCompile(`^(?:0x|0X)[0-9A-Fa-f]{4,8}$`)
var boardModelPattern = regexp.MustCompile(`^D[CM]-[A-Za-z0-9-]{1,24}$`)
var boardHexPattern = regexp.MustCompile(`0x[0-9A-Fa-f]{8}\b`)
var boardSafeCode = map[string]bool{"storage_not_addressable": true, "not_needed": true, "device_changed": true, "cancelled": true, "transport_error": true, "no_camera": true, "worker_incomplete": true, "unsupported_trial_pair": true, "invalid_device_info": true, "unsupported_structure": true}

func boardErrorCode(err error) string {
	if err == nil {
		return ""
	}
	if m := boardHexPattern.FindString(err.Error()); m != "" {
		return strings.ToUpper(m)
	}
	if boardSafeCode[err.Error()] {
		return err.Error()
	}
	return "transport_error"
}
func newBoardDiagnostic(session bool) boardDiagnostic {
	p := "legacy"
	if session {
		p = "diagnostic-session"
	}
	access := "GENERIC_READ"
	if session {
		access = "GENERIC_READ|GENERIC_WRITE (session transport only)"
	}
	var nonce [16]byte
	_, _ = rand.Read(nonce[:])
	return boardDiagnostic{RunID: fmt.Sprintf("%x", nonce), Schema: 2, Version: appVersion, Started: time.Now(), Policy: p, DeviceAccess: access, Steps: []boardStep{}}
}
func (d *boardDiagnostic) step(stage, status string, rc uint32, err error, since time.Time) {
	s := boardStep{Stage: stage, Status: status, Milliseconds: time.Since(since).Milliseconds(), Code: boardErrorCode(err)}
	var we *boardWPDError
	if errors.As(err, &we) {
		if boardWPDOrigins[we.Origin] {
			s.ErrorSource = we.Origin
		}
		if rc == 0 {
			rc = we.PTP
		}
	}
	if rc != 0 {
		s.PTP = fmt.Sprintf("0x%04X", rc)
	}
	d.Steps = append(d.Steps, s)
	if status == "failed" || status == "unconfirmed" {
		if d.FailedStage == "" {
			d.FailedStage = stage
		}
	}
}
func boardQuery(dev boardTransport, session bool, emit func(boardReply)) (out boardReply) {
	return boardQueryOptions(dev, session, false, emit)
}
func boardQueryOptions(dev boardTransport, session, extras bool, emit func(boardReply)) (out boardReply) {
	out = boardReply{Kind: "complete", Diagnostic: newBoardDiagnostic(session)}
	owned := false
	send := func(kind string) { v := out; v.Kind = kind; emit(v) }
	defer func() {
		if owned {
			send("closing")
			start := time.Now()
			rc, err := dev.SendSession(0x9103)
			status := "ok"
			if err != nil || rc != PTP_RC_OK {
				status = "failed"
			}
			out.Diagnostic.step("close_session", status, rc, err, start)
		}
		dev.Close()
		out.Diagnostic.step("release_device", "ok", 0, nil, time.Now())
		send("complete")
	}()
	start := time.Now()
	send("identifying")
	b, rc, err := dev.SendReadPTP(PTP_GET_DEVICE_INFO, nil)
	if err != nil || rc != PTP_RC_OK {
		out.Diagnostic.step("device_info", "failed", rc, err, start)
		return
	}
	di, err := parsePTPDeviceInfo(b)
	if err != nil || !boardModelPattern.MatchString(strings.TrimSpace(di.Model)) {
		out.Diagnostic.step("device_info", "failed", rc, fmt.Errorf("invalid_device_info"), start)
		return
	}
	r := UsageResult{Manufacturer: strings.TrimSpace(di.Manufacturer), Model: strings.TrimSpace(di.Model), Firmware: normalizeFirmware(di.DeviceVersion), Serial: strings.TrimSpace(di.Serial), Timestamp: time.Now()}
	out.Result = &r
	out.Diagnostic.Model = r.Model
	if l10VersionPattern.MatchString(strings.TrimSpace(di.DeviceVersion)) {
		out.Diagnostic.Firmware = strings.TrimSpace(di.DeviceVersion)
	}
	out.Diagnostic.step("device_info", "ok", rc, nil, start)
	send("identified")
	if session {
		// This first trial is scoped to the exact camera/firmware from the received log.
		if r.Model != "DC-S1" || firmwareKey(r.Firmware) != "2.4" {
			out.Diagnostic.step("session_policy", "failed", 0, fmt.Errorf("unsupported_trial_pair"), time.Now())
			return
		}
		send("preparing")
		start = time.Now()
		rc, err = dev.SendSession(0x9102)
		if err != nil || rc != PTP_RC_OK {
			out.Diagnostic.step("open_session", "failed", rc, err, start)
			return
		}
		owned = true
		out.Diagnostic.step("open_session", "ok", rc, nil, start)
	} else {
		out.Diagnostic.step("open_session", "skipped", 0, nil, time.Now())
	}

	func() {
		send("reading")
		start = time.Now()
		raw, rc, err := dev.SendReadPTP(PANASONIC_GET_SETUP_INFO, []uint32{SETUP_INFO_TAG})
		if err != nil || rc != PTP_RC_OK {
			out.Diagnostic.step("usage", "failed", rc, err, start)
			return
		}
		counts, err := parseUsageRaw(raw)
		if err != nil {
			out.Diagnostic.step("parse_usage", "failed", rc, err, start)
			return
		}
		if r.Model == "DC-L10" && l10ParseCounters(raw, rc).Status != "PASS" {
			out.Diagnostic.step("parse_usage", "failed", rc, fmt.Errorf("unsupported_structure"), start)
			return
		}
		r.PowerWake, r.Shutter = counts.PowerWake, counts.Shutter
		r.Raw3, r.Raw4, r.Raw5, r.Raw6, r.Raw7 = counts.Raw3, counts.Raw4, counts.Raw5, counts.Raw6, counts.Raw7
		r.Errors, r.RawLen = counts.Errors, len(raw)
		applyValidationStatus(&r)
		out.Valid = true
		out.Diagnostic.step("usage", "ok", rc, nil, start)
		// Deliver valid counters before cleanup; a blocked cleanup must not erase them.
		send("usage_complete")
	}()
	if extras {
		boardReadCameraStatus(dev, &out, send)
	}
	return
}
func boardWorker() int {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	enc := json.NewEncoder(os.Stdout)
	emit := func(v boardReply) { _ = enc.Encode(v) }
	var req boardRequest
	if json.NewDecoder(io.LimitReader(os.Stdin, 65536)).Decode(&req) != nil {
		return 2
	}
	d := newBoardDiagnostic(req.Session)
	if req.Suite {
		d.Policy = "s1-comprehensive-v1"
		d.Schema = 3
		d.Outcome = "stopped"
	}
	fail := func(stage string, err error) {
		d.step(stage, "failed", 0, err, time.Now())
		emit(boardReply{Kind: "complete", Diagnostic: d})
	}
	hr, _, _ := procCoInitializeEx.Call(0, COINIT_APARTMENTTHREADED)
	if hrFailed(hr) {
		fail("initialize", fmt.Errorf("%s", hrString(hr)))
		return 0
	}
	defer procCoUninitialize.Call()
	ds, err := listDevices()
	if err != nil {
		fail("enumerate", err)
		return 0
	}
	candidates := []DeviceInfo{}
	for _, v := range ds {
		if isLumix(v) {
			candidates = append(candidates, v)
		}
	}
	if len(candidates) == 0 {
		fail("enumerate", fmt.Errorf("no_camera"))
		return 0
	}
	chosen := -1
	for i, v := range candidates {
		if v.PnPID == req.Device {
			chosen = i
		}
	}
	if chosen < 0 && req.Device == "" && len(candidates) == 1 {
		chosen = 0
	}
	if chosen < 0 {
		emit(boardReply{Kind: "select_device", Devices: candidates, Diagnostic: d})
		return 0
	}
	if req.Suite {
		boardSuite(func(session bool) (boardTransport, error) { return openDeviceForSession(candidates[chosen], session) }, emit, time.Sleep)
		return 0
	}
	dev, err := openDeviceForSession(candidates[chosen], req.Session)
	if err != nil {
		fail("open_device", err)
		return 0
	}
	boardQueryOptions(dev, req.Session, req.Extras, emit)
	return 0
}
func boardRunWorker(req boardRequest, progress func(boardReply)) (boardReply, error) {
	return boardRunWorkerContext(context.Background(), req, progress)
}
func boardRunWorkerContext(parent context.Context, req boardRequest, progress func(boardReply)) (boardReply, error) {
	latest := boardReply{Diagnostic: newBoardDiagnostic(req.Session)}
	if req.Suite {
		latest.Diagnostic.Policy = "s1-comprehensive-v1"
		latest.Diagnostic.Schema = 3
	}
	fail := func(err error) (boardReply, error) {
		latest.Kind = "complete"
		latest.Diagnostic.Outcome = "interrupted"
		latest.Diagnostic.step("worker", "failed", 0, err, time.Now())
		return latest, fmt.Errorf("worker_incomplete")
	}
	exe, err := os.Executable()
	if err != nil {
		return fail(err)
	}
	limit := 20 * time.Second
	if req.Extras {
		limit = 90 * time.Second
	}
	if req.Suite {
		limit = 120 * time.Second
	}
	ctx, cancel := context.WithTimeout(parent, limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "--board-worker")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: CREATE_NO_WINDOW}
	in, _ := json.Marshal(req)
	cmd.Stdin = strings.NewReader(string(in))
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return fail(err)
	}
	if err = cmd.Start(); err != nil {
		return fail(err)
	}
	idle := time.AfterFunc(20*time.Second, cancel)
	defer idle.Stop()

	scan := bufio.NewScanner(pipe)
	scan.Buffer(make([]byte, 4096), 1024*1024)
	for scan.Scan() {
		var v boardReply
		if e := json.Unmarshal(scan.Bytes(), &v); e != nil {
			err = e
			break
		}
		idle.Reset(20 * time.Second)
		latest = v
		progress(v)
	}
	if err != nil {
		_ = cmd.Process.Kill()
	}
	waitErr := cmd.Wait()
	if err == nil {
		err = scan.Err()
	}
	if err == nil {
		err = waitErr
	}
	if ctx.Err() != nil || err != nil || latest.Kind != "complete" && latest.Kind != "select_device" {
		latest.Diagnostic.Outcome = "interrupted"
		code := "worker_incomplete"
		if parent.Err() != nil {
			code = "cancelled"
		}
		latest.Diagnostic.step("worker", "failed", 0, fmt.Errorf("%s", code), time.Now())
		if req.Session || req.Suite {
			latest.Diagnostic.step("close_session", "unconfirmed", 0, nil, time.Now())
		}
		latest.Kind = "complete"
		// Only the structured error is retained; process errors can contain user paths.
		return latest, fmt.Errorf("worker_incomplete")
	}
	return latest, nil
}

var boardReadAccess = PROPERTYKEY{GUID{0x204d9f0c, 0x2292, 0x4080, [8]byte{0x9f, 0x42, 0x40, 0x66, 0x4e, 0x70, 0xf8, 0x59}}, 9}

func boardSetReadAccess(client unsafe.Pointer) error {
	return boardSetDeviceAccess(client, false)
}
func boardSetDeviceAccess(client unsafe.Pointer, session bool) error {
	access := uint32(0x80000000)
	if session {
		access |= 0x40000000
	}
	return valuesSetUI4(client, &boardReadAccess, access)
}

func boardReadWords(raw []byte, tag uint32, words int) ([]uint16, error) {
	frames, e := checkedServiceFrames(raw)
	if e != nil {
		return nil, e
	}
	var result []uint16
	for _, f := range frames {
		if f.Tag != tag {
			continue
		}
		if result != nil || len(f.Data) != words*2 {
			return nil, fmt.Errorf("unexpected reply structure")
		}
		result = make([]uint16, words)
		for i := range result {
			result[i] = binary.LittleEndian.Uint16(f.Data[i*2:])
		}
	}
	if result == nil {
		return nil, fmt.Errorf("missing reply")
	}
	return result, nil
}
