//go:build windows

package main

// L10 diagnostics are independent of the normal camera/report pipeline. Device
// paths stay in the private worker/session model; reports are built by whitelist.
import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"unsafe"
)

type l10Interface struct {
	Number    string `json:"number"`
	Class     string `json:"class"`
	Subclass  string `json:"subclass"`
	Protocol  string `json:"protocol"`
	Endpoints string `json:"endpoints"`
}
type l10USBType struct {
	Name             string         `json:"display_name"`
	VID              string         `json:"vid"`
	PID              string         `json:"pid"`
	Driver           string         `json:"driver"`
	Interfaces       []l10Interface `json:"interfaces"`
	WPDExposed       bool           `json:"wpd_path_exposed"`
	WPDStatus        string         `json:"wpd_enumeration_status"`
	NewSinceBaseline *bool          `json:"new_since_baseline"`
}
type l10Candidate struct {
	USB       l10USBType
	ModelHint string
	USBID     string   // Private OS instance ID: never included in exported report.
	WPDIDs    []string // Private paths, passed to the child on stdin, not command line.
}
type l10PTPResult struct {
	Uncertain       bool     `json:"command_execution_uncertain"`
	Attempted       bool     `json:"attempted"`
	Status          string   `json:"get_device_info_status"`
	Reason          string   `json:"reason"`
	Response        *uint32  `json:"ptp_response_code"`
	Model           string   `json:"model_confirmed"`
	Firmware        string   `json:"firmware_observed"`
	StandardVersion uint16   `json:"standard_version"`
	VendorExtension uint32   `json:"vendor_extension_id"`
	Operations      []uint16 `json:"supported_operations"`
}
type l10ReadResult struct {
	Uncertain   bool     `json:"command_execution_uncertain"`
	Attempted   bool     `json:"attempted"`
	Status      string   `json:"response_status"`
	Reason      string   `json:"reason"`
	Response    *uint32  `json:"ptp_response_code"`
	ExpectedTag *bool    `json:"expected_tag_found"`
	Length      *int     `json:"response_length"` // SetupInfo payload length, not TLV frame length.
	Counters    []uint64 `json:"raw_counters_unverified"`
	Validated   bool     `json:"counter_meanings_validated"`
}
type l10WorkerRequest struct {
	Stage    string
	Target   l10Candidate
	Consent  bool
	Identity string // Session-private fingerprint, never exported.
}
type l10WorkerReply struct {
	Status     string
	Reason     string
	Candidates []l10Candidate
	PTP        l10PTPResult
	Read       l10ReadResult
	Identity   string
}

var (
	l10VIDPID         = regexp.MustCompile(`(?i)VID_([0-9A-F]{4})&PID_([0-9A-F]{4})`)
	l10ModelPattern   = regexp.MustCompile(`(?i)\b(?:DC|DMC)-[A-Z0-9]+\b`)
	l10VersionPattern = regexp.MustCompile(`^(?:[Vv][Ee][Rr]\. ?)?[0-9]{1,3}(?:\.[0-9]{1,3}){1,3}$`)
	l10Compatible     = regexp.MustCompile(`(?i)Class_([0-9A-F]{2})(?:&SubClass_([0-9A-F]{2}))?(?:&Prot_([0-9A-F]{2}))?`)
	l10MI             = regexp.MustCompile(`(?i)&MI_([0-9A-F]{2})`)
	l10DriverPattern  = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,63}$`)
	l10SetupDLL       = syscall.NewLazyDLL("setupapi.dll")
	l10ConfigDLL      = syscall.NewLazyDLL("cfgmgr32.dll")
	l10GetClassDevs   = l10SetupDLL.NewProc("SetupDiGetClassDevsW")
	l10EnumDev        = l10SetupDLL.NewProc("SetupDiEnumDeviceInfo")
	l10GetID          = l10SetupDLL.NewProc("SetupDiGetDeviceInstanceIdW")
	l10GetProperty    = l10SetupDLL.NewProc("SetupDiGetDeviceRegistryPropertyW")
	l10DestroyList    = l10SetupDLL.NewProc("SetupDiDestroyDeviceInfoList")
	l10LocateNode     = l10ConfigDLL.NewProc("CM_Locate_DevNodeW")
	l10GetParent      = l10ConfigDLL.NewProc("CM_Get_Parent")
	l10NodeID         = l10ConfigDLL.NewProc("CM_Get_Device_IDW")
)

type l10DevInfo struct {
	Size     uint32
	Class    GUID
	Instance uint32
	Reserved uintptr
}

func l10Property(set uintptr, dev *l10DevInfo, key uint32) string {
	var size, kind uint32
	l10GetProperty.Call(set, uintptr(unsafe.Pointer(dev)), uintptr(key), uintptr(unsafe.Pointer(&kind)), 0, 0, uintptr(unsafe.Pointer(&size)))
	if size == 0 || size > 65536 || size%2 != 0 {
		return ""
	}
	buf := make([]uint16, size/2)
	ok, _, _ := l10GetProperty.Call(set, uintptr(unsafe.Pointer(dev)), uintptr(key), uintptr(unsafe.Pointer(&kind)), uintptr(unsafe.Pointer(&buf[0])), uintptr(size), 0)
	if ok == 0 {
		return ""
	}
	// Registry multi-strings are flattened only for matching class/VID metadata.
	for i := range buf {
		if buf[i] == 0 {
			buf[i] = ' '
		}
	}
	return strings.TrimSpace(syscall.UTF16ToString(buf))
}
func l10InstanceID(set uintptr, dev *l10DevInfo) string {
	buf := make([]uint16, 1024)
	ok, _, _ := l10GetID.Call(set, uintptr(unsafe.Pointer(dev)), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), 0)
	if ok == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}
func l10NormalizePath(path string) string {
	path = strings.ToUpper(strings.TrimSpace(path))
	path = strings.TrimPrefix(path, `\\?\`)
	if i := strings.Index(path, "#{"); i >= 0 {
		path = path[:i]
	}
	return strings.ReplaceAll(path, "#", `\`)
}
func l10Ancestry(path string) []string {
	id := l10NormalizePath(path)
	out := []string{id}
	var node uint32
	if rc, _, _ := l10LocateNode.Call(uintptr(unsafe.Pointer(&node)), uintptr(unsafe.Pointer(wstr(id))), 0); rc != 0 {
		return out
	}
	for i := 0; i < 12; i++ {
		buf := make([]uint16, 1024)
		if rc, _, _ := l10NodeID.Call(uintptr(node), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), 0); rc == 0 {
			out = append(out, l10NormalizePath(syscall.UTF16ToString(buf)))
		}
		var parent uint32
		if rc, _, _ := l10GetParent.Call(uintptr(unsafe.Pointer(&parent)), uintptr(node), 0); rc != 0 {
			break
		}
		node = parent
	}
	return out
}
func l10RootID(path string) string {
	root := l10NormalizePath(path)
	for _, p := range l10Ancestry(path) {
		if strings.HasPrefix(p, `USB\VID_04DA&PID_`) {
			if !l10MI.MatchString(p) {
				return p
			}
		}
	}
	return root
}
func l10Enumerate() l10WorkerReply {
	reply := l10WorkerReply{Status: "PASS", Reason: "OS_ENUMERATION_ONLY"}
	enumerator := wstr("USB")
	set, _, _ := l10GetClassDevs.Call(0, uintptr(unsafe.Pointer(enumerator)), 0, 6)
	if set == ^uintptr(0) {
		reply.Status = "ERROR"
		reply.Reason = "USB_ENUMERATION_FAILED"
		return reply
	}
	defer l10DestroyList.Call(set)
	candidates := map[string]*l10Candidate{}
	for i := uint32(0); ; i++ {
		dev := l10DevInfo{Size: uint32(unsafe.Sizeof(l10DevInfo{}))}
		ok, _, err := l10EnumDev.Call(set, uintptr(i), uintptr(unsafe.Pointer(&dev)))
		if ok == 0 {
			if err != syscall.Errno(259) {
				reply.Status = "ERROR"
				reply.Reason = "USB_ENUMERATION_INCOMPLETE"
			}
			break
		}
		id := l10InstanceID(set, &dev)
		vp := l10VIDPID.FindStringSubmatch(id)
		if len(vp) != 3 || !strings.EqualFold(vp[1], "04DA") {
			continue
		}
		root := l10RootID(id)
		c := candidates[root]
		if c == nil {
			c = &l10Candidate{USBID: root, USB: l10USBType{Name: "Panasonic USB device", VID: "04DA", PID: strings.ToUpper(vp[2]), Driver: "not available", Interfaces: []l10Interface{}, WPDStatus: "NOT_RUN"}}
			candidates[root] = c
		}
		name := l10Property(set, &dev, 12) + " " + l10Property(set, &dev, 0)
		if model := l10ModelPattern.FindString(name); model != "" {
			model = strings.ToUpper(model)
			if c.ModelHint == "" || model == "DC-L10" {
				c.ModelHint = model
				c.USB.Name = model
			}
		}
		if service := l10Property(set, &dev, 4); l10DriverPattern.MatchString(service) {
			c.USB.Driver = service
		}
		compat := l10Property(set, &dev, 2)
		mi := "not available"
		if m := l10MI.FindStringSubmatch(id); len(m) == 2 {
			mi = strings.ToUpper(m[1])
		}
		entry := l10Interface{Number: mi, Class: "not available", Subclass: "not available", Protocol: "not available", Endpoints: "not available"}
		if cl := l10Compatible.FindStringSubmatch(compat); len(cl) == 4 {
			entry.Class = strings.ToUpper(cl[1])
			if cl[2] != "" {
				entry.Subclass = strings.ToUpper(cl[2])
			}
			if cl[3] != "" {
				entry.Protocol = strings.ToUpper(cl[3])
			}
		}
		c.USB.Interfaces = append(c.USB.Interfaces, entry)
	}
	// Enumeration is metadata only: no device is opened and no PTP operation is sent.
	ds, err := listDevices()
	for _, c := range candidates {
		c.USB.WPDStatus = "PASS"
		if err != nil {
			c.USB.WPDStatus = "INACCESSIBLE"
		}
		for _, d := range ds {
			matched := false
			for _, parent := range l10Ancestry(d.PnPID) {
				if parent == c.USBID {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
			c.WPDIDs = append(c.WPDIDs, d.PnPID)
			if model := l10ModelPattern.FindString(d.FriendlyName + " " + d.Description); model != "" {
				c.ModelHint = strings.ToUpper(model)
				c.USB.Name = c.ModelHint
			}
		}
		c.USB.WPDExposed = len(c.WPDIDs) > 0
		reply.Candidates = append(reply.Candidates, *c)
	}
	sort.Slice(reply.Candidates, func(i, j int) bool { return reply.Candidates[i].USBID < reply.Candidates[j].USBID })
	if reply.Status == "PASS" && len(reply.Candidates) == 0 {
		reply.Status = "DEVICE_NOT_FOUND"
		reply.Reason = "NO_PANASONIC_USB_CANDIDATE"
	}
	return reply
}
func l10Recognized(d DeviceInfo) bool {
	return strings.EqualFold(l10ModelPattern.FindString(d.FriendlyName+" "+d.Description), "DC-L10")
}
func l10AllowedCandidate(c l10Candidate) bool {
	return c.USB.VID == "04DA" && (c.ModelHint == "" || c.ModelHint == "DC-L10")
}
func l10HasOperation(ops []uint16, op uint16) bool {
	for _, v := range ops {
		if v == op {
			return true
		}
	}
	return false
}
func l10ParsePTP(data []byte) (l10PTPResult, string, error) {
	out := l10PTPResult{Attempted: true, Status: "ERROR", Reason: "MALFORMED_DEVICE_INFO", Operations: []uint16{}}
	di, err := parsePTPDeviceInfo(data)
	if err != nil {
		return out, "", err
	}
	off := 8
	if _, ok := readPTPString(data, &off); !ok {
		return out, "", errors.New("vendor string")
	}
	off += 2
	if off+4 > len(data) {
		return out, "", errors.New("operation length")
	}
	n := uint64(binary.LittleEndian.Uint32(data[off : off+4]))
	off += 4
	if n > 65536 || uint64(off)+n*2 > uint64(len(data)) {
		return out, "", errors.New("operation bounds")
	}
	out.StandardVersion = binary.LittleEndian.Uint16(data[:2])
	out.VendorExtension = binary.LittleEndian.Uint32(data[2:6])
	for i := uint64(0); i < n; i++ {
		out.Operations = append(out.Operations, binary.LittleEndian.Uint16(data[off+int(i)*2:]))
	}
	if !strings.EqualFold(strings.TrimSpace(di.Model), "DC-L10") {
		out.Reason = "MODEL_MISMATCH"
		out.Model = "NOT_L10"
		return out, "", errors.New("not DC-L10")
	}
	out.Model = "DC-L10"
	out.Firmware = "not available"
	if fw := strings.TrimSpace(di.DeviceVersion); l10VersionPattern.MatchString(fw) {
		out.Firmware = fw
	}
	out.Status = "PASS"
	out.Reason = "STANDARD_PTP_RESPONSE_ONLY"
	if strings.TrimSpace(di.Serial) == "" {
		out.Reason = "DEVICE_IDENTITY_UNAVAILABLE"
		return out, "", nil
	}
	// Used only to prevent a different observed device/firmware from inheriting B consent.
	digest := sha256.Sum256([]byte(di.Model + "\x00" + di.Serial + "\x00" + di.DeviceVersion))
	return out, hex.EncodeToString(digest[:]), nil
}
func l10ParseCounters(raw []byte, rc uint32) l10ReadResult {
	out := l10ReadResult{Attempted: true, Status: "ERROR", Reason: "TAG_NOT_FOUND", Response: &rc}
	found := false
	out.ExpectedTag = &found
	if rc != PTP_RC_OK {
		out.Status, out.Reason = l10PTPFailure(rc, "READ_OPERATION_UNSUPPORTED")
		return out
	}
	// Strict TLV walk: malformed tails, duplicate reply tags and non-146-byte
	// payloads are rejected before using the legacy parser, which accepts >=146.
	var payload []byte
	for off := 0; off < len(raw); {
		if off+8 > len(raw) {
			out.Reason = "MALFORMED_TLV"
			return out
		}
		tag := binary.LittleEndian.Uint32(raw[off:])
		n := uint64(binary.LittleEndian.Uint32(raw[off+4:]))
		if uint64(off)+8+n > uint64(len(raw)) {
			out.Reason = "MALFORMED_TLV"
			return out
		}
		if tag == SETUP_INFO_REPLY {
			if found {
				out.Reason = "DUPLICATE_REPLY_TAG"
				return out
			}
			found = true
			size := int(n)
			out.Length = &size
			payload = raw[off+8 : off+8+int(n)]
		}
		off += 8 + int(n)
	}
	if !found {
		return out
	}
	if len(payload) != 146 {
		out.Reason = "LENGTH_MISMATCH"
		return out
	}
	legacy, err := parseUsageRaw(raw)
	if err != nil {
		out.Reason = "PARSER_REJECTED"
		return out
	}
	out.Counters = []uint64{uint64(legacy.PowerWake), uint64(legacy.Shutter), uint64(legacy.Raw3), uint64(legacy.Raw4), uint64(legacy.Raw5), uint64(legacy.Raw6), uint64(legacy.Raw7)}
	out.Status = "PASS"
	out.Reason = "RAW_COUNTERS_ONLY_UNVERIFIED"
	// legacy.Errors and the original response are intentionally discarded.
	return out
}
func l10PTPFailure(rc uint32, unsupported string) (string, string) {
	switch rc {
	case 0x2005:
		return "UNSUPPORTED", unsupported
	case 0x200F:
		return "INACCESSIBLE", "ACCESS_DENIED"
	case 0x2019:
		return "INACCESSIBLE", "DEVICE_BUSY"
	case 0x201F:
		return "CANCELLED", "TRANSACTION_CANCELLED"
	default:
		return "ERROR", "PTP_RESPONSE_REJECTED"
	}
}
func l10Failure(err error) (string, string) {
	s := strings.ToUpper(err.Error())
	switch {
	case strings.Contains(s, "80070005"):
		return "INACCESSIBLE", "ACCESS_DENIED"
	case strings.Contains(s, "80070020"), strings.Contains(s, "800700AA"), strings.Contains(s, "80070021"):
		return "INACCESSIBLE", "DEVICE_BUSY"
	case strings.Contains(s, "8007048F"), strings.Contains(s, "8007001F"), strings.Contains(s, "80070002"):
		return "INACCESSIBLE", "DEVICE_DISCONNECTED"
	case strings.Contains(s, "80070032"), strings.Contains(s, "80004001"):
		return "UNSUPPORTED", "PTP_NOT_EXPOSED"
	default:
		return "ERROR", "WPD_CALL_FAILED"
	}
}
func l10Execute(req l10WorkerRequest) l10WorkerReply {
	out := l10WorkerReply{Status: "ERROR", Reason: "INVALID_REQUEST", PTP: l10PTPResult{Status: "NOT_RUN", Operations: []uint16{}}, Read: l10ReadResult{Status: "NOT_RUN"}}
	if req.Stage != "baseline" && req.Stage != "A" && req.Stage != "B" && req.Stage != "C" {
		return out
	}
	if (req.Stage == "B" || req.Stage == "C") && (!req.Consent || !l10AllowedCandidate(req.Target)) {
		out.Reason = "CONSENT_OR_TARGET_MISSING"
		return out
	}
	if req.Stage == "C" && len(req.Identity) != 64 {
		out.Reason = "READ_GATE_NOT_MET"
		return out
	}
	fresh := l10Enumerate()
	if req.Stage == "A" || req.Stage == "baseline" {
		return fresh
	}
	if fresh.Status != "PASS" {
		out.Status, out.Reason = fresh.Status, fresh.Reason
		return out
	}
	if !req.Consent || !l10AllowedCandidate(req.Target) {
		out.Reason = "CONSENT_OR_TARGET_MISSING"
		return out
	}
	// Match the private instance selected in A; VID/PID alone never selects a camera.
	var selected *l10Candidate
	for i := range fresh.Candidates {
		if fresh.Candidates[i].USBID == req.Target.USBID {
			selected = &fresh.Candidates[i]
			break
		}
	}
	if selected == nil {
		out.Status = "INACCESSIBLE"
		out.Reason = "DEVICE_DISCONNECTED"
		return out
	}
	if !l10AllowedCandidate(*selected) {
		out.Reason = "MODEL_MISMATCH"
		return out
	}
	if selected.USB.WPDStatus != "PASS" || len(selected.WPDIDs) != 1 {
		out.Status = "UNSUPPORTED"
		out.Reason = "PTP_NOT_EXPOSED"
		if len(selected.WPDIDs) > 1 {
			out.Reason = "AMBIGUOUS_WPD_PATH"
		}
		return out
	}
	dev, err := openDevice(DeviceInfo{PnPID: selected.WPDIDs[0]})
	if err != nil {
		out.Status, out.Reason = l10Failure(err)
		return out
	}
	defer dev.Close()
	return l10ProbeTransport(req, dev, out)
}

type l10Transport interface {
	SendReadPTP(uint32, []uint32) ([]byte, uint32, error)
}

func l10ProbeTransport(req l10WorkerRequest, dev l10Transport, out l10WorkerReply) l10WorkerReply {
	if !req.Consent || !l10AllowedCandidate(req.Target) || (req.Stage != "B" && req.Stage != "C") {
		out.Reason = "CONSENT_OR_TARGET_MISSING"
		return out
	}
	if req.Stage == "C" && len(req.Identity) != 64 {
		out.Reason = "READ_GATE_NOT_MET"
		return out
	}
	data, rc, err := dev.SendReadPTP(PTP_GET_DEVICE_INFO, nil)
	out.PTP.Attempted = true
	if err != nil {
		out.PTP.Status, out.PTP.Reason = l10Failure(err)
		out.Status, out.Reason = out.PTP.Status, out.PTP.Reason
		return out
	}
	out.PTP.Response = &rc
	if rc != PTP_RC_OK {
		out.Status, out.Reason = l10PTPFailure(rc, "GET_DEVICE_INFO_UNSUPPORTED")
		out.PTP.Status, out.PTP.Reason = out.Status, out.Reason
		return out
	}
	info, identity, err := l10ParsePTP(data)
	info.Response = &rc
	out.PTP = info
	if err != nil {
		out.Status, out.Reason = info.Status, info.Reason
		return out
	}
	out.Identity = identity
	if req.Stage == "B" {
		out.Status = "PASS"
		out.Reason = info.Reason
		return out
	}
	if req.Identity != identity {
		out.Reason = "DEVICE_IDENTITY_CHANGED"
		return out
	}
	if !l10HasOperation(info.Operations, uint16(PANASONIC_GET_SETUP_INFO)) {
		out.Status = "UNSUPPORTED"
		out.Reason = "READ_OPERATION_NOT_ADVERTISED"
		return out
	}
	raw, rc, err := dev.SendReadPTP(PANASONIC_GET_SETUP_INFO, []uint32{SETUP_INFO_TAG})
	out.Read.Attempted = true
	if err != nil {
		out.Read.Status, out.Read.Reason = l10Failure(err)
	} else {
		out.Read = l10ParseCounters(raw, rc)
	}
	out.Status, out.Reason = out.Read.Status, out.Read.Reason
	return out
}
func runL10Worker() int {
	var req l10WorkerRequest
	if err := json.NewDecoder(io.LimitReader(os.Stdin, 65536)).Decode(&req); err != nil {
		return 2
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	hr, _, _ := procCoInitializeEx.Call(0, COINIT_APARTMENTTHREADED)
	if hrFailed(hr) {
		json.NewEncoder(os.Stdout).Encode(l10WorkerReply{Status: "ERROR", Reason: "COM_INITIALIZATION_FAILED"})
		return 0
	}
	defer procCoUninitialize.Call()
	if err := json.NewEncoder(os.Stdout).Encode(l10Execute(req)); err != nil {
		return 2
	}
	return 0
}
func l10Worker(ctx context.Context, req l10WorkerRequest) l10WorkerReply {
	fail := l10WorkerReply{Status: "ERROR", Reason: "WORKER_FAILED"}
	exe, err := os.Executable()
	if err != nil {
		return fail
	}
	data, err := json.Marshal(req)
	if err != nil {
		return fail
	}
	cmd := exec.CommandContext(ctx, exe, "--l10-worker")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: CREATE_NO_WINDOW}
	cmd.Stdin = strings.NewReader(string(data))
	raw, err := cmd.Output()
	if ctx.Err() != nil {
		fail.Status = "CANCELLED"
		fail.Reason = "USER_CANCELLED"
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			fail.Status = "TIMEOUT"
			fail.Reason = "WORKER_TIMEOUT"
		}
		return fail
	}
	if err != nil {
		return fail
	}
	var out l10WorkerReply
	if len(raw) > 1048576 || json.Unmarshal(raw, &out) != nil {
		return fail
	}
	return out
}

// Only synthetic, unidentifiable fixtures. These never enumerate or open USB.
func l10Fixture(stage string) l10WorkerReply {
	c := l10Candidate{USBID: "FIXTURE_ONLY", ModelHint: "DC-L10", WPDIDs: []string{"FIXTURE_ONLY"}, USB: l10USBType{Name: "DC-L10", VID: "04DA", PID: "0000", Driver: "fixture", WPDExposed: true, WPDStatus: "PASS", Interfaces: []l10Interface{{Number: "00", Class: "06", Subclass: "01", Protocol: "01", Endpoints: "not available"}}}}
	r := l10WorkerReply{Status: "PASS", Reason: "SYNTHETIC_FIXTURE_NOT_HARDWARE", Candidates: []l10Candidate{c}}
	if stage == "B" {
		r.PTP = l10PTPResult{Attempted: true, Status: "PASS", Reason: r.Reason, Model: "DC-L10", Firmware: "1.2", Operations: []uint16{0x1001, 0x9414}}
		r.Identity = "FIXTURE_ONLY"
	}
	if stage == "C" {
		n, tag, rc := 146, true, uint32(0x2001)
		r.Read = l10ReadResult{Attempted: true, Status: "PASS", Reason: r.Reason, Response: &rc, Length: &n, ExpectedTag: &tag, Counters: []uint64{42, 100, 0, 0, 0, 0, 0}}
	}
	return r
}
