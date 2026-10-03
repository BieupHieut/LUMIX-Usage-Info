//go:build windows

package main

import (
	"context"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

const (
	appVersion     = "1.0.0"
	releaseDate    = "2026-10-03"
	validationDate = "2026-09-16"
	validatedBy    = "@bieup_hieut"
	workerTimeout  = 12 * time.Second

	CLSCTX_INPROC_SERVER     = 0x1
	COINIT_APARTMENTTHREADED = 0x2
	FILE_READ_ACCESS         = 0x0001
	VT_UI4                   = 19

	PTP_RC_OK                = 0x2001
	PTP_GET_DEVICE_INFO      = 0x1001
	PANASONIC_GET_SETUP_INFO = 0x9414
	SETUP_INFO_TAG           = 0x15C00010
	SETUP_INFO_REPLY         = 0x15C00011

	CMD_EXECUTE_DATA_TO_READ = 13
	CMD_READ_DATA            = 15
	CMD_END_DATA_TRANSFER    = 17

	WM_DESTROY               = 0x0002
	WM_PAINT                 = 0x000F
	WM_SETTINGCHANGE         = 0x001A
	WM_KEYDOWN               = 0x0100
	WM_TIMER                 = 0x0113
	WM_LBUTTONUP             = 0x0202
	WM_SETCURSOR             = 0x0020
	WM_DEVICECHANGE          = 0x0219
	DBT_DEVNODES_CHANGED     = 0x0007
	DBT_DEVICEARRIVAL        = 0x8000
	DBT_DEVICEREMOVECOMPLETE = 0x8004
	WM_APP_RESULT            = 0x8001
	SW_SHOW                  = 5
	CS_HREDRAW               = 0x0002
	CS_VREDRAW               = 0x0001
	IDC_ARROW                = 32512
	MB_OK                    = 0x00000000
	MB_ICONINFO              = 0x00000040
	MB_ICONERROR             = 0x00000010
	TRANSPARENT              = 1
	DT_LEFT                  = 0x0000
	DT_CENTER                = 0x0001
	DT_RIGHT                 = 0x0002
	DT_VCENTER               = 0x0004
	DT_SINGLELINE            = 0x0020
	DT_NOPREFIX              = 0x0800
	FW_NORMAL                = 400
	FW_SEMIBOLD              = 600
	FW_BOLD                  = 700
	NULL_PEN                 = 8
	IMAGE_BITMAP             = 0
	LR_LOADFROMFILE          = 0x0010
	LR_CREATEDIBSECTION      = 0x2000
	SRCCOPY                  = 0x00CC0020
	BI_RGB                   = 0
	DIB_RGB_COLORS           = 0
	OFN_OVERWRITEPROMPT      = 0x00000002
	OFN_NOCHANGEDIR          = 0x00000008
	OFN_PATHMUSTEXIST        = 0x00000800
	CF_UNICODETEXT           = 13
	GMEM_MOVEABLE            = 0x0002
	CREATE_NO_WINDOW         = 0x08000000
	ERROR_ALREADY_EXISTS     = 183
)

// Embedded visual assets supplied by the user. They are extracted to a temporary
// directory at runtime so the final EXE stays self-contained.
//
//go:embed assets/lumix_logo.bmp
var embeddedLogoBMP []byte

//go:embed assets/mascot_default.bmp
var embeddedMascotDefaultBMP []byte

//go:embed assets/mascot_info.bmp
var embeddedMascotInfoBMP []byte

//go:embed assets/mascot_export.bmp
var embeddedMascotExportBMP []byte

//go:embed assets/mascot_error.bmp
var embeddedMascotErrorBMP []byte

// Bieup_Hieut in-app character supplied by the creator for Developer Credits.
//
//go:embed assets/creator_character.bmp
var embeddedCreatorCharacterBMP []byte

type WorkerResponse struct {
	Result *UsageResult `json:"result,omitempty"`
	Error  string       `json:"error,omitempty"`
}

type BitmapInfo struct {
	Handle uintptr
	Width  int32
	Height int32
}

type BITMAP struct {
	BmType       int32
	BmWidth      int32
	BmHeight     int32
	BmWidthBytes int32
	BmPlanes     uint16
	BmBitsPixel  uint16
	BmBits       unsafe.Pointer
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]uint32
}

type OPENFILENAMEW struct {
	LStructSize       uint32
	HwndOwner         uintptr
	HInstance         uintptr
	LpstrFilter       *uint16
	LpstrCustomFilter *uint16
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         *uint16
	NMaxFile          uint32
	LpstrFileTitle    *uint16
	NMaxFileTitle     uint32
	LpstrInitialDir   *uint16
	LpstrTitle        *uint16
	Flags             uint32
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       *uint16
	LCustData         uintptr
	LpfnHook          uintptr
	LpTemplateName    *uint16
	PvReserved        unsafe.Pointer
	DwReserved        uint32
	FlagsEx           uint32
}

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type PROPERTYKEY struct {
	Fmtid GUID
	Pid   uint32
}

type PROPVARIANT struct {
	Vt        uint16
	Reserved1 uint16
	Reserved2 uint16
	Reserved3 uint16
	Value     uint64
	Value2    uint64
}

type DeviceInfo struct {
	PnPID        string
	FriendlyName string
	Description  string
	Manufacturer string
}

type WPDDevice struct {
	ptr  unsafe.Pointer
	info DeviceInfo
}

type TLV struct {
	Tag  uint32
	Data []byte
}

type PTPDeviceInfo struct {
	Manufacturer  string
	Model         string
	DeviceVersion string
	Serial        string
}

type ErrorRecord struct {
	Index       int
	DateRaw     uint32
	Code        uint32
	DateText    string
	CodeText    string
	Category    string
	Description string
	Known       bool
}

type UsageResult struct {
	Manufacturer      string
	Model             string
	Firmware          string
	Serial            string
	PowerWake         uint32
	Shutter           uint32
	Raw3              uint16
	Raw4              uint16
	Raw5              uint16
	Raw6              uint16
	Raw7              uint16
	Errors            []ErrorRecord
	VerifiedModel     bool
	VerifiedFirmware  bool
	SemanticsVerified bool
	RawLen            int
	Timestamp         time.Time
}

type Rect struct{ L, T, R, B int32 }
type WinRect struct{ Left, Top, Right, Bottom int32 }
type PaintStruct struct {
	Hdc         uintptr
	FErase      int32
	RcPaint     WinRect
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}
type WndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}
type Msg struct {
	Hwnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       struct{ X, Y int32 }
	LPrivate uint32
}

var (
	ole32    = syscall.NewLazyDLL("ole32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")

	procCoInitializeEx   = ole32.NewProc("CoInitializeEx")
	procCoUninitialize   = ole32.NewProc("CoUninitialize")
	procCoCreateInstance = ole32.NewProc("CoCreateInstance")
	procCoTaskMemFree    = ole32.NewProc("CoTaskMemFree")

	procGetModuleHandleW         = kernel32.NewProc("GetModuleHandleW")
	procGlobalAlloc              = kernel32.NewProc("GlobalAlloc")
	procGlobalLock               = kernel32.NewProc("GlobalLock")
	procGlobalUnlock             = kernel32.NewProc("GlobalUnlock")
	procGlobalFree               = kernel32.NewProc("GlobalFree")
	procRtlMoveMemory            = kernel32.NewProc("RtlMoveMemory")
	procCreateMutexW             = kernel32.NewProc("CreateMutexW")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
	procGetLastError             = kernel32.NewProc("GetLastError")
	procGetUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")
	procRegisterClassExW         = user32.NewProc("RegisterClassExW")
	procCreateWindowExW          = user32.NewProc("CreateWindowExW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procShowWindow               = user32.NewProc("ShowWindow")
	procUpdateWindow             = user32.NewProc("UpdateWindow")
	procSetWindowTextW           = user32.NewProc("SetWindowTextW")
	procGetMessageW              = user32.NewProc("GetMessageW")
	procTranslateMessage         = user32.NewProc("TranslateMessage")
	procDispatchMessageW         = user32.NewProc("DispatchMessageW")
	procPostQuitMessage          = user32.NewProc("PostQuitMessage")
	procPostMessageW             = user32.NewProc("PostMessageW")
	procInvalidateRect           = user32.NewProc("InvalidateRect")
	procBeginPaint               = user32.NewProc("BeginPaint")
	procEndPaint                 = user32.NewProc("EndPaint")
	procGetClientRect            = user32.NewProc("GetClientRect")
	procFillRect                 = user32.NewProc("FillRect")
	procSetTimer                 = user32.NewProc("SetTimer")
	procLoadCursorW              = user32.NewProc("LoadCursorW")
	procSetCursor                = user32.NewProc("SetCursor")
	procMessageBoxW              = user32.NewProc("MessageBoxW")
	procLoadImageW               = user32.NewProc("LoadImageW")
	procSetProcessDPIAware       = user32.NewProc("SetProcessDPIAware")
	procGetDC                    = user32.NewProc("GetDC")
	procReleaseDC                = user32.NewProc("ReleaseDC")
	procOpenClipboard            = user32.NewProc("OpenClipboard")
	procCloseClipboard           = user32.NewProc("CloseClipboard")
	procEmptyClipboard           = user32.NewProc("EmptyClipboard")
	procSetClipboardData         = user32.NewProc("SetClipboardData")

	procCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procGetStockObject         = gdi32.NewProc("GetStockObject")
	procRoundRect              = gdi32.NewProc("RoundRect")
	procSetTextColor           = gdi32.NewProc("SetTextColor")
	procSetBkMode              = gdi32.NewProc("SetBkMode")
	procDrawTextW              = user32.NewProc("DrawTextW")
	procCreateFontW            = gdi32.NewProc("CreateFontW")
	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procGetObjectW             = gdi32.NewProc("GetObjectW")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procGetDIBits              = gdi32.NewProc("GetDIBits")
	procCreateDIBSection       = gdi32.NewProc("CreateDIBSection")

	procGetSaveFileNameW     = comdlg32.NewProc("GetSaveFileNameW")
	procCommDlgExtendedError = comdlg32.NewProc("CommDlgExtendedError")
)

var (
	CLSID_PortableDeviceManager               = GUID{0x0af10cec, 0x2ecd, 0x4b92, [8]byte{0x95, 0x81, 0x34, 0xf6, 0xae, 0x06, 0x37, 0xf3}}
	IID_IPortableDeviceManager                = GUID{0xa1567595, 0x4c2f, 0x4574, [8]byte{0xa6, 0xfa, 0xec, 0xef, 0x91, 0x7b, 0x9a, 0x40}}
	CLSID_PortableDeviceFTM                   = GUID{0xf7c0039a, 0x4762, 0x488a, [8]byte{0xb4, 0xb3, 0x76, 0x0e, 0xf9, 0xa1, 0xba, 0x9b}}
	IID_IPortableDevice                       = GUID{0x625e2df8, 0x6392, 0x4cf0, [8]byte{0x9a, 0xd1, 0x3c, 0xfa, 0x5f, 0x17, 0x77, 0x5c}}
	CLSID_PortableDeviceValues                = GUID{0x0c15d503, 0xd017, 0x47ce, [8]byte{0x90, 0x16, 0x7b, 0x3f, 0x97, 0x87, 0x21, 0xcc}}
	IID_IPortableDeviceValues                 = GUID{0x6848f6f2, 0x3155, 0x4f86, [8]byte{0xb6, 0xf5, 0x26, 0x3e, 0xee, 0xab, 0x31, 0x43}}
	CLSID_PortableDevicePropVariantCollection = GUID{0x08a99e2f, 0x6d6d, 0x4b80, [8]byte{0xaf, 0x5a, 0xba, 0xf2, 0xbc, 0xbe, 0x4c, 0xb9}}
	IID_IPortableDevicePropVariantCollection  = GUID{0x89b2e422, 0x4f1b, 0x4316, [8]byte{0xbc, 0xef, 0xa4, 0x4a, 0xfe, 0xa8, 0x3e, 0xb3}}
	WPD_COMMON_GUID                           = GUID{0xf0422a9c, 0x5dc8, 0x4440, [8]byte{0xb5, 0xbd, 0x5d, 0xf2, 0x88, 0x35, 0x65, 0x8a}}
	WPD_MTP_GUID                              = GUID{0x4d545058, 0x1a2e, 0x4106, [8]byte{0xa3, 0x57, 0x77, 0x1e, 0x08, 0x19, 0xfc, 0x56}}
	WPD_API_GUID                              = GUID{0x10e54a3e, 0x052d, 0x4777, [8]byte{0xa1, 0x3c, 0xde, 0x76, 0x14, 0xbe, 0x2b, 0xc4}}

	PK_COMMON_COMMAND_CATEGORY = PROPERTYKEY{WPD_COMMON_GUID, 1001}
	PK_COMMON_COMMAND_ID       = PROPERTYKEY{WPD_COMMON_GUID, 1002}
	PK_COMMON_HRESULT          = PROPERTYKEY{WPD_COMMON_GUID, 1003}
	PK_MTP_OPERATION_CODE      = PROPERTYKEY{WPD_MTP_GUID, 1001}
	PK_MTP_OPERATION_PARAMS    = PROPERTYKEY{WPD_MTP_GUID, 1002}
	PK_MTP_RESPONSE_CODE       = PROPERTYKEY{WPD_MTP_GUID, 1003}
	PK_MTP_CONTEXT             = PROPERTYKEY{WPD_MTP_GUID, 1006}
	PK_MTP_TOTAL_SIZE          = PROPERTYKEY{WPD_MTP_GUID, 1007}
	PK_MTP_NUM_TO_READ         = PROPERTYKEY{WPD_MTP_GUID, 1008}
	PK_MTP_NUM_READ            = PROPERTYKEY{WPD_MTP_GUID, 1009}
	PK_MTP_DATA                = PROPERTYKEY{WPD_MTP_GUID, 1012}
	PK_IOCTL_ACCESS            = PROPERTYKEY{WPD_API_GUID, 3}
)

var (
	hwndMain           uintptr
	stateMu            sync.Mutex
	appState           = "waiting" // waiting, connecting, connected, disconnected, error
	lastResult         UsageResult
	lastError          string
	probeBusy          bool
	currentView        = "main" // main, details, history, guide, validation, export, cameras, versions, credits
	historyPage        int
	serialVisible      bool
	exportPreset       = "public"
	exportSerialMode   = "masked"
	deltaShutter       int64
	deltaPowerWake     int64
	hasDelta           bool
	uiNotice           string
	pngBusy            bool
	reportBusy         bool
	historyOpenedAt    time.Time
	lastProbeAt        time.Time
	className          = "LUMIXUsageInfoWindow"
	logoBitmap         BitmapInfo
	mascotBitmap       BitmapInfo
	mascotInfoBitmap   BitmapInfo
	mascotExportBitmap BitmapInfo
	mascotErrorBitmap  BitmapInfo
	creatorBitmap      BitmapInfo
	logoAssetPath      string
	mascotAssetPath    string
	mascotInfoPath     string
	mascotExportPath   string
	mascotErrorPath    string
	creatorAssetPath   string

	fontTitle, fontSubtitle, fontSection, fontLabel, fontValue, fontSmall, fontTiny, fontMicro, fontButton uintptr

	connectRect            = Rect{70, 552, 470, 616}
	welcomeValidationRect  = Rect{510, 552, 910, 616}
	refreshRect            = Rect{40, 496, 330, 558}
	historyRect            = Rect{40, 574, 330, 636}
	detailsRect            = Rect{345, 574, 635, 636}
	validationRect         = Rect{650, 574, 940, 636}
	exportRect             = Rect{345, 496, 635, 558}
	saveRect               = Rect{650, 496, 940, 558}
	guideRect              = Rect{776, 411, 910, 440}
	serialToggleRect       = Rect{826, 279, 906, 312}
	backRect               = Rect{390, 675, 590, 724}
	prevRect               = Rect{70, 675, 220, 724}
	nextRect               = Rect{760, 675, 910, 724}
	exportPublicRect       = Rect{75, 296, 450, 360}
	exportVerifyRect       = Rect{530, 296, 905, 360}
	exportMaskedRect       = Rect{75, 434, 330, 477}
	exportLast4Rect        = Rect{365, 434, 620, 477}
	exportFullRect         = Rect{655, 434, 910, 477}
	exportSaveRect         = Rect{300, 560, 680, 615}
	exportCopyRect         = Rect{690, 560, 910, 615}
	validationCamerasRect  = Rect{72, 545, 330, 592}
	validationVersionsRect = Rect{72, 675, 330, 724}
	validationCreditsRect  = Rect{362, 675, 620, 724}
	footerCreditsRect      = Rect{340, 728, 640, 770}
	footerVersionRect      = Rect{40, 728, 300, 770}
	languageButtonRect     = Rect{604, 54, 790, 90}
	languageMenuRect       = Rect{604, 98, 790, 222}
	languageAutoRect       = Rect{612, 105, 782, 140}
	languageKoreanRect     = Rect{612, 142, 782, 177}
	languageEnglishRect    = Rect{612, 179, 782, 214}
	languageMenuOpen       bool
)

func rgb(r, g, b byte) uintptr                { return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16 }
func wstr(s string) *uint16                   { p, _ := syscall.UTF16PtrFromString(s); return p }
func hrFailed(hr uintptr) bool                { return int32(uint32(hr)) < 0 }
func hrString(hr uintptr) string              { return fmt.Sprintf("0x%08X", uint32(hr)) }
func comVtbl(obj unsafe.Pointer) *[64]uintptr { return (*[64]uintptr)(*(*unsafe.Pointer)(obj)) }
func comCall(obj unsafe.Pointer, index int, args ...uintptr) uintptr {
	a := make([]uintptr, 0, len(args)+1)
	a = append(a, uintptr(obj))
	a = append(a, args...)
	r1, _, _ := syscall.SyscallN(comVtbl(obj)[index], a...)
	return r1
}
func release(obj unsafe.Pointer) {
	if obj != nil {
		comCall(obj, 2)
	}
}
func coTaskMemFree(p unsafe.Pointer) {
	if p != nil {
		procCoTaskMemFree.Call(uintptr(p))
	}
}
func coCreate(clsid, iid *GUID) (unsafe.Pointer, error) {
	var obj unsafe.Pointer
	hr, _, _ := procCoCreateInstance.Call(uintptr(unsafe.Pointer(clsid)), 0, CLSCTX_INPROC_SERVER, uintptr(unsafe.Pointer(iid)), uintptr(unsafe.Pointer(&obj)))
	if hrFailed(hr) {
		return nil, fmt.Errorf("CoCreateInstance failed: %s", hrString(hr))
	}
	return obj, nil
}
func utf16PtrToString(p *uint16) string {
	if p == nil {
		return ""
	}
	arr := (*[1 << 20]uint16)(unsafe.Pointer(p))[:]
	n := 0
	for n < len(arr) && arr[n] != 0 {
		n++
	}
	return syscall.UTF16ToString(arr[:n])
}
func strPtr(s string) (*uint16, error) { return syscall.UTF16PtrFromString(s) }

func valuesSetGuid(v unsafe.Pointer, key *PROPERTYKEY, val *GUID) error {
	hr := comCall(v, 27, uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(val)))
	if hrFailed(hr) {
		return fmt.Errorf("SetGuidValue: %s", hrString(hr))
	}
	return nil
}
func valuesSetUI4(v unsafe.Pointer, key *PROPERTYKEY, val uint32) error {
	hr := comCall(v, 9, uintptr(unsafe.Pointer(key)), uintptr(val))
	if hrFailed(hr) {
		return fmt.Errorf("SetUnsignedIntegerValue: %s", hrString(hr))
	}
	return nil
}
func valuesGetUI4(v unsafe.Pointer, key *PROPERTYKEY) (uint32, error) {
	var val uint32
	hr := comCall(v, 10, uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(&val)))
	if hrFailed(hr) {
		return 0, fmt.Errorf("GetUnsignedIntegerValue: %s", hrString(hr))
	}
	return val, nil
}
func valuesGetUI8(v unsafe.Pointer, key *PROPERTYKEY) (uint64, error) {
	var val uint64
	hr := comCall(v, 14, uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(&val)))
	if hrFailed(hr) {
		return 0, fmt.Errorf("GetUnsignedLargeIntegerValue: %s", hrString(hr))
	}
	return val, nil
}
func valuesGetError(v unsafe.Pointer, key *PROPERTYKEY) (int32, error) {
	var val int32
	hr := comCall(v, 20, uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(&val)))
	if hrFailed(hr) {
		return 0, fmt.Errorf("GetErrorValue: %s", hrString(hr))
	}
	return val, nil
}
func valuesSetString(v unsafe.Pointer, key *PROPERTYKEY, s string) error {
	p, err := strPtr(s)
	if err != nil {
		return err
	}
	hr := comCall(v, 7, uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(p)))
	runtime.KeepAlive(p)
	if hrFailed(hr) {
		return fmt.Errorf("SetStringValue: %s", hrString(hr))
	}
	return nil
}
func valuesGetString(v unsafe.Pointer, key *PROPERTYKEY) (string, error) {
	var p *uint16
	hr := comCall(v, 8, uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(&p)))
	if hrFailed(hr) {
		return "", fmt.Errorf("GetStringValue: %s", hrString(hr))
	}
	s := utf16PtrToString(p)
	coTaskMemFree(unsafe.Pointer(p))
	return s, nil
}
func valuesSetBuffer(v unsafe.Pointer, key *PROPERTYKEY, b []byte) error {
	var p uintptr
	if len(b) > 0 {
		p = uintptr(unsafe.Pointer(&b[0]))
	}
	hr := comCall(v, 29, uintptr(unsafe.Pointer(key)), p, uintptr(uint32(len(b))))
	runtime.KeepAlive(b)
	if hrFailed(hr) {
		return fmt.Errorf("SetBufferValue: %s", hrString(hr))
	}
	return nil
}
func valuesGetBuffer(v unsafe.Pointer, key *PROPERTYKEY) ([]byte, error) {
	var p *byte
	var n uint32
	hr := comCall(v, 30, uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(&p)), uintptr(unsafe.Pointer(&n)))
	if hrFailed(hr) {
		return nil, fmt.Errorf("GetBufferValue: %s", hrString(hr))
	}
	if p == nil || n == 0 {
		coTaskMemFree(unsafe.Pointer(p))
		return []byte{}, nil
	}
	out := append([]byte(nil), unsafe.Slice(p, n)...)
	coTaskMemFree(unsafe.Pointer(p))
	return out, nil
}
func valuesSetPVC(v unsafe.Pointer, key *PROPERTYKEY, coll unsafe.Pointer) error {
	hr := comCall(v, 33, uintptr(unsafe.Pointer(key)), uintptr(coll))
	if hrFailed(hr) {
		return fmt.Errorf("SetPVC: %s", hrString(hr))
	}
	return nil
}
func newValues() (unsafe.Pointer, error) {
	return coCreate(&CLSID_PortableDeviceValues, &IID_IPortableDeviceValues)
}
func newUI4Collection(vals []uint32) (unsafe.Pointer, error) {
	coll, err := coCreate(&CLSID_PortableDevicePropVariantCollection, &IID_IPortableDevicePropVariantCollection)
	if err != nil {
		return nil, err
	}
	for _, v := range vals {
		pv := PROPVARIANT{Vt: VT_UI4, Value: uint64(v)}
		hr := comCall(coll, 5, uintptr(unsafe.Pointer(&pv)))
		if hrFailed(hr) {
			release(coll)
			return nil, fmt.Errorf("Collection.Add: %s", hrString(hr))
		}
	}
	return coll, nil
}
func setupCommandValues(commandID uint32) (unsafe.Pointer, error) {
	v, err := newValues()
	if err != nil {
		return nil, err
	}
	if err = valuesSetGuid(v, &PK_COMMON_COMMAND_CATEGORY, &WPD_MTP_GUID); err != nil {
		release(v)
		return nil, err
	}
	if err = valuesSetUI4(v, &PK_COMMON_COMMAND_ID, commandID); err != nil {
		release(v)
		return nil, err
	}
	// Hard safety invariant: every WPD command is requested with FILE_READ_ACCESS.
	if err = valuesSetUI4(v, &PK_IOCTL_ACCESS, FILE_READ_ACCESS); err != nil {
		release(v)
		return nil, err
	}
	return v, nil
}
func checkCommandHRESULT(result unsafe.Pointer) error {
	e, err := valuesGetError(result, &PK_COMMON_HRESULT)
	if err != nil {
		return err
	}
	if e < 0 {
		return fmt.Errorf("WPD command failed: 0x%08X", uint32(e))
	}
	return nil
}

func managerString(mgr unsafe.Pointer, method int, id string) string {
	idp, err := strPtr(id)
	if err != nil {
		return ""
	}
	var n uint32
	comCall(mgr, method, uintptr(unsafe.Pointer(idp)), 0, uintptr(unsafe.Pointer(&n)))
	if n == 0 {
		return ""
	}
	buf := make([]uint16, n)
	hr := comCall(mgr, method, uintptr(unsafe.Pointer(idp)), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&n)))
	runtime.KeepAlive(idp)
	if hrFailed(hr) {
		return ""
	}
	return syscall.UTF16ToString(buf)
}
func listDevices() ([]DeviceInfo, error) {
	mgr, err := coCreate(&CLSID_PortableDeviceManager, &IID_IPortableDeviceManager)
	if err != nil {
		return nil, err
	}
	defer release(mgr)
	var count uint32
	hr := comCall(mgr, 3, 0, uintptr(unsafe.Pointer(&count)))
	if hrFailed(hr) && count == 0 {
		return nil, fmt.Errorf("GetDevices(count): %s", hrString(hr))
	}
	if count == 0 {
		return nil, nil
	}
	ids := make([]*uint16, count)
	hr = comCall(mgr, 3, uintptr(unsafe.Pointer(&ids[0])), uintptr(unsafe.Pointer(&count)))
	if hrFailed(hr) {
		return nil, fmt.Errorf("GetDevices: %s", hrString(hr))
	}
	out := make([]DeviceInfo, 0, count)
	for _, p := range ids {
		id := utf16PtrToString(p)
		d := DeviceInfo{PnPID: id}
		d.FriendlyName = managerString(mgr, 5, id)
		d.Description = managerString(mgr, 6, id)
		d.Manufacturer = managerString(mgr, 7, id)
		out = append(out, d)
		coTaskMemFree(unsafe.Pointer(p))
	}
	return out, nil
}
func (d *WPDDevice) Close() {
	if d.ptr != nil {
		comCall(d.ptr, 8)
		release(d.ptr)
		d.ptr = nil
	}
}
func openDevice(info DeviceInfo) (*WPDDevice, error) {
	dev, err := coCreate(&CLSID_PortableDeviceFTM, &IID_IPortableDevice)
	if err != nil {
		return nil, err
	}
	client, err := newValues()
	if err != nil {
		release(dev)
		return nil, err
	}
	defer release(client)
	idp, err := strPtr(info.PnPID)
	if err != nil {
		release(dev)
		return nil, err
	}
	hr := comCall(dev, 3, uintptr(unsafe.Pointer(idp)), uintptr(client))
	runtime.KeepAlive(idp)
	if hrFailed(hr) {
		release(dev)
		return nil, fmt.Errorf("Camera open failed (%s). Close LUMIX Tether and try again.", hrString(hr))
	}
	return &WPDDevice{ptr: dev, info: info}, nil
}
func (d *WPDDevice) send(params unsafe.Pointer) (unsafe.Pointer, error) {
	var result unsafe.Pointer
	hr := comCall(d.ptr, 4, 0, uintptr(params), uintptr(unsafe.Pointer(&result)))
	if hrFailed(hr) {
		return nil, fmt.Errorf("SendCommand: %s", hrString(hr))
	}
	if result == nil {
		return nil, errors.New("no WPD result")
	}
	if err := checkCommandHRESULT(result); err != nil {
		release(result)
		return nil, err
	}
	return result, nil
}
func (d *WPDDevice) SendReadPTP(opcode uint32, args []uint32) ([]byte, uint32, error) {
	p, err := setupCommandValues(CMD_EXECUTE_DATA_TO_READ)
	if err != nil {
		return nil, 0, err
	}
	if err = valuesSetUI4(p, &PK_MTP_OPERATION_CODE, opcode); err != nil {
		release(p)
		return nil, 0, err
	}
	coll, err := newUI4Collection(args)
	if err != nil {
		release(p)
		return nil, 0, err
	}
	if err = valuesSetPVC(p, &PK_MTP_OPERATION_PARAMS, coll); err != nil {
		release(coll)
		release(p)
		return nil, 0, err
	}
	release(coll)
	initial, err := d.send(p)
	release(p)
	if err != nil {
		return nil, 0, err
	}
	context, err := valuesGetString(initial, &PK_MTP_CONTEXT)
	if err != nil {
		release(initial)
		return nil, 0, err
	}
	total, err8 := valuesGetUI8(initial, &PK_MTP_TOTAL_SIZE)
	if err8 != nil {
		if v, e := valuesGetUI4(initial, &PK_MTP_TOTAL_SIZE); e == nil {
			total = uint64(v)
		} else {
			release(initial)
			return nil, 0, err8
		}
	}
	release(initial)
	const maxRead = 2 * 1024 * 1024
	const chunkSize = 64 * 1024
	unknown := total == 0xffffffff || total == 0xffffffffffffffff
	target := total
	if unknown || target > maxRead {
		target = maxRead
	}
	data := make([]byte, 0, int(target))
	for uint64(len(data)) < target {
		remain := target - uint64(len(data))
		n := uint32(chunkSize)
		if remain < uint64(n) {
			n = uint32(remain)
		}
		rp, e := setupCommandValues(CMD_READ_DATA)
		if e != nil {
			return nil, 0, e
		}
		if e = valuesSetString(rp, &PK_MTP_CONTEXT, context); e != nil {
			release(rp)
			return nil, 0, e
		}
		if e = valuesSetUI4(rp, &PK_MTP_NUM_TO_READ, n); e != nil {
			release(rp)
			return nil, 0, e
		}
		scratch := make([]byte, n)
		if e = valuesSetBuffer(rp, &PK_MTP_DATA, scratch); e != nil {
			release(rp)
			return nil, 0, e
		}
		rr, e := d.send(rp)
		release(rp)
		if e != nil {
			return nil, 0, e
		}
		chunk, e := valuesGetBuffer(rr, &PK_MTP_DATA)
		if e != nil {
			release(rr)
			return nil, 0, e
		}
		actual, ae := valuesGetUI4(rr, &PK_MTP_NUM_READ)
		release(rr)
		if ae == nil && actual < uint32(len(chunk)) {
			chunk = chunk[:actual]
		}
		data = append(data, chunk...)
		if len(chunk) < int(n) || len(chunk) == 0 {
			break
		}
	}
	ep, err := setupCommandValues(CMD_END_DATA_TRANSFER)
	if err != nil {
		return data, 0, err
	}
	if err = valuesSetString(ep, &PK_MTP_CONTEXT, context); err != nil {
		release(ep)
		return data, 0, err
	}
	er, err := d.send(ep)
	release(ep)
	if err != nil {
		return data, 0, err
	}
	response, rerr := valuesGetUI4(er, &PK_MTP_RESPONSE_CODE)
	release(er)
	if rerr != nil {
		return data, 0, rerr
	}
	return data, response, nil
}

func readPTPString(b []byte, off *int) (string, bool) {
	if *off >= len(b) {
		return "", false
	}
	n := int(b[*off])
	*off = *off + 1
	if n == 0 {
		return "", true
	}
	need := n * 2
	if *off+need > len(b) {
		return "", false
	}
	u := make([]uint16, 0, n-1)
	for i := 0; i < n; i++ {
		v := binary.LittleEndian.Uint16(b[*off+i*2 : *off+i*2+2])
		if v != 0 {
			u = append(u, v)
		}
	}
	*off += need
	return syscall.UTF16ToString(u), true
}
func skipU16Array(b []byte, off *int) bool {
	if *off+4 > len(b) {
		return false
	}
	n := int(binary.LittleEndian.Uint32(b[*off : *off+4]))
	*off += 4
	need := n * 2
	if *off+need > len(b) {
		return false
	}
	*off += need
	return true
}
func parsePTPDeviceInfo(b []byte) (PTPDeviceInfo, error) {
	var x PTPDeviceInfo
	if len(b) < 10 {
		return x, errors.New("short PTP DeviceInfo")
	}
	off := 0
	// StandardVersion(2), VendorExtensionID(4), VendorExtensionVersion(2)
	if off+8 > len(b) {
		return x, errors.New("short header")
	}
	off += 8
	if _, ok := readPTPString(b, &off); !ok {
		return x, errors.New("bad vendor string")
	}
	if off+2 > len(b) {
		return x, errors.New("bad functional mode")
	}
	off += 2
	for i := 0; i < 5; i++ {
		if !skipU16Array(b, &off) {
			return x, errors.New("bad DeviceInfo array")
		}
	}
	var ok bool
	if x.Manufacturer, ok = readPTPString(b, &off); !ok {
		return x, errors.New("bad manufacturer")
	}
	if x.Model, ok = readPTPString(b, &off); !ok {
		return x, errors.New("bad model")
	}
	if x.DeviceVersion, ok = readPTPString(b, &off); !ok {
		return x, errors.New("bad device version")
	}
	if x.Serial, ok = readPTPString(b, &off); !ok {
		return x, errors.New("bad serial")
	}
	return x, nil
}
func parsePanasonicDate(raw uint32) string {
	if raw == 0 {
		return "Not recorded"
	}
	s := fmt.Sprintf("%08X", raw)
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return "Raw " + s
		}
	}
	yy, e1 := strconv.Atoi(s[0:2])
	mo, e2 := strconv.Atoi(s[2:4])
	dd, e3 := strconv.Atoi(s[4:6])
	hh, e4 := strconv.Atoi(s[6:8])
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || mo < 1 || mo > 12 || dd < 1 || dd > 31 || hh > 23 {
		return "Raw " + s
	}
	t := time.Date(2000+yy, time.Month(mo), dd, hh, 0, 0, 0, time.UTC)
	if t.Year() != 2000+yy || int(t.Month()) != mo || t.Day() != dd || t.Hour() != hh {
		return "Raw " + s
	}
	return fmt.Sprintf("20%02d-%02d-%02d %02dh", yy, mo, dd, hh)
}

func decodePanasonicError(code uint32) (string, string, bool) {
	prefix := byte(code >> 24)
	low := uint16(code & 0xFFFF)
	switch prefix {
	case 0x28:
		m := map[uint16]string{
			0x0000: "Flash charge timeout error",
			0x0001: "External flash EEPROM error",
			0x0002: "External flash zoom function error",
			0x0003: "External flash function error",
			0x0010: "BIS HP encoder X low-detect error",
			0x0020: "BIS HP encoder X high-detect error",
			0x0030: "BIS HP encoder Y low-detect error",
			0x0040: "BIS HP encoder Y high-detect error",
			0x0050: "BIS gyro X error",
			0x0060: "BIS gyro Y error",
			0x0070: "BIS gyro communication error",
			0x0080: "BIS gyro R error",
			0x0090: "BIS APU timeout error",
			0x0100: "BIS position sensor X1 error",
			0x0200: "BIS position sensor X2 error",
			0x0300: "BIS position sensor Y error",
			0x0400: "BIS drive voltage X1 error",
			0x0500: "BIS drive voltage X2 error",
			0x0600: "BIS drive voltage Y error",
			0x0700: "BIS differential signal X1 error",
			0x0800: "BIS differential signal X2 error",
			0x0900: "BIS differential signal Y error",
		}
		if d, ok := m[low]; ok {
			return "Flash / BIS", d, true
		}
		return "Flash / BIS", "Unmapped Flash/BIS service code", false
	case 0x2B:
		m := map[uint16]string{
			0x0001: "Flash-ROM / EEPROM data read error",
			0x0002: "Flash-ROM / EEPROM data write error",
			0x0005: "Firmware update error",
			0x0006: "Firmware update error (USB Micon)",
			0x000C: "Lens-FPGA firmware update error",
			0x000D: "Image-FPGA firmware update error",
			0x000E: "TC-FPGA firmware update error",
		}
		if d, ok := m[low]; ok {
			return "Flash-ROM / Firmware", d, true
		}
		return "Flash-ROM / Firmware", "Unmapped Flash-ROM/Firmware service code", false
	case 0x30:
		if low == 0x0000 {
			return "CPU", "NMI reset", true
		}
		return "CPU", "Unmapped CPU reset code", false
	case 0x31:
		m := map[uint16]string{0x0002: "Memory card physical error", 0x0004: "Memory card writing error"}
		if d, ok := m[low]; ok {
			return "Recording Media", d, true
		}
		return "Recording Media", "Unmapped memory-card service code", false
	case 0x38:
		if low == 0x0001 {
			return "CPU / ASIC", "Zoom processing not completed", true
		}
		return "CPU / ASIC", "Unmapped CPU/ASIC halt code", false
	case 0x3A:
		if low == 0x0000 {
			return "Wi-Fi / Bluetooth", "Wi-Fi/Bluetooth module initialization error", true
		}
		return "Wi-Fi / Bluetooth", "Unmapped wireless service code", false
	case 0x3B:
		if low == 0x0000 {
			return "Camera System", "Camera initialization / Flash-ROM initialization error", true
		}
		return "Camera System", "Unmapped camera initialization code", false
	case 0x3C:
		return "Lens", "Lens communication error", true
	case 0x3D:
		if low == 0x0000 {
			return "Camera System", "Assert occurrence", true
		}
		return "Camera System", "Unmapped assert code", false
	case 0x3E:
		m := map[uint16]string{
			0x0001: "Exposure charging operation failure",
			0x0002: "Shutter return-to-home operation failure",
			0x0003: "Mechanical shutter sensor failure",
			0x0004: "Exposure charging operation failure 1",
			0x0005: "Exposure charging operation failure 2",
			0x0006: "Exposure charging recovery operation failure",
			0x0011: "Single-curtain mechanical shutter PR1 operation failure",
			0x0012: "Single-curtain mechanical shutter PR2 operation failure",
			0x0013: "Single-curtain mechanical shutter reset operation failure",
			0x0014: "Abnormal shutter-drive motor current",
			0x0101: "Electronic shutter front-curtain operation failure",
			0x0102: "Electronic shutter front-curtain operation failure",
			0x0112: "Electronic shutter rear-curtain operation failure",
			0x1102: "Mechanical shutter front-curtain set PI1 detection failure",
			0x1103: "Mechanical shutter front-curtain set PI1 detection failure",
			0x1104: "Mechanical shutter front-curtain set PI2 detection failure",
			0x1105: "Mechanical shutter front-curtain set PI2 detection failure",
			0x1106: "Mechanical shutter front-curtain set PI3 detection failure",
			0x1107: "Mechanical shutter front-curtain set PI3 detection failure",
			0x1108: "Mechanical shutter front-curtain set PI4 detection failure",
			0x1109: "Mechanical shutter front-curtain set PI4 detection failure",
			0x1202: "Mechanical shutter exposure-control PI1 detection failure",
			0x1203: "Mechanical shutter exposure-control PI1 detection failure",
			0x1204: "Mechanical shutter exposure-control PI2 detection failure",
			0x1205: "Mechanical shutter exposure-control PI2 detection failure",
			0x1206: "Mechanical shutter exposure-control PI3 detection failure",
			0x1207: "Mechanical shutter exposure-control PI3 detection failure",
			0x1208: "Mechanical shutter exposure-control PI4 detection failure",
			0x1209: "Mechanical shutter exposure-control PI4 detection failure",
			0x1302: "Mechanical shutter release-control 1 PI1 detection failure",
			0x1303: "Mechanical shutter release-control 1 PI1 detection failure",
			0x1304: "Mechanical shutter release-control 1 PI2 detection failure",
			0x1305: "Mechanical shutter release-control 1 PI2 detection failure",
			0x1306: "Mechanical shutter release-control 1 PI3 detection failure",
			0x1307: "Mechanical shutter release-control 1 PI3 detection failure",
			0x1308: "Mechanical shutter release-control 1 PI4 detection failure",
			0x1309: "Mechanical shutter release-control 1 PI4 detection failure",
			0x1402: "Mechanical shutter release-control 2 PI1 detection failure",
			0x1403: "Mechanical shutter release-control 2 PI1 detection failure",
			0x1404: "Mechanical shutter release-control 2 PI2 detection failure",
			0x1405: "Mechanical shutter release-control 2 PI2 detection failure",
			0x1406: "Mechanical shutter release-control 2 PI3 detection failure",
			0x1407: "Mechanical shutter release-control 2 PI3 detection failure",
			0x1408: "Mechanical shutter release-control 2 PI4 detection failure",
			0x1409: "Mechanical shutter release-control 2 PI4 detection failure",
			0x140A: "Mechanical shutter release-control 2 home-position failure",
		}
		if d, ok := m[low]; ok {
			return "Camera / Shutter", d, true
		}
		return "Camera / Shutter", "Unmapped shutter service code", false
	case 0x3F:
		m := map[uint16]string{0x0001: "Movie recording file timeout", 0x0002: "Movie recording file-data cue-send error"}
		if d, ok := m[low]; ok {
			return "Movie Recording", d, true
		}
		return "Movie Recording", "Unmapped movie-recording service code", false
	default:
		return "Unknown", "Service error code not mapped", false
	}
}

func parseUsageRaw(raw []byte) (UsageResult, error) {
	var out UsageResult
	out.RawLen = len(raw)
	frames, err := checkedServiceFrames(raw)
	if err != nil {
		return out, err
	}
	for _, t := range frames {
		if t.Tag == SETUP_INFO_REPLY && len(t.Data) >= 146 {
			d := t.Data
			for i := 0; i < 16; i++ {
				off := i * 8
				dateRaw := binary.LittleEndian.Uint32(d[off : off+4])
				code := binary.LittleEndian.Uint32(d[off+4 : off+8])
				if code == 0 {
					continue
				}
				cat, desc, known := decodePanasonicError(code)
				out.Errors = append(out.Errors, ErrorRecord{
					Index: i + 1, DateRaw: dateRaw, Code: code,
					DateText: parsePanasonicDate(dateRaw), CodeText: fmt.Sprintf("%08X", code),
					Category: cat, Description: desc, Known: known,
				})
			}
			out.PowerWake = binary.LittleEndian.Uint32(d[128:132])
			out.Shutter = binary.LittleEndian.Uint32(d[132:136])
			out.Raw3 = binary.LittleEndian.Uint16(d[136:138])
			out.Raw4 = binary.LittleEndian.Uint16(d[138:140])
			out.Raw5 = binary.LittleEndian.Uint16(d[140:142])
			out.Raw6 = binary.LittleEndian.Uint16(d[142:144])
			out.Raw7 = binary.LittleEndian.Uint16(d[144:146])
			return out, nil
		}
	}
	return out, fmt.Errorf("expected service reply 0x%08X was not found", SETUP_INFO_REPLY)
}
func isLumix(d DeviceInfo) bool {
	h := strings.ToLower(d.Manufacturer + " " + d.FriendlyName + " " + d.Description + " " + d.PnPID)
	return strings.Contains(h, "panasonic") || strings.Contains(h, "lumix") || strings.Contains(h, "vid_04da")
}
func normalizeFirmware(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "Not available"
	}
	if strings.HasPrefix(strings.ToLower(s), "ver") {
		return s
	}
	return "Ver. " + s
}
func probeOneDevice(chosen DeviceInfo) (UsageResult, error) {
	dev, err := openDevice(chosen)
	if err != nil {
		return UsageResult{}, err
	}
	defer dev.Close()
	return probeUsageTransport(chosen, dev)
}

func probeUsageTransport(chosen DeviceInfo, dev l10Transport) (UsageResult, error) {
	result := UsageResult{Timestamp: time.Now()}
	infoConfirmed := false
	if b, rc, e := dev.SendReadPTP(PTP_GET_DEVICE_INFO, nil); e == nil && rc == PTP_RC_OK {
		if di, e2 := parsePTPDeviceInfo(b); e2 == nil {
			infoConfirmed = true
			result.Manufacturer = strings.TrimSpace(di.Manufacturer)
			result.Model = strings.TrimSpace(di.Model)
			result.Firmware = normalizeFirmware(di.DeviceVersion)
			result.Serial = strings.TrimSpace(di.Serial)
		}
	}
	if result.Model == "" {
		result.Model = strings.TrimSpace(chosen.FriendlyName)
		if result.Model == "" {
			result.Model = strings.TrimSpace(chosen.Description)
		}
	}
	if result.Manufacturer == "" {
		result.Manufacturer = strings.TrimSpace(chosen.Manufacturer)
		if result.Manufacturer == "" {
			result.Manufacturer = "Panasonic"
		}
	}
	if result.Firmware == "" {
		result.Firmware = "Not available"
	}
	if result.Serial == "" {
		result.Serial = "Not available"
	}

	if l10Recognized(chosen) && (!infoConfirmed || !strings.EqualFold(result.Model, "DC-L10")) {
		return UsageResult{}, errors.New("L10 standard device information could not be confirmed.")
	}
	raw, rc, err := dev.SendReadPTP(PANASONIC_GET_SETUP_INFO, []uint32{SETUP_INFO_TAG})
	if err != nil {
		return UsageResult{}, fmt.Errorf("Operation History read failed: %v", err)
	}
	if rc != PTP_RC_OK {
		return UsageResult{}, fmt.Errorf("Operation History returned PTP 0x%04X", rc)
	}
	if strings.EqualFold(result.Model, "DC-L10") {
		strict := l10ParseCounters(raw, rc)
		if strict.Status != "PASS" {
			return UsageResult{}, fmt.Errorf("L10 response structure not supported: %s", strict.Reason)
		}
	}
	counts, err := parseUsageRaw(raw)
	if err != nil {
		return UsageResult{}, err
	}

	result.PowerWake = counts.PowerWake
	result.Shutter = counts.Shutter
	result.Raw3 = counts.Raw3
	result.Raw4 = counts.Raw4
	result.Raw5 = counts.Raw5
	result.Raw6 = counts.Raw6
	result.Raw7 = counts.Raw7
	result.Errors = counts.Errors
	result.RawLen = len(raw)

	applyValidationStatus(&result)
	return result, nil
}

func probeCamera() (UsageResult, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hr, _, _ := procCoInitializeEx.Call(0, COINIT_APARTMENTTHREADED)
	coOK := !hrFailed(hr)
	if hrFailed(hr) && uint32(hr) != 0x80010106 {
		return UsageResult{}, fmt.Errorf("COM initialization failed: %s", hrString(hr))
	}
	if coOK {
		defer procCoUninitialize.Call()
	}

	ds, err := listDevices()
	if err != nil {
		return UsageResult{}, err
	}

	var candidates []DeviceInfo
	for i := range ds {
		if isLumix(ds[i]) {
			candidates = append(candidates, ds[i])
		}
	}
	if len(candidates) == 0 {
		return UsageResult{}, errors.New("No LUMIX camera found. Connect USB and select PC(Tether), or LUMIX Lab for DC-L10.")
	}

	var failures []string
	for _, candidate := range candidates {
		res, e := probeOneDevice(candidate)
		if e == nil {
			return res, nil
		}
		name := strings.TrimSpace(candidate.FriendlyName)
		if name == "" {
			name = strings.TrimSpace(candidate.Description)
		}
		if name == "" {
			name = "LUMIX device"
		}
		failures = append(failures, name+": "+e.Error())
	}

	if len(failures) > 0 {
		return UsageResult{}, errors.New("Could not read a compatible LUMIX camera. Close other camera apps and check USB mode (L10: LUMIX Lab). " + strings.Join(failures, " | "))
	}
	return UsageResult{}, errors.New("No readable LUMIX camera found.")
}

func formatNumber(v uint64) string {
	s := fmt.Sprintf("%d", v)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func pointIn(r Rect, x, y int32) bool { return x >= r.L && x < r.R && y >= r.T && y < r.B }

func knownSerial(serial string) bool {
	serial = strings.TrimSpace(serial)
	return serial != "" && !strings.EqualFold(serial, "Not available")
}

func sameCamera(a, b UsageResult) bool {
	if strings.TrimSpace(a.Model) == "" || strings.TrimSpace(b.Model) == "" {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(a.Model), strings.TrimSpace(b.Model)) {
		return false
	}
	as := strings.TrimSpace(a.Serial)
	bs := strings.TrimSpace(b.Serial)
	if !knownSerial(as) || !knownSerial(bs) {
		return false
	}
	return as == bs
}

func serialForMode(serial, mode string) string {
	serial = strings.TrimSpace(serial)
	if !knownSerial(serial) {
		return "Not available"
	}
	last4 := serial
	runes := []rune(serial)
	if len(runes) > 4 {
		last4 = string(runes[len(runes)-4:])
	}
	switch mode {
	case "full":
		return serial
	case "last4":
		return "********" + last4
	default:
		// Fixed-width mask keeps the export layout stable regardless of serial length.
		return "************"
	}
}

func screenSerial(serial string) string {
	stateMu.Lock()
	visible := serialVisible
	stateMu.Unlock()
	if visible {
		return serialForMode(serial, "full")
	}
	return serialForMode(serial, "last4")
}

func serialDisplayFont(text string, full bool) uintptr {
	n := len([]rune(text))
	if full && n > 24 {
		return fontMicro
	}
	if full && n > 15 {
		return fontTiny
	}
	if n > 14 {
		return fontSmall
	}
	return fontValue
}

func validationLabel(r UsageResult) (string, uintptr) {
	if r.SemanticsVerified {
		return "VERIFIED · " + r.Model + " / FW " + strings.TrimPrefix(strings.TrimSpace(r.Firmware), "Ver. "), rgb(167, 200, 240)
	}
	if _, ok := readSuccessDetails(r); ok {
		return "READ REPORTED · COUNTERS UNVERIFIED", rgb(241, 163, 169)
	}
	if r.VerifiedModel {
		return "MODEL VERIFIED · FIRMWARE UNVERIFIED", rgb(241, 163, 169)
	}
	return "UNVERIFIED MODEL / FIRMWARE", rgb(241, 163, 169)
}

func formatDelta(v int64) string {
	if v > 0 {
		return fmt.Sprintf("+%d", v)
	}
	if v < 0 {
		return fmt.Sprintf("%d", v)
	}
	return "+0"
}

func buildSummary(r UsageResult, serialMode string) string {
	status, _ := validationLabel(r)
	var b strings.Builder
	fmt.Fprintf(&b, "%s v%s\r\n", uiText("LUMIX Usage Info", "LUMIX 사용 정보"), appVersion)
	fmt.Fprintf(&b, "%s: %s\r\n", uiText("Camera data revision", "기종 데이터 갱신일"), cameraData.Revision)
	fmt.Fprintf(&b, "%s: %s\r\n", uiText("Model", "모델"), r.Model)
	fmt.Fprintf(&b, "%s: %s\r\n", uiText("Firmware", "펌웨어"), r.Firmware)
	fmt.Fprintf(&b, "%s: %s\r\n", uiText("Serial Number", "시리얼 번호"), serialForMode(r.Serial, serialMode))
	if r.SemanticsVerified {
		fmt.Fprintf(&b, "%s: %s\r\n", uiText("Shutter Actuations", "셔터 작동 횟수"), formatNumber(uint64(r.Shutter)))
		fmt.Fprintf(&b, "%s: %s\r\n", uiText("Power / Wake Activations", "전원 / 깨우기 횟수"), formatNumber(uint64(r.PowerWake)))
	} else {
		fmt.Fprintf(&b, "%s: %s\r\n", counterName(2, r), formatNumber(uint64(r.Shutter)))
		fmt.Fprintf(&b, "%s: %s\r\n", counterName(1, r), formatNumber(uint64(r.PowerWake)))
		fmt.Fprintf(&b, "%s\r\n", uiText("Meanings have not been verified for this model / firmware. Raw fields: shutter = Counter 2, power/wake = Counter 1.", "이 모델/펌웨어에서는 의미가 검증되지 않았습니다. 원본 필드: 셔터=카운터 2, 전원/깨우기=카운터 1."))
	}
	fmt.Fprintf(&b, "%s: %s\r\n", uiText("Validation", "검증 상태"), localizeText(status))
	writeReadSuccess(&b, r, isKoreanUI())
	fmt.Fprintf(&b, "%s\r\n", uiText("Read directly from the connected camera · Read-only", "연결된 카메라에서 읽음 · 읽기 전용"))
	fmt.Fprintf(&b, "%s: %s\r\n", uiText("Generated", "생성 시각"), time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "%s", uiText("Made by @bieup_hieut", "제작 @bieup_hieut"))
	return b.String()
}

func copyTextToClipboard(text string) error {
	if text == "" {
		return nil
	}
	var ok uintptr
	for i := 0; i < 3; i++ {
		ok, _, _ = procOpenClipboard.Call(hwndMain)
		if ok != 0 {
			break
		}
		time.Sleep(35 * time.Millisecond)
	}
	if ok == 0 {
		return errors.New("could not open clipboard")
	}
	defer procCloseClipboard.Call()
	procEmptyClipboard.Call()
	utf := syscall.StringToUTF16(text)
	sz := uintptr(len(utf) * 2)
	h, _, _ := procGlobalAlloc.Call(GMEM_MOVEABLE, sz)
	if h == 0 {
		return errors.New("clipboard allocation failed")
	}
	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		procGlobalFree.Call(h)
		return errors.New("clipboard lock failed")
	}
	if len(utf) > 0 {
		procRtlMoveMemory.Call(ptr, uintptr(unsafe.Pointer(&utf[0])), sz)
	}
	procGlobalUnlock.Call(h)
	set, _, _ := procSetClipboardData.Call(CF_UNICODETEXT, h)
	if set == 0 {
		procGlobalFree.Call(h)
		return errors.New("could not set clipboard data")
	}
	return nil
}

func drawChoiceButton(hdc uintptr, r Rect, label string, selected bool) {
	defer decorateAction(hdc, r)
	bg := actionColor(r, rgb(38, 38, 43))
	fg := actionColor(r, rgb(220, 220, 226))
	if selected {
		bg = actionColor(r, rgb(45, 75, 112))
		fg = actionColor(r, rgb(250, 250, 252))
	}
	fillRound(hdc, r, bg, 10)
	setFont(hdc, fontSmall)
	drawText(hdc, label, r, DT_CENTER|DT_VCENTER|DT_SINGLELINE, fg)
}

func paintExportCard(hdc uintptr, r UsageResult, preset, serialMode string, width, height int32, logo, mascot BitmapInfo) {
	bg, _, _ := procCreateSolidBrush.Call(rgb(18, 18, 20))
	wr := WinRect{0, 0, width, height}
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&wr)), bg)
	procDeleteObject.Call(bg)

	drawBitmap(hdc, logo, 48, 34)
	drawBitmap(hdc, mascot, width-176, 26)
	setFont(hdc, fontSection)
	drawText(hdc, productTitle(), Rect{52, 116, 520, 156}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(245, 245, 247))
	setFont(hdc, fontSmall)
	badge := "PUBLIC SHARE"
	if preset == "verify" {
		badge = "DEVICE VERIFICATION"
	}
	drawText(hdc, badge+" · v"+appVersion, Rect{520, 150, width - 52, 178}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(167, 200, 240))

	setFont(hdc, fontTiny)
	drawText(hdc, uiText("Camera data: ", "기종 데이터: ")+cameraData.Revision, Rect{520, 181, width - 52, 205}, DT_RIGHT|DT_VCENTER|DT_SINGLELINE, rgb(150, 157, 169))
	fillRound(hdc, Rect{48, 214, width - 48, 548}, rgb(28, 28, 31), 18)
	setFont(hdc, fontSection)
	drawText(hdc, "CAMERA INFORMATION", Rect{78, 238, 450, 270}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(240, 240, 244))
	drawHLine(hdc, 78, 279, width-78, rgb(52, 52, 58))

	labels := []string{"Model", "Firmware", "Serial Number"}
	vals := []string{r.Model, r.Firmware, serialForMode(r.Serial, serialMode)}
	y := int32(302)
	for i := range labels {
		setFont(hdc, fontLabel)
		drawText(hdc, labels[i], Rect{80, y, 280, y + 30}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(160, 160, 170))
		if i == 2 {
			setFont(hdc, serialDisplayFont(vals[i], serialMode == "full"))
		} else {
			setFont(hdc, fontValue)
		}
		drawText(hdc, vals[i], Rect{280, y - 2, width - 80, y + 32}, DT_RIGHT|DT_VCENTER|DT_SINGLELINE, rgb(246, 246, 249))
		y += 50
	}

	drawHLine(hdc, 78, 444, width-78, rgb(52, 52, 58))
	setFont(hdc, fontLabel)
	drawText(hdc, counterName(2, r), Rect{80, 462, 580, 494}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(205, 205, 212))
	drawText(hdc, counterName(1, r), Rect{80, 502, 580, 534}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(205, 205, 212))
	setFont(hdc, fontValue)
	drawText(hdc, formatNumber(uint64(r.Shutter)), Rect{620, 458, width - 80, 494}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(250, 250, 252))
	drawText(hdc, formatNumber(uint64(r.PowerWake)), Rect{620, 498, width - 80, 534}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(250, 250, 252))

	setFont(hdc, fontSmall)
	validationText, vc := validationLabel(r)
	drawText(hdc, "● "+localizeText(validationText), Rect{52, height - 108, width - 52, height - 82}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, vc)
	drawText(hdc, "Read directly from the connected camera · Read-only", Rect{52, height - 70, 620, height - 44}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(146, 154, 168))
	drawText(hdc, time.Now().Format("2006-01-02 15:04")+" · @bieup_hieut", Rect{620, height - 70, width - 52, height - 44}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(160, 160, 168))
}

func utf16MultiString(parts ...string) []uint16 {
	var out []uint16
	for _, part := range parts {
		out = append(out, syscall.StringToUTF16(part)...)
	}
	// Windows common-dialog filter lists end with an extra NUL.
	out = append(out, 0)
	return out
}

func savePNGDialog(defaultName string, owner uintptr) (string, bool, error) {
	buf := make([]uint16, 1024)
	nameUTF := syscall.StringToUTF16(defaultName)
	if len(nameUTF) > len(buf) {
		nameUTF = nameUTF[:len(buf)]
	}
	copy(buf, nameUTF)
	filter := utf16MultiString(uiText("PNG Image (*.png)", "PNG 이미지 (*.png)"), "*.png", uiText("All Files (*.*)", "모든 파일 (*.*)"), "*.*")
	defaultDir := preferredReportDir()
	if err := os.MkdirAll(defaultDir, 0755); err != nil {
		return "", false, fmt.Errorf("default save folder unavailable: %w", err)
	}
	initialDir := syscall.StringToUTF16(defaultDir)
	title := syscall.StringToUTF16(uiText("Save LUMIX Usage Info as PNG", "LUMIX 사용 정보 PNG 저장"))
	defExt := syscall.StringToUTF16("png")
	ofn := OPENFILENAMEW{
		LStructSize:     uint32(unsafe.Sizeof(OPENFILENAMEW{})),
		HwndOwner:       owner,
		LpstrFilter:     &filter[0],
		NFilterIndex:    1,
		LpstrFile:       &buf[0],
		NMaxFile:        uint32(len(buf)),
		LpstrInitialDir: &initialDir[0],
		LpstrTitle:      &title[0],
		Flags:           OFN_OVERWRITEPROMPT | OFN_NOCHANGEDIR | OFN_PATHMUSTEXIST,
		LpstrDefExt:     &defExt[0],
	}
	ok, _, _ := procGetSaveFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ok == 0 {
		code, _, _ := procCommDlgExtendedError.Call()
		if code == 0 {
			return "", false, nil // user cancelled
		}
		return "", false, fmt.Errorf("save dialog failed (0x%04X)", uint32(code))
	}
	path := syscall.UTF16ToString(buf)
	if strings.TrimSpace(path) == "" {
		return "", false, errors.New("no save path selected")
	}
	if !strings.EqualFold(filepath.Ext(path), ".png") {
		path += ".png"
	}
	return path, true, nil
}

func defaultExportName(preset string) string {
	kind := "Public"
	if preset == "verify" {
		kind = "Verification"
	}
	return fmt.Sprintf("LUMIX_Usage_Info_%s_%s.png", kind, time.Now().Format("20060102_150405"))
}

func saveExportPNGToPath(r UsageResult, preset, serialMode, path string) error {
	// PNG export runs on its own locked OS thread and owns its GDI resources.
	// Do not share the UI HBITMAP handles here; doing so can stall WM_PAINT if a GDI call blocks.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	const width int32 = 1000
	const height int32 = 650
	logo := loadBitmap(logoAssetPath)
	mascot := loadBitmap(mascotExportPath)
	if logo.Handle != 0 {
		defer procDeleteObject.Call(logo.Handle)
	}
	if mascot.Handle != 0 {
		defer procDeleteObject.Call(mascot.Handle)
	}

	screenDC, _, _ := procGetDC.Call(0)
	if screenDC == 0 {
		return errors.New("could not get display context")
	}
	defer procReleaseDC.Call(0, screenDC)
	memDC, _, _ := procCreateCompatibleDC.Call(screenDC)
	if memDC == 0 {
		return errors.New("could not create export context")
	}
	defer procDeleteDC.Call(memDC)

	// A top-down 32-bit DIB section avoids the fragile GetDIBits path used in beta.5.
	bmi := BITMAPINFO{BmiHeader: BITMAPINFOHEADER{
		BiSize: uint32(unsafe.Sizeof(BITMAPINFOHEADER{})), BiWidth: width, BiHeight: -height,
		BiPlanes: 1, BiBitCount: 32, BiCompression: BI_RGB,
	}}
	var bits unsafe.Pointer
	bmp, _, _ := procCreateDIBSection.Call(screenDC, uintptr(unsafe.Pointer(&bmi)), DIB_RGB_COLORS, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if bmp == 0 || bits == nil {
		return errors.New("could not create PNG render surface")
	}
	defer procDeleteObject.Call(bmp)
	old, _, _ := procSelectObject.Call(memDC, bmp)
	if old == 0 {
		return errors.New("could not select PNG render surface")
	}

	paintExportCard(memDC, r, preset, serialMode, width, height, logo, mascot)
	procSelectObject.Call(memDC, old)

	raw := unsafe.Slice((*byte)(bits), int(width*height*4))
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			si := (y*int(width) + x) * 4
			di := si
			// Windows DIB32 = B,G,R,X; Go RGBA = R,G,B,A.
			img.Pix[di+0] = raw[si+2]
			img.Pix[di+1] = raw[si+1]
			img.Pix[di+2] = raw[si+0]
			img.Pix[di+3] = 255
		}
	}

	return writePNGFile(path, img)
}

func extractVisualAssets() (string, string, string, string, string, string, error) {
	dir := filepath.Join(os.TempDir(), "LUMIX_Usage_Info_"+strings.ReplaceAll(appVersion, ".", "_"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", "", "", "", "", err
	}
	logo := filepath.Join(dir, "lumix_logo.bmp")
	mascot := filepath.Join(dir, "mascot_default.bmp")
	mascotInfo := filepath.Join(dir, "mascot_info.bmp")
	mascotExport := filepath.Join(dir, "mascot_export.bmp")
	mascotError := filepath.Join(dir, "mascot_error.bmp")
	creator := filepath.Join(dir, "creator_character.bmp")
	assets := []struct {
		path string
		data []byte
	}{
		{logo, embeddedLogoBMP},
		{mascot, embeddedMascotDefaultBMP},
		{mascotInfo, embeddedMascotInfoBMP},
		{mascotExport, embeddedMascotExportBMP},
		{mascotError, embeddedMascotErrorBMP},
		{creator, embeddedCreatorCharacterBMP},
	}
	for _, a := range assets {
		if err := os.WriteFile(a.path, a.data, 0644); err != nil {
			return "", "", "", "", "", "", err
		}
	}
	return logo, mascot, mascotInfo, mascotExport, mascotError, creator, nil
}

func loadBitmap(path string) BitmapInfo {
	p := wstr(path)
	h, _, _ := procLoadImageW.Call(0, uintptr(unsafe.Pointer(p)), IMAGE_BITMAP, 0, 0, LR_LOADFROMFILE|LR_CREATEDIBSECTION)
	if h == 0 {
		return BitmapInfo{}
	}
	var bm BITMAP
	procGetObjectW.Call(h, unsafe.Sizeof(bm), uintptr(unsafe.Pointer(&bm)))
	return BitmapInfo{Handle: h, Width: bm.BmWidth, Height: bm.BmHeight}
}

func drawBitmap(hdc uintptr, bmp BitmapInfo, x, y int32) {
	if bmp.Handle == 0 || bmp.Width <= 0 || bmp.Height <= 0 {
		return
	}
	// Artwork uses logical dimensions. The destination viewport transform scales it with monitor DPI.
	// The UI bitmaps are used only by the UI thread. PNG export loads separate bitmap handles.
	mem, _, _ := procCreateCompatibleDC.Call(hdc)
	if mem == 0 {
		return
	}
	old, _, _ := procSelectObject.Call(mem, bmp.Handle)
	oldMode, _, _ := procSetStretchBltMode.Call(hdc, 4)
	var origin struct{ X, Y int32 }
	procSetBrushOrgEx.Call(hdc, 0, 0, uintptr(unsafe.Pointer(&origin)))
	procStretchBlt.Call(hdc, uintptr(x), uintptr(y), uintptr(bmp.Width), uintptr(bmp.Height), mem, 0, 0, uintptr(bmp.Width), uintptr(bmp.Height), SRCCOPY)
	procSetBrushOrgEx.Call(hdc, uintptr(origin.X), uintptr(origin.Y), 0)
	procSetStretchBltMode.Call(hdc, oldMode)
	procSelectObject.Call(mem, old)
	procDeleteDC.Call(mem)
}

func runProbeWorker() int {
	res, err := probeCamera()
	wr := WorkerResponse{}
	if err != nil {
		wr.Error = err.Error()
	} else {
		wr.Result = &res
	}
	if e := json.NewEncoder(os.Stdout).Encode(wr); e != nil {
		return 2
	}
	return 0
}

func probeCameraViaWorker() (UsageResult, error) {
	exePath, err := os.Executable()
	if err != nil {
		return UsageResult{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), workerTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, exePath, "--probe-worker")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: CREATE_NO_WINDOW}
	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return UsageResult{}, errors.New("Camera response timed out. Close other camera apps and check USB mode: PC(Tether), or LUMIX Lab for DC-L10.")
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			return UsageResult{}, fmt.Errorf("camera worker failed: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return UsageResult{}, fmt.Errorf("camera worker failed: %v", err)
	}
	var wr WorkerResponse
	if err := json.Unmarshal(out, &wr); err != nil {
		return UsageResult{}, fmt.Errorf("camera worker returned invalid data: %v", err)
	}
	if wr.Error != "" {
		return UsageResult{}, errors.New(wr.Error)
	}
	if wr.Result == nil {
		return UsageResult{}, errors.New("camera worker returned no result")
	}
	r := *wr.Result
	// Classify using the UI process registry loaded at startup. A data file changed
	// while running must not change worker flags before the user restarts the UI.
	applyValidationStatus(&r)
	return r, nil
}

func createFont(height int32, weight int32) uintptr {
	face := wstr("Malgun Gothic")
	h, _, _ := procCreateFontW.Call(uintptr(height), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(face)))
	return h
}
func setFont(hdc, font uintptr) { procSelectObject.Call(hdc, font) }
func drawText(hdc uintptr, text string, r Rect, flags uint32, color uintptr) {
	text = localizeText(text)
	if flags&DT_SINGLELINE != 0 {
		flags |= DT_VCENTER
	}
	procSetTextColor.Call(hdc, color)
	procSetBkMode.Call(hdc, TRANSPARENT)
	p := wstr(text)
	wr := WinRect{r.L, r.T, r.R, r.B}
	procDrawTextW.Call(hdc, uintptr(unsafe.Pointer(p)), ^uintptr(0), uintptr(unsafe.Pointer(&wr)), uintptr(flags|DT_NOPREFIX))
}
func fillRound(hdc uintptr, r Rect, color uintptr, radius int32) {
	b, _, _ := procCreateSolidBrush.Call(color)
	oldB, _, _ := procSelectObject.Call(hdc, b)
	pen, _, _ := procGetStockObject.Call(NULL_PEN)
	oldP, _, _ := procSelectObject.Call(hdc, pen)
	procRoundRect.Call(hdc, uintptr(r.L), uintptr(r.T), uintptr(r.R), uintptr(r.B), uintptr(radius), uintptr(radius))
	procSelectObject.Call(hdc, oldP)
	procSelectObject.Call(hdc, oldB)
	procDeleteObject.Call(b)
}
func drawButton(hdc uintptr, r Rect, label string) {
	defer decorateAction(hdc, r)
	fillRound(hdc, r, actionColor(r, rgb(245, 245, 245)), 12)
	setFont(hdc, fontButton)
	drawText(hdc, label, Rect{r.L, r.T, r.R, r.B}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, actionColor(r, rgb(22, 22, 24)))
}
func drawOutlineButton(hdc uintptr, r Rect, label string) {
	defer decorateAction(hdc, r)
	fillRound(hdc, r, actionColor(r, rgb(42, 42, 46)), 10)
	setFont(hdc, fontSmall)
	drawText(hdc, label, Rect{r.L, r.T, r.R, r.B}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, actionColor(r, rgb(225, 225, 230)))
}

func drawDashboardCard(hdc uintptr, r Rect, number, title, subtitle string, accent uintptr) {
	defer decorateAction(hdc, r)
	fillRound(hdc, r, actionColor(r, rgb(71, 71, 77)), 18)
	fillRound(hdc, Rect{r.L + 2, r.T + 2, r.R - 2, r.B - 2}, rgb(28, 28, 31), 17)
	fillRound(hdc, Rect{r.L + 15, r.T + 13, r.L + 58, r.B - 13}, rgb(39, 39, 45), 11)
	setFont(hdc, fontSection)
	drawText(hdc, number, Rect{r.L + 18, r.T + 13, r.L + 55, r.B - 13}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, actionColor(r, accent))
	// Keep both text lines inside short action cards and taller welcome cards.
	textTop := r.T + (r.B-r.T-46)/2
	setFont(hdc, fontLabel)
	drawText(hdc, title, Rect{r.L + 71, textTop, r.R - 10, textTop + 24}, DT_LEFT|DT_VCENTER|DT_SINGLELINE, actionColor(r, rgb(243, 243, 247)))
	setFont(hdc, fontSmall)
	drawText(hdc, subtitle, Rect{r.L + 71, textTop + 25, r.R - 10, textTop + 46}, DT_LEFT|DT_VCENTER|DT_SINGLELINE, actionColor(r, rgb(169, 175, 188)))
}

func paintLanguageMenu(hdc uintptr) {
	fillRound(hdc, languageMenuRect, rgb(72, 86, 108), 13)
	fillRound(hdc, Rect{606, 100, 788, 220}, rgb(31, 34, 42), 12)
	mode := atomic.LoadUint32(&languageMode)
	items := []struct {
		rect  Rect
		mode  uint32
		label string
	}{
		{languageAutoRect, languageAuto, uiText("Windows Auto", "Windows 자동")},
		{languageKoreanRect, languageKorean, "한국어"},
		{languageEnglishRect, languageEnglish, "English"},
	}
	for _, item := range items {
		if item.mode == mode {
			fillRound(hdc, item.rect, rgb(45, 69, 100), 8)
		}
		setFont(hdc, fontSmall)
		text := item.label
		if item.mode == mode {
			text = "✓  " + text
		}
		drawText(hdc, text, Rect{item.rect.L + 10, item.rect.T + 3, item.rect.R - 5, item.rect.B - 1}, DT_LEFT|DT_VCENTER|DT_SINGLELINE, rgb(237, 239, 245))
	}
}

func updateWindowLanguage(hwnd uintptr) {
	title := wstr(productTitle() + "  ·  v" + appVersion)
	procSetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(title)))
	procInvalidateRect.Call(hwnd, 0, 0)
}

func drawHLine(hdc uintptr, x1, y, x2 int32, color uintptr) {
	b, _, _ := procCreateSolidBrush.Call(color)
	r := WinRect{x1, y, x2, y + 1}
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), b)
	procDeleteObject.Call(b)
}

func isProjectInfoView(view string) bool {
	return view == "validation" || view == "cameras" || view == "versions" || view == "credits" || l10View(view)
}

func isLiveConnection() bool {
	stateMu.Lock()
	defer stateMu.Unlock()
	return appState == "connected"
}

func selectMascot(st, view, notice string) BitmapInfo {
	if view == "export" {
		if mascotExportBitmap.Handle != 0 {
			return mascotExportBitmap
		}
	}
	if st == "disconnected" || st == "error" {
		if mascotErrorBitmap.Handle != 0 {
			return mascotErrorBitmap
		}
	}
	if isProjectInfoView(view) || view == "guide" || view == "details" || view == "history" || st == "waiting" || st == "connecting" || strings.HasPrefix(notice, "✓") {
		if mascotInfoBitmap.Handle != 0 {
			return mascotInfoBitmap
		}
	}
	return mascotBitmap
}

func paintHeader(hdc uintptr, st, view, notice string) {
	drawBitmap(hdc, logoBitmap, 40, 24)
	// A fixed character slot keeps language controls stable across every page.
	drawBitmap(hdc, selectMascot(st, view, notice), 802, 24)
	setFont(hdc, fontSection)
	drawText(hdc, productTitle(), Rect{48, 110, 505, 157}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(245, 245, 247))
	setFont(hdc, fontSmall)
	drawText(hdc, uiText("Language", "언어"), Rect{530, 54, 594, 90}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(173, 181, 194))
	fillRound(hdc, languageButtonRect, actionColor(languageButtonRect, rgb(42, 48, 60)), 9)
	decorateAction(hdc, languageButtonRect)
	drawText(hdc, languageButtonLabel(), languageButtonRect, DT_VCENTER|DT_CENTER|DT_SINGLELINE, actionColor(languageButtonRect, rgb(235, 238, 246)))
	if languageSaveError != nil {
		setFont(hdc, fontTiny)
		drawText(hdc, uiText("Language changed; preference was not saved.", "언어 변경됨 · 설정 저장 실패"), Rect{505, 136, 790, 163}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(241, 163, 169))
	}
}

func paintWindow(hwnd uintptr) {
	var ps PaintStruct
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	var cr WinRect
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&cr)))
	paintBuffered(hdc, cr.Right-cr.Left, cr.Bottom-cr.Top)
}

func paintScene(hdc uintptr, width, height int32) {
	cr := WinRect{0, 0, width, height}
	bg, _, _ := procCreateSolidBrush.Call(rgb(18, 18, 20))
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&cr)), bg)
	procDeleteObject.Call(bg)
	restore := beginViewport(hdc)
	defer restore()
	defer func() {
		for _, action := range visibleActions() {
			decorateAction(hdc, action.Rect)
		}
		if languageMenuOpen {
			paintLanguageMenu(hdc)
		}
	}()

	stateMu.Lock()
	st := appState
	res := lastResult
	errMsg := lastError
	view := currentView
	notice := uiNotice
	stateMu.Unlock()

	paintHeader(hdc, st, view, notice)

	// Project information remains accessible even without a connected camera.
	if isProjectInfoView(view) {
		switch view {
		case "l10":
			paintL10(hdc)
		case "l10report":
			paintL10Report(hdc)
		case "validation":
			paintValidation(hdc, res)
		case "cameras":
			paintCameraMatrix(hdc, res)
		case "versions":
			paintVersionHistory(hdc, res)
		case "credits":
			paintCredits(hdc, res)
		}
		return
	}

	// When a previously-read camera disconnects, keep last-known data visible.
	if st != "connected" && !(st == "disconnected" && res.Model != "") {
		paintWelcome(hdc, st, errMsg)
		return
	}

	switch view {
	case "details":
		paintDetails(hdc, res)
	case "history":
		paintHistory(hdc, res)
	case "guide":
		paintGuide(hdc, res)
	case "export":
		paintExport(hdc, res)
	default:
		paintUsage(hdc, res)
	}
}

func readOnlyNotice() string {
	return uiText("Read-only — camera settings and data are not changed.", "읽기 전용 — 카메라의 설정과 데이터를 변경하지 않습니다.")
}

func paintWelcome(hdc uintptr, st, errMsg string) {
	fillRound(hdc, Rect{40, 180, 940, 418}, rgb(56, 108, 168), 20)
	fillRound(hdc, Rect{42, 182, 938, 416}, rgb(18, 18, 20), 19)
	fillRound(hdc, Rect{64, 204, 386, 239}, rgb(29, 43, 61), 14)
	setFont(hdc, fontSmall)
	drawText(hdc, "●  USB CONNECTION GUIDE", Rect{82, 204, 369, 239}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(167, 200, 240))
	setFont(hdc, fontTitle)
	drawText(hdc, "Connect your camera", Rect{72, 242, 906, 280}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(250, 247, 244))
	setFont(hdc, fontLabel)
	steps := []string{
		"Turn on the camera.",
		"Connect the camera to this computer with USB.",
		uiText("USB Mode: PC(Tether) · DC-L10: LUMIX Lab.", "USB 모드: PC(테더) · DC-L10: LUMIX Lab을 선택하세요."),
	}
	for i, step := range steps {
		y := int32(284 + i*26)
		drawText(hdc, fmt.Sprintf("%02d", i+1), Rect{74, y, 108, y + 24}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(241, 163, 169))
		drawText(hdc, step, Rect{112, y, 906, y + 24}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(224, 224, 230))
	}
	status := "Waiting for a camera"
	statusColor := rgb(167, 200, 240)
	if st == "connecting" {
		status = "Checking camera connection..."
		statusColor = rgb(167, 200, 240)
	} else if st == "disconnected" {
		status = "Camera disconnected · reconnect USB"
		statusColor = rgb(241, 163, 169)
	} else if st == "error" && errMsg != "" {
		status = errMsg
		statusColor = rgb(241, 163, 169)
	}
	fillRound(hdc, Rect{66, 363, 914, 403}, rgb(35, 35, 39), 13)
	setFont(hdc, fontSmall)
	drawText(hdc, status, Rect{84, 366, 898, 401}, DT_LEFT|0x10|0x2000|0x8000, statusColor)

	drawDashboardCard(hdc, Rect{40, 435, 330, 531}, "01", "CAMERA DATA", "Model, firmware and serial", rgb(167, 200, 240))
	drawDashboardCard(hdc, Rect{345, 435, 635, 531}, "02", "USAGE COUNTERS", "Shutter and power / wake", rgb(167, 200, 240))
	drawDashboardCard(hdc, Rect{650, 435, 940, 531}, "03", "SAVE & SHARE", "PNG and text reports", rgb(241, 163, 169))
	fillRound(hdc, connectRect, actionColor(connectRect, rgb(181, 71, 84)), 12)
	decorateAction(hdc, connectRect)
	setFont(hdc, fontButton)
	drawText(hdc, func() string {
		if st == "connecting" {
			return "CONNECTING..."
		}
		return "CONNECT CAMERA"
	}(), connectRect, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(255, 255, 255))
	drawOutlineButton(hdc, welcomeValidationRect, "VALIDATION & INFO")
	setFont(hdc, fontSmall)
	drawText(hdc, uiText("PC(Tether) is required for listed models. DC-L10 connects using LUMIX Lab.", "PC(테더)가 없는 기종은 지원하지 않습니다. DC-L10은 LUMIX Lab으로 연결합니다."), Rect{40, 628, 940, 653}, DT_VCENTER|DT_CENTER|DT_SINGLELINE, rgb(241, 163, 169))
	drawText(hdc, readOnlyNotice(), Rect{40, 656, 940, 680}, DT_VCENTER|DT_CENTER|DT_SINGLELINE, rgb(167, 200, 240))
	drawText(hdc, uiText("Close other camera apps on this computer before connecting.", "연결 전 컴퓨터에서 실행 중인 다른 카메라 앱을 종료하세요."), Rect{40, 682, 940, 706}, DT_VCENTER|DT_CENTER|DT_SINGLELINE, rgb(146, 154, 168))
	drawFooter(hdc, UsageResult{}, false)
}

func paintUsage(hdc uintptr, r UsageResult) {
	stateMu.Lock()
	st, busy, reportSaving := appState, probeBusy, reportBusy
	visible, notice := serialVisible, uiNotice
	hd, ds, dp := hasDelta, deltaShutter, deltaPowerWake
	stateMu.Unlock()

	// Dark hero keeps the mascot's prepared black background seamless.
	heroAccent := rgb(56, 108, 168)
	if st == "disconnected" || !r.SemanticsVerified {
		heroAccent = rgb(165, 68, 77)
	}
	fillRound(hdc, Rect{40, 180, 940, 343}, heroAccent, 20)
	fillRound(hdc, Rect{42, 182, 938, 341}, rgb(18, 18, 20), 19)
	statusText := "●  CAMERA CONNECTED"
	statusColor := rgb(167, 200, 240)
	if st == "disconnected" {
		statusText = "●  DISCONNECTED · LAST READ"
		statusColor = rgb(241, 163, 169)
	} else if !r.SemanticsVerified {
		statusText = "●  CAMERA READ · COUNTERS UNVERIFIED"
		statusColor = rgb(241, 163, 169)
	}
	statusBg := rgb(29, 43, 61)
	if st == "disconnected" || !r.SemanticsVerified {
		statusBg = rgb(65, 35, 41)
	}
	fillRound(hdc, Rect{64, 199, 435, 234}, statusBg, 13)
	setFont(hdc, fontSmall)
	drawText(hdc, statusText, Rect{80, 199, 424, 234}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, statusColor)
	if !r.Timestamp.IsZero() {
		setFont(hdc, fontTiny)
		drawText(hdc, "Read "+r.Timestamp.Format("15:04:05"), Rect{680, 199, 906, 234}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(150, 157, 169))
	}
	setFont(hdc, fontTiny)
	drawText(hdc, "CAMERA MODEL", Rect{72, 252, 290, 274}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(153, 153, 164))
	drawText(hdc, "FIRMWARE", Rect{340, 252, 490, 274}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(153, 153, 164))
	drawText(hdc, "SERIAL NUMBER", Rect{530, 252, 765, 274}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(153, 153, 164))
	setFont(hdc, fontSection)
	drawText(hdc, r.Model, Rect{72, 279, 327, 315}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(249, 249, 251))
	drawText(hdc, r.Firmware, Rect{340, 279, 517, 315}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(249, 249, 251))
	serialText := screenSerial(r.Serial)
	setFont(hdc, serialDisplayFont(serialText, visible))
	drawText(hdc, serialText, Rect{530, 279, 810, 315}, DT_LEFT|DT_VCENTER|DT_SINGLELINE, rgb(249, 249, 251))
	if visible {
		drawChoiceButton(hdc, serialToggleRect, "HIDE", true)
	} else {
		drawChoiceButton(hdc, serialToggleRect, "SHOW", false)
	}

	setFont(hdc, fontSmall)
	drawText(hdc, readOnlyNotice(), Rect{72, 316, 906, 338}, DT_CENTER|DT_VCENTER|DT_SINGLELINE, rgb(173, 181, 194))

	// Keep both verified meanings and raw labels tied to the existing validation state.
	fillRound(hdc, Rect{40, 358, 480, 479}, rgb(165, 68, 77), 18)
	fillRound(hdc, Rect{42, 360, 478, 477}, rgb(28, 28, 31), 17)
	fillRound(hdc, Rect{500, 358, 940, 479}, rgb(56, 108, 168), 18)
	fillRound(hdc, Rect{502, 360, 938, 477}, rgb(28, 28, 31), 17)
	leftLabel, rightLabel := counterName(2, r), counterName(1, r)
	setFont(hdc, fontSmall)
	drawText(hdc, leftLabel, Rect{66, 378, 400, 400}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(241, 163, 169))
	drawText(hdc, rightLabel, Rect{525, 378, 916, 400}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(167, 200, 240))
	setFont(hdc, fontTitle)
	drawText(hdc, formatNumber(uint64(r.Shutter)), Rect{66, 413, 350, 452}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(250, 249, 251))
	drawText(hdc, formatNumber(uint64(r.PowerWake)), Rect{525, 413, 756, 452}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(250, 249, 251))
	if !r.SemanticsVerified {
		setFont(hdc, fontTiny)
		drawText(hdc, "UNVERIFIED FOR THIS MODEL / FIRMWARE", Rect{66, 453, 452, 473}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(241, 163, 169))
		drawText(hdc, "UNVERIFIED FOR THIS MODEL / FIRMWARE", Rect{525, 453, 916, 473}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(167, 200, 240))
	}
	if hd && r.SemanticsVerified {
		setFont(hdc, fontSmall)
		drawText(hdc, formatDelta(ds), Rect{350, 422, 452, 447}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(167, 200, 240))
		drawText(hdc, formatDelta(dp), Rect{782, 448, 914, 473}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(167, 200, 240))
	}
	drawOutlineButton(hdc, guideRect, "WHAT COUNTS?")

	refreshTitle := "REFRESH CAMERA"
	if busy {
		refreshTitle = "REFRESHING..."
	}
	reportTitle := "SAVE REPORT"
	if reportSaving {
		reportTitle = "SAVING..."
	}
	drawDashboardCard(hdc, refreshRect, "01", refreshTitle, "Read current camera data", rgb(241, 163, 169))
	drawDashboardCard(hdc, historyRect, "04", "ERROR HISTORY", fmt.Sprintf("%d recorded codes", len(r.Errors)), rgb(167, 200, 240))
	drawDashboardCard(hdc, detailsRect, "05", "TECHNICAL DETAILS", "Raw values and status", rgb(167, 200, 240))
	drawDashboardCard(hdc, validationRect, "06", "VALIDATION & INFO", "Supported camera pairs", rgb(167, 200, 240))
	drawDashboardCard(hdc, exportRect, "02", "SAVE PNG", "Privacy options and share", rgb(241, 163, 169))
	drawDashboardCard(hdc, saveRect, "03", reportTitle, "Text report", rgb(241, 163, 169))
	if st == "disconnected" {
		notice = "USB is disconnected. Last successfully read values are shown."
	}
	if busy && notice == "" {
		notice = "Refreshing camera data..."
	}
	if notice != "" {
		setFont(hdc, fontSmall)
		drawText(hdc, notice, Rect{40, 643, 940, 685}, DT_CENTER|0x10|0x2000|0x8000, rgb(164, 189, 209))
	}
	paintSavedActions(hdc, false)
	drawFooter(hdc, r, true)
}

func paintDetails(hdc uintptr, r UsageResult) {
	fillRound(hdc, Rect{40, 180, 940, 655}, rgb(28, 28, 31), 18)
	setFont(hdc, fontSection)
	drawText(hdc, "TECHNICAL DETAILS", Rect{70, 204, 420, 235}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(242, 242, 245))
	setFont(hdc, fontSmall)
	drawText(hdc, "Read-only service data · No camera data modified", Rect{470, 208, 906, 233}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(146, 154, 168))
	drawHLine(hdc, 70, 244, 910, rgb(50, 50, 55))

	type drow struct{ label, value, status string }
	var rows []drow
	values := []uint64{uint64(r.PowerWake), uint64(r.Shutter), uint64(r.Raw3), uint64(r.Raw4), uint64(r.Raw5), uint64(r.Raw6), uint64(r.Raw7)}
	for i, v := range values {
		status := counterName(i+1, r)
		if i < 2 && r.SemanticsVerified {
			status += " · " + uiText("VERIFIED", "검증됨")
		}
		bits := "u16"
		if i < 2 {
			bits = "u32"
		}
		rows = append(rows, drow{fmt.Sprintf("Raw Counter %d (%s)", i+1, bits), formatNumber(v), status})
	}

	y := int32(270)
	for i, row := range rows {
		setFont(hdc, fontSmall)
		drawText(hdc, row.label, Rect{72, y, 270, y + 24}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(183, 183, 191))
		setFont(hdc, fontValue)
		drawText(hdc, row.value, Rect{275, y - 3, 400, y + 27}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(248, 248, 250))
		setFont(hdc, fontSmall)
		c := rgb(150, 150, 159)
		if r.SemanticsVerified && strings.Contains(row.status, "VERIFIED") {
			c = rgb(167, 200, 240)
		}
		drawText(hdc, row.status, Rect{430, y, 906, y + 24}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, c)
		if i < len(rows)-1 {
			drawHLine(hdc, 72, y+31, 906, rgb(44, 44, 49))
		}
		y += 45
	}
	setFont(hdc, fontSmall)
	drawText(hdc, uiText("3/4: legacy candidates STBCNT / PSVCNT · Meanings unverified. 5–7: unidentified.", "3/4: 기존 명칭 STBCNT / PSVCNT 기반 추정 · 의미 미검증. 5~7: 의미 미확인."), Rect{72, 590, 906, 615}, DT_LEFT|DT_VCENTER|DT_SINGLELINE, rgb(166, 177, 195))

	drawText(hdc, fmt.Sprintf("Raw service response: %d bytes · Error records: %d/16", r.RawLen, len(r.Errors)), Rect{72, 620, 906, 644}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(146, 154, 168))
	drawOutlineButton(hdc, backRect, "BACK")
	drawFooter(hdc, r, true)
}

func paintGuide(hdc uintptr, r UsageResult) {
	fillRound(hdc, Rect{40, 180, 940, 655}, rgb(28, 28, 31), 18)
	setFont(hdc, fontSection)
	drawText(hdc, "WHAT COUNTS?", Rect{70, 204, 330, 235}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(242, 242, 245))
	setFont(hdc, fontSmall)
	drawText(hdc, "Observed on DC-S1RM2 · Firmware 1.5", Rect{560, 208, 906, 233}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(146, 154, 168))
	drawHLine(hdc, 70, 244, 910, rgb(50, 50, 55))

	// Left column: shutter behavior
	setFont(hdc, fontSection)
	drawText(hdc, "SHUTTER ACTUATIONS", Rect{72, 266, 455, 295}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(238, 238, 242))
	setFont(hdc, fontSmall)
	left := []struct {
		text  string
		color uintptr
	}{
		{"✓ MECH exposure: +1 / physical actuation", rgb(167, 200, 240)},
		{"✓ Power OFF + shutter CLOSE: +1", rgb(167, 200, 240)},
		{"✓ Electronic shutter exposure: +0", rgb(167, 200, 240)},
		{"✓ High Resolution test: +0 (tested setup)", rgb(167, 200, 240)},
		{"• Sensor Cleaning + restart: +1 total", rgb(190, 190, 198)},
		{"! Pixel Refresh + restart: +3 total*", rgb(241, 163, 169)},
	}
	y := int32(306)
	for _, row := range left {
		drawText(hdc, row.text, Rect{72, y, 475, y + 24}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, row.color)
		y += 38
	}

	// Right column: power/wake behavior
	setFont(hdc, fontSection)
	drawText(hdc, "POWER / WAKE ACTIVATIONS", Rect{500, 266, 906, 295}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(238, 238, 242))
	setFont(hdc, fontSmall)
	right := []struct {
		text  string
		color uintptr
	}{
		{"✓ Camera OFF -> ON: +1", rgb(167, 200, 240)},
		{"✓ Sleep -> wake cycle: +1", rgb(167, 200, 240)},
		{"✓ USB reconnect only: +0", rgb(167, 200, 240)},
		{"• Sensor Cleaning restart: +1", rgb(190, 190, 198)},
		{"• Pixel Refresh restart: +1", rgb(190, 190, 198)},
		{"* Sleep cycle verified; increment timing unknown", rgb(150, 150, 160)},
	}
	y = 306
	for _, row := range right {
		drawText(hdc, row.text, Rect{500, y, 906, y + 24}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, row.color)
		y += 38
	}

	drawHLine(hdc, 72, 548, 906, rgb(50, 50, 55))
	setFont(hdc, fontSmall)
	drawText(hdc, "Pixel Refresh: +3 shutter counts observed with Power-off Shutter=CLOSE.", Rect{72, 566, 906, 590}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(241, 163, 169))
	drawText(hdc, "Exact cause unknown · DC-S1RM2 / 1.5 observations · Not official counter definitions", Rect{72, 596, 906, 620}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(146, 154, 168))
	drawOutlineButton(hdc, backRect, "BACK")
	drawFooter(hdc, r, true)
}

func paintExport(hdc uintptr, r UsageResult) {
	fillRound(hdc, Rect{40, 180, 940, 669}, rgb(28, 28, 31), 18)
	setFont(hdc, fontSection)
	drawText(hdc, "EXPORT / SHARE", Rect{70, 204, 390, 235}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(242, 242, 245))
	setFont(hdc, fontSmall)
	drawText(hdc, "Create a clean PNG or copy a text summary", Rect{520, 208, 906, 233}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(146, 154, 168))
	drawHLine(hdc, 70, 244, 910, rgb(50, 50, 55))

	stateMu.Lock()
	preset := exportPreset
	serialMode := exportSerialMode
	stateMu.Unlock()

	setFont(hdc, fontSection)
	drawText(hdc, "1. CHOOSE PURPOSE", Rect{72, 260, 420, 292}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(238, 238, 242))
	drawChoiceButton(hdc, exportPublicRect, "PUBLIC SHARE", preset == "public")
	drawChoiceButton(hdc, exportVerifyRect, "DEVICE VERIFICATION", preset == "verify")
	setFont(hdc, fontSmall)
	drawText(hdc, "Public Share masks the serial. Device Verification starts with Last 4 digits; Full Serial is optional.", Rect{75, 368, 905, 394}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(150, 150, 160))

	setFont(hdc, fontSection)
	drawText(hdc, "2. SERIAL NUMBER IN EXPORT", Rect{72, 398, 500, 430}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(238, 238, 242))
	drawChoiceButton(hdc, exportMaskedRect, "MASKED", serialMode == "masked")
	drawChoiceButton(hdc, exportLast4Rect, "LAST 4 DIGITS", serialMode == "last4")
	drawChoiceButton(hdc, exportFullRect, "FULL SERIAL", serialMode == "full")
	setFont(hdc, fontSmall)
	serialPreview := serialForMode(r.Serial, serialMode)
	drawText(hdc, "Preview: "+serialPreview, Rect{75, 482, 905, 508}, DT_VCENTER|DT_CENTER|DT_SINGLELINE, rgb(196, 196, 204))
	if serialMode == "full" {
		drawText(hdc, "Full serial number will be included. Avoid posting the exported image publicly.", Rect{75, 514, 905, 540}, DT_VCENTER|DT_CENTER|DT_SINGLELINE, rgb(241, 163, 169))
	}

	stateMu.Lock()
	pb := pngBusy
	notice := uiNotice
	stateMu.Unlock()
	if pb {
		drawButton(hdc, exportSaveRect, uiText("SAVING…", "저장 중…"))
	} else {
		drawButton(hdc, exportSaveRect, "SAVE PNG")
	}
	drawOutlineButton(hdc, exportCopyRect, "COPY SUMMARY")
	paintSavedActions(hdc, true)
	setFont(hdc, fontSmall)
	if notice != "" {
		drawText(hdc, notice, Rect{75, 624, 905, 668}, DT_CENTER|0x10|0x2000|0x8000, rgb(167, 200, 240))
	} else {
		drawText(hdc, uiText("Choose where to save PNG. Default: Local AppData · Lumerian · Reports.", "PNG 저장 위치를 선택하세요. 기본 위치: 로컬 AppData · Lumerian · Reports."), Rect{75, 624, 905, 668}, DT_CENTER|0x10|0x2000|0x8000, rgb(146, 154, 168))
	}
	drawOutlineButton(hdc, backRect, "BACK")
	drawFooter(hdc, r, true)
}

func paintValidation(hdc uintptr, r UsageResult) { paintCameraMatrix(hdc, r) }

func paintVersionHistory(hdc uintptr, r UsageResult) {
	fillRound(hdc, Rect{40, 180, 940, 655}, rgb(28, 28, 31), 18)
	setFont(hdc, fontSection)
	drawText(hdc, "VERSION HISTORY", Rect{70, 204, 520, 235}, DT_VCENTER|DT_SINGLELINE, rgb(242, 242, 245))
	drawHLine(hdc, 70, 244, 910, rgb(50, 50, 55))
	fillRound(hdc, Rect{70, 265, 910, 478}, rgb(34, 34, 38), 12)
	setFont(hdc, fontLabel)
	drawText(hdc, "v"+appVersion+uiText(" · Stable release", " · 정식 버전"), Rect{88, 280, 520, 315}, DT_VCENTER|DT_SINGLELINE, rgb(167, 200, 240))
	setFont(hdc, fontSmall)
	drawText(hdc, uiText("Released ", "릴리즈 날짜 ")+releaseDate, Rect{520, 280, 892, 315}, DT_RIGHT|DT_VCENTER|DT_SINGLELINE, rgb(166, 177, 195))
	drawHLine(hdc, 88, 326, 892, rgb(50, 50, 55))
	lines := []string{
		uiText("Camera information · Shutter / power counts · Error history", "카메라 정보 · 셔터/전원 사용 횟수 · 오류 기록"),
		uiText("Save PNG / TXT · Copy summary · Serial privacy options", "PNG / TXT 저장 · 요약 복사 · 시리얼 공개 범위 선택"),
		uiText("Korean / English · Read-only USB · Lumerian design", "한/영 전환 · 읽기 전용 USB · 루메리안 디자인"),
	}
	for i, line := range lines {
		y := int32(339 + i*40)
		drawText(hdc, line, Rect{88, y, 892, y + 30}, DT_VCENTER|DT_SINGLELINE, rgb(203, 210, 221))
	}
	drawOutlineButton(hdc, backRect, "BACK")
	drawFooter(hdc, r, isLiveConnection())
}

func paintHistory(hdc uintptr, r UsageResult) {
	// Dedicated focused view: the ERROR HISTORY button must never feel like a no-op.
	fillRound(hdc, Rect{38, 178, 942, 657}, rgb(56, 108, 168), 18)
	fillRound(hdc, Rect{42, 182, 938, 653}, rgb(28, 28, 31), 16)
	setFont(hdc, fontSection)
	drawText(hdc, "ERROR HISTORY DETAILS", Rect{70, 204, 420, 235}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(242, 242, 245))
	setFont(hdc, fontSmall)
	stateMu.Lock()
	opened := historyOpenedAt
	stateMu.Unlock()
	if !opened.IsZero() && time.Since(opened) < 4*time.Second {
		drawText(hdc, "● FOCUSED VIEW · Error History opened", Rect{430, 208, 906, 233}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(167, 200, 240))
	} else {
		drawText(hdc, "Up to 16 service-history records · legacy Panasonic reference codes", Rect{390, 208, 906, 233}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE, rgb(146, 154, 168))
	}
	drawHLine(hdc, 70, 244, 910, rgb(50, 50, 55))
	if len(r.Errors) == 0 {
		setFont(hdc, fontLabel)
		drawText(hdc, "No recorded non-zero error codes found.", Rect{72, 300, 906, 332}, DT_VCENTER|DT_CENTER|DT_SINGLELINE, rgb(205, 205, 212))
		setFont(hdc, fontSmall)
		drawText(hdc, "A slot may contain a date marker while the error code is 00000000; this app does not treat that as a recorded error.", Rect{72, 345, 906, 372}, DT_VCENTER|DT_CENTER|DT_SINGLELINE, rgb(146, 154, 168))
	} else {
		setFont(hdc, fontSmall)
		drawText(hdc, "#", Rect{72, 262, 110, 285}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(145, 145, 154))
		drawText(hdc, "DATE / TIME", Rect{115, 262, 300, 285}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(145, 145, 154))
		drawText(hdc, "CODE", Rect{305, 262, 420, 285}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(145, 145, 154))
		drawText(hdc, "DESCRIPTION", Rect{430, 262, 906, 285}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(145, 145, 154))
		drawHLine(hdc, 72, 289, 906, rgb(48, 48, 53))
		perPage := 5
		pages := (len(r.Errors) + perPage - 1) / perPage
		stateMu.Lock()
		pg := historyPage
		stateMu.Unlock()
		if pg >= pages {
			pg = pages - 1
		}
		if pg < 0 {
			pg = 0
		}
		start := pg * perPage
		end := start + perPage
		if end > len(r.Errors) {
			end = len(r.Errors)
		}
		y := int32(310)
		for _, e := range r.Errors[start:end] {
			setFont(hdc, fontSmall)
			drawText(hdc, fmt.Sprintf("%02d", e.Index), Rect{72, y, 110, y + 24}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(215, 215, 220))
			drawText(hdc, e.DateText, Rect{115, y, 300, y + 24}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(205, 205, 212))
			drawText(hdc, e.CodeText, Rect{305, y, 420, y + 24}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(225, 225, 230))
			drawText(hdc, e.Description, Rect{430, y, 906, y + 24}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, rgb(238, 238, 242))
			drawHLine(hdc, 72, y+34, 906, rgb(43, 43, 48))
			y += 58
		}
		setFont(hdc, fontSmall)
		drawText(hdc, fmt.Sprintf("Page %d / %d", pg+1, pages), Rect{370, 620, 610, 644}, DT_VCENTER|DT_CENTER|DT_SINGLELINE, rgb(130, 130, 139))
		if pages > 1 {
			drawOutlineButton(hdc, prevRect, "PREVIOUS")
			drawOutlineButton(hdc, nextRect, "NEXT")
		}
	}
	drawOutlineButton(hdc, backRect, "BACK")
	drawFooter(hdc, r, true)
}

func drawFooter(hdc uintptr, r UsageResult, connected bool) {
	setFont(hdc, fontSmall)
	drawText(hdc, "v"+appVersion, Rect{40, 736, 300, 764}, DT_VCENTER|DT_LEFT|DT_SINGLELINE, actionColor(footerVersionRect, rgb(173, 181, 194)))
	drawText(hdc, "Made by @bieup_hieut · CREDITS", Rect{340, 736, 640, 764}, DT_VCENTER|DT_CENTER|DT_SINGLELINE, actionColor(footerCreditsRect, rgb(173, 181, 194)))
	if connected {
		label := uiText("● VERIFIED · ", "● 검증됨 · ") + r.Model + " / " + uiText("FW ", "펌웨어 ") + firmwareKey(r.Firmware)
		c := rgb(167, 200, 240)
		if r.VerifiedModel && !r.VerifiedFirmware {
			label = "● MODEL VERIFIED · FW UNVERIFIED"
			c = rgb(241, 163, 169)
		} else if !r.VerifiedModel {
			label = "● UNVERIFIED MODEL · " + r.Model
			c = rgb(241, 163, 169)
		}
		if _, ok := readSuccessDetails(r); ok && !r.SemanticsVerified {
			label = uiText("● READ REPORTED · ", "● 읽기 성공 제보 · ") + r.Model
			c = rgb(241, 163, 169)
		}
		drawText(hdc, label, Rect{650, 736, 940, 764}, DT_VCENTER|DT_RIGHT|DT_SINGLELINE|0x8000, c)
	}
}

func startProbe(force bool) {
	stateMu.Lock()
	if l10View(currentView) {
		stateMu.Unlock()
		return
	}
	if probeBusy {
		stateMu.Unlock()
		return
	}
	if !force && appState == "connected" {
		stateMu.Unlock()
		return
	}
	if !force && !lastProbeAt.IsZero() && time.Since(lastProbeAt) < 4*time.Second {
		stateMu.Unlock()
		return
	}
	preserveExisting := lastResult.Model != ""
	probeBusy = true
	lastError = ""
	lastProbeAt = time.Now()
	if preserveExisting {
		uiNotice = "Refreshing camera data..."
	} else {
		appState = "connecting"
		uiNotice = ""
	}
	stateMu.Unlock()
	procInvalidateRect.Call(hwndMain, 0, 0)

	go func(keepExisting bool) {
		res, err := probeCameraViaWorker()
		stateMu.Lock()
		probeBusy = false
		if err != nil {
			if keepExisting {
				appState = "disconnected"
				lastError = err.Error()
				uiNotice = "Camera disconnected or unavailable. Showing last known data."
			} else {
				appState = "error"
				lastError = err.Error()
				if isProjectInfoView(currentView) {
					uiNotice = "Camera not connected. Project information remains available offline."
				} else {
					uiNotice = ""
				}
			}
		} else {
			if keepExisting && sameCamera(lastResult, res) {
				deltaShutter = int64(res.Shutter) - int64(lastResult.Shutter)
				deltaPowerWake = int64(res.PowerWake) - int64(lastResult.PowerWake)
				hasDelta = true
			} else {
				hasDelta = false
				deltaShutter = 0
				deltaPowerWake = 0
			}
			lastResult = res
			appState = "connected"
			lastError = ""
			historyPage = 0
			uiNotice = "Camera read completed."
		}
		stateMu.Unlock()
		procPostMessageW.Call(hwndMain, WM_APP_RESULT, 0, 0)
	}(preserveExisting)
}

func saveReport() {
	stateMu.Lock()
	if reportBusy {
		stateMu.Unlock()
		return
	}
	r := lastResult
	reportBusy = true
	uiNotice = "Saving report..."
	stateMu.Unlock()
	procInvalidateRect.Call(hwndMain, 0, 0)
	go func() {
		dir := preferredReportDir()

		status := "UNVERIFIED MODEL / FIRMWARE"
		validator, verifiedDate := validationDetails(r)
		if r.SemanticsVerified {
			status = "VERIFIED " + r.Model + " / " + r.Firmware
		} else if r.VerifiedModel {
			status = "MODEL VERIFIED / FIRMWARE UNVERIFIED"
		}

		data := buildReport(r, status, validator, verifiedDate)
		path, err := writeUniqueReport(dir, []byte(data), time.Now())
		name := filepath.Base(path)
		stateMu.Lock()
		reportBusy = false
		if err != nil {
			uiNotice = uiText("Could not save report. Check folder access / Windows Security. ", "리포트를 저장하지 못했습니다. 폴더 접근 권한 / Windows 보안을 확인하세요. ") + err.Error()
		} else {
			lastSavedPath = path
			uiNotice = "✓ Report saved: " + name
		}
		stateMu.Unlock()
		procPostMessageW.Call(hwndMain, WM_APP_RESULT, 0, 0)
	}()
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	if handled, result := handleUIMessage(hwnd, msg, wParam, lParam); handled {
		return result
	}
	switch msg {
	case WM_SETTINGCHANGE:
		if refreshUILanguage() {
			updateWindowLanguage(hwnd)
		}
		return 0
	case WM_PAINT:
		paintWindow(hwnd)
		return 0
	case WM_TIMER:
		stateMu.Lock()
		st := appState
		busy := probeBusy
		last := lastProbeAt
		stateMu.Unlock()
		if !busy && st == "waiting" && (last.IsZero() || time.Since(last) >= 4*time.Second) {
			startProbe(false)
		}
		return 0
	case WM_APP_RESULT:
		procInvalidateRect.Call(hwnd, 0, 0)
		return 0
	case WM_DEVICECHANGE:
		evt := uint32(wParam)
		if evt == DBT_DEVICEREMOVECOMPLETE || evt == DBT_DEVNODES_CHANGED {
			stateMu.Lock()
			shouldCheck := lastResult.Model != "" && !probeBusy
			if shouldCheck {
				uiNotice = "USB device change detected. Checking camera connection..."
			}
			stateMu.Unlock()
			if shouldCheck {
				startProbe(true)
			}
		} else if evt == DBT_DEVICEARRIVAL {
			stateMu.Lock()
			st := appState
			busy := probeBusy
			stateMu.Unlock()
			if !busy && (st == "waiting" || st == "error" || st == "disconnected") {
				startProbe(true)
			}
		}
		return 0
	case WM_DESTROY:
		stateMu.Lock()
		cancel := l10State.Cancel
		stateMu.Unlock()
		if cancel != nil {
			cancel()
		}
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func handleClick(hwnd uintptr, x, y int32) uintptr {
	stateMu.Lock()
	beforeView := currentView
	stateMu.Unlock()
	defer settleNavigation(hwnd, beforeView)
	if !allowActivation(x, y) {
		return 0
	}
	if pointIn(languageButtonRect, x, y) {
		languageMenuOpen = !languageMenuOpen
		procInvalidateRect.Call(hwnd, 0, 0)
		return 0
	}
	if languageMenuOpen {
		for _, option := range []struct {
			rect Rect
			mode uint32
		}{
			{languageAutoRect, languageAuto},
			{languageKoreanRect, languageKorean},
			{languageEnglishRect, languageEnglish},
		} {
			if pointIn(option.rect, x, y) {
				languageSaveError = setLanguagePreference(option.mode)
				languageMenuOpen = false
				updateWindowLanguage(hwnd)
				return 0
			}
		}
		languageMenuOpen = false
		procInvalidateRect.Call(hwnd, 0, 0)
		return 0
	}
	stateMu.Lock()
	st := appState
	view := currentView
	res := lastResult
	busyPNG := pngBusy
	stateMu.Unlock()

	if l10View(view) {
		l10Click(x, y)
		return 0
	}
	if openSavedAt(x, y) {
		return 0
	}
	// Global project-info shortcuts are available even with no camera connected.
	if pointIn(footerCreditsRect, x, y) {
		stateMu.Lock()
		currentView = "credits"
		uiNotice = "Developer Credits opened."
		stateMu.Unlock()
		procInvalidateRect.Call(hwndMain, 0, 0)
		return 0
	}
	if pointIn(footerVersionRect, x, y) {
		stateMu.Lock()
		currentView = "versions"
		uiNotice = "Version History opened."
		stateMu.Unlock()
		procInvalidateRect.Call(hwndMain, 0, 0)
		return 0
	}

	hasReadableData := st == "connected" || (st == "disconnected" && res.Model != "")
	if !hasReadableData && !isProjectInfoView(view) {
		if pointIn(connectRect, x, y) {
			startProbe(true)
			return 0
		}
		if pointIn(welcomeValidationRect, x, y) {
			stateMu.Lock()
			currentView = "validation"
			uiNotice = "Validation & Info opened."
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		return 0
	}

	if view == "details" || view == "guide" || view == "validation" {
		returnRect := backRect
		if view == "validation" {
			returnRect = validationBackRect
		}
		if pointIn(returnRect, x, y) {
			stateMu.Lock()
			currentView = "main"
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if view == "validation" {
			if pointIn(cameraPreviousRect, x, y) && cameraPage > 0 {
				cameraPage--
				procInvalidateRect.Call(hwnd, 0, 0)
				return 0
			}
			if pointIn(cameraNextRect, x, y) && cameraPage+1 < cameraListPages() {
				cameraPage++
				procInvalidateRect.Call(hwnd, 0, 0)
				return 0
			}
			if pointIn(validationVersionsRect, x, y) {
				stateMu.Lock()
				currentView = "versions"
				uiNotice = "Version History opened."
				stateMu.Unlock()
				procInvalidateRect.Call(hwndMain, 0, 0)
				return 0
			}
			if pointIn(validationCreditsRect, x, y) {
				stateMu.Lock()
				currentView = "credits"
				uiNotice = "Developer Credits opened."
				stateMu.Unlock()
				procInvalidateRect.Call(hwndMain, 0, 0)
				return 0
			}
		}
	} else if view == "cameras" || view == "versions" || view == "credits" {
		if view == "credits" {
			if pointIn(cameraPreviousRect, x, y) && creditsPage > 0 {
				creditsPage--
				procInvalidateRect.Call(hwnd, 0, 0)
				return 0
			}
			if pointIn(cameraNextRect, x, y) && creditsPage+1 < creditPages() {
				creditsPage++
				procInvalidateRect.Call(hwnd, 0, 0)
				return 0
			}
		}
		if pointIn(backRect, x, y) {
			stateMu.Lock()
			currentView = "validation"
			uiNotice = "Validation & Info opened."
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
	} else if view == "export" {
		if busyPNG {
			stateMu.Lock()
			uiNotice = "PNG operation in progress..."
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pointIn(backRect, x, y) {
			stateMu.Lock()
			currentView = "main"
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pointIn(exportPublicRect, x, y) {
			stateMu.Lock()
			exportPreset = "public"
			exportSerialMode = "masked"
			uiNotice = "Public Share selected. Serial number is masked by default."
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pointIn(exportVerifyRect, x, y) {
			stateMu.Lock()
			exportPreset = "verify"
			// Keep the first transition lightweight and privacy-safe.
			// Full Serial remains available as an explicit choice below.
			exportSerialMode = "last4"
			uiNotice = "Device Verification selected. Choose serial visibility, then SAVE PNG."
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pointIn(exportMaskedRect, x, y) {
			stateMu.Lock()
			exportSerialMode = "masked"
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pointIn(exportLast4Rect, x, y) {
			stateMu.Lock()
			exportSerialMode = "last4"
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pointIn(exportFullRect, x, y) {
			stateMu.Lock()
			exportSerialMode = "full"
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pointIn(exportSaveRect, x, y) {
			stateMu.Lock()
			if pngBusy {
				stateMu.Unlock()
				return 0
			}
			preset, sm := exportPreset, exportSerialMode
			pngBusy = true
			uiNotice = "Choose a save location..."
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)

			// Do not invoke the native common dialog from the window callback.
			// It gets its own locked OS thread, so the main message loop keeps pumping.
			go func(rr UsageResult, pp, ss string) {
				runtime.LockOSThread()
				path, ok, err := savePNGDialog(defaultExportName(pp), hwndMain)
				runtime.UnlockOSThread()
				if err != nil {
					stateMu.Lock()
					pngBusy = false
					uiNotice = "✕ Could not open Save PNG dialog: " + err.Error()
					stateMu.Unlock()
					procPostMessageW.Call(hwndMain, WM_APP_RESULT, 0, 0)
					return
				}
				if !ok {
					stateMu.Lock()
					pngBusy = false
					uiNotice = "PNG save cancelled."
					stateMu.Unlock()
					procPostMessageW.Call(hwndMain, WM_APP_RESULT, 0, 0)
					return
				}

				stateMu.Lock()
				uiNotice = "Saving PNG..."
				stateMu.Unlock()
				procPostMessageW.Call(hwndMain, WM_APP_RESULT, 0, 0)

				err = saveExportPNGToPath(rr, pp, ss, path)
				stateMu.Lock()
				pngBusy = false
				if err != nil {
					uiNotice = uiText("PNG save failed. Choose another folder or check Windows Security. ", "PNG 저장 실패. 다른 폴더를 선택하거나 Windows 보안을 확인하세요. ") + err.Error()
				} else {
					lastSavedPath = path
					uiNotice = "✓ PNG saved: " + filepath.Base(path)
				}
				stateMu.Unlock()
				procPostMessageW.Call(hwndMain, WM_APP_RESULT, 0, 0)
			}(res, preset, sm)
			return 0
		}
		if pointIn(exportCopyRect, x, y) {
			stateMu.Lock()
			sm := exportSerialMode
			uiNotice = "Copying summary..."
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			err := copyTextToClipboard(buildSummary(res, sm))
			stateMu.Lock()
			if err != nil {
				uiNotice = "✕ Could not copy summary: " + err.Error()
			} else {
				uiNotice = "✓ Summary copied to clipboard."
			}
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
	} else if view == "history" {
		if pointIn(backRect, x, y) {
			stateMu.Lock()
			currentView = "main"
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		pages := (len(res.Errors) + 4) / 5
		if pages > 1 && pointIn(prevRect, x, y) {
			stateMu.Lock()
			if historyPage > 0 {
				historyPage--
			}
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pages > 1 && pointIn(nextRect, x, y) {
			stateMu.Lock()
			if historyPage+1 < pages {
				historyPage++
			}
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
	} else {
		if pointIn(serialToggleRect, x, y) {
			stateMu.Lock()
			serialVisible = !serialVisible
			if serialVisible {
				uiNotice = "Serial number shown."
			} else {
				uiNotice = "Serial number masked."
			}
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pointIn(guideRect, x, y) {
			stateMu.Lock()
			currentView = "guide"
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pointIn(refreshRect, x, y) {
			startProbe(true)
			return 0
		}
		if pointIn(historyRect, x, y) {
			stateMu.Lock()
			currentView = "history"
			historyPage = 0
			historyOpenedAt = time.Now()
			uiNotice = "Error History opened."
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pointIn(detailsRect, x, y) {
			stateMu.Lock()
			currentView = "details"
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pointIn(validationRect, x, y) {
			stateMu.Lock()
			currentView = "validation"
			uiNotice = "Validation & Info opened."
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pointIn(exportRect, x, y) {
			if st != "connected" {
				stateMu.Lock()
				uiNotice = "Reconnect and REFRESH before exporting a verification PNG."
				stateMu.Unlock()
				procInvalidateRect.Call(hwndMain, 0, 0)
				return 0
			}
			stateMu.Lock()
			currentView = "export"
			exportPreset = "public"
			exportSerialMode = "masked"
			uiNotice = "Choose export privacy, then click SAVE PNG."
			stateMu.Unlock()
			procInvalidateRect.Call(hwndMain, 0, 0)
			return 0
		}
		if pointIn(saveRect, x, y) {
			if st != "connected" {
				stateMu.Lock()
				uiNotice = "Reconnect and REFRESH before saving a new report."
				stateMu.Unlock()
				procInvalidateRect.Call(hwndMain, 0, 0)
				return 0
			}
			saveReport()
			return 0
		}
	}

	return 0
}

func main() {
	loadCameraData()
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		json.NewEncoder(os.Stdout).Encode(struct {
			App      string `json:"app_version"`
			Data     string `json:"camera_data_revision"`
			Pairs    int    `json:"verified_pairs"`
			Fallback bool   `json:"invalid_external_data"`
		}{appVersion, cameraData.Revision, len(cameraData.Profiles), cameraDataWarning})
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "--probe-worker" {
		os.Exit(runProbeWorker())
	}
	if len(os.Args) > 1 {
		return
	}
	loadLanguagePreference()
	refreshUILanguage()

	// Keep one UI instance only, preventing two processes from competing for the same WPD/PTP camera.
	mutexName := wstr("Local\\LUMIXUsageInfo_UI_Instance")
	hInstanceMutex, _, _ := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(mutexName)))
	if hInstanceMutex != 0 {
		lastErr, _, _ := procGetLastError.Call()
		if uint32(lastErr) == ERROR_ALREADY_EXISTS {
			procMessageBoxW.Call(0, uintptr(unsafe.Pointer(wstr(uiText("LUMIX Usage Info is already running.", "LUMIX 사용 정보가 이미 실행 중입니다.")))), uintptr(unsafe.Pointer(wstr("LUMIX Usage Info"))), MB_OK|MB_ICONINFO)
			procCloseHandle.Call(hInstanceMutex)
			return
		}
		defer procCloseHandle.Call(hInstanceMutex)
	}

	// Win32 window creation/message dispatch must stay on one OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	enableDPIAwareness()
	if logoPath, mascotPath, infoPath, exportPath, errorPath, creatorPath, err := extractVisualAssets(); err == nil {
		logoAssetPath = logoPath
		mascotAssetPath = mascotPath
		mascotInfoPath = infoPath
		mascotExportPath = exportPath
		mascotErrorPath = errorPath
		creatorAssetPath = creatorPath
		logoBitmap = loadBitmap(logoPath)
		mascotBitmap = loadBitmap(mascotPath)
		mascotInfoBitmap = loadBitmap(infoPath)
		mascotExportBitmap = loadBitmap(exportPath)
		mascotErrorBitmap = loadBitmap(errorPath)
		creatorBitmap = loadBitmap(creatorPath)
	}

	hInst, _, _ := procGetModuleHandleW.Call(0)
	cursor, _, _ := procLoadCursorW.Call(0, IDC_ARROW)
	cls := wstr(className)
	wc := WndClassEx{CbSize: uint32(unsafe.Sizeof(WndClassEx{})), Style: CS_HREDRAW | CS_VREDRAW, LpfnWndProc: syscall.NewCallback(wndProc), HInstance: hInst, HCursor: cursor, LpszClassName: cls}
	atom, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		return
	}

	fontTitle = createFont(-28, FW_BOLD)
	fontSubtitle = createFont(-15, FW_NORMAL)
	fontSection = createFont(-19, FW_SEMIBOLD)
	fontLabel = createFont(-16, FW_SEMIBOLD)
	fontValue = createFont(-18, FW_SEMIBOLD)
	fontSmall = createFont(-14, FW_NORMAL)
	fontTiny = createFont(-12, FW_NORMAL)
	fontMicro = createFont(-11, FW_NORMAL)
	fontButton = createFont(-15, FW_SEMIBOLD)

	title := wstr(productTitle() + "  ·  v" + appVersion)
	// Preserve a readable logical canvas; small viewports scroll and monitors scale it.
	hwnd, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(title)), windowStyle, 80, 60, 996, 828, 0, 0, hInst, 0)
	if hwnd == 0 {
		return
	}
	hwndMain = hwnd
	fitInitialWindow(hwnd)
	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)
	procSetTimer.Call(hwnd, 1, 2500, 0)

	// Start one non-blocking auto-detect attempt after the window is visible.
	startProbe(false)

	var m Msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	for _, f := range []uintptr{fontTitle, fontSubtitle, fontSection, fontLabel, fontValue, fontSmall, fontTiny, fontMicro, fontButton} {
		if f != 0 {
			procDeleteObject.Call(f)
		}
	}
	if logoBitmap.Handle != 0 {
		procDeleteObject.Call(logoBitmap.Handle)
	}
	if mascotBitmap.Handle != 0 {
		procDeleteObject.Call(mascotBitmap.Handle)
	}
	if mascotInfoBitmap.Handle != 0 {
		procDeleteObject.Call(mascotInfoBitmap.Handle)
	}
	if mascotExportBitmap.Handle != 0 {
		procDeleteObject.Call(mascotExportBitmap.Handle)
	}
	if mascotErrorBitmap.Handle != 0 {
		procDeleteObject.Call(mascotErrorBitmap.Handle)
	}
	if creatorBitmap.Handle != 0 {
		procDeleteObject.Call(creatorBitmap.Handle)
	}
}
