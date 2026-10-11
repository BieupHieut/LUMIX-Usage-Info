//go:build windows

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var boardEnabled bool
var boardCancel context.CancelFunc
var boardSuiteRuns []boardDiagnostic // Current app session only, newest ten attempts.

// Runtime fields and store are guarded by stateMu; UI-only selection indices run on the message thread.
var boardData boardStore
var boardStoreError error
var boardLatest, boardFailure boardReply
var boardStatus, boardNotice string
var boardSession, boardExtras, boardClosing bool
var boardDevices []DeviceInfo
var boardDeviceIndex, boardPage, boardBodyIndex int

func boardInitialize() {
	boardEnabled = true
	boardExtras = false
	boardData, boardStoreError = boardLoad(boardStorePath())
	if boardStoreError != nil {
		boardNotice = "store_error"
	}
}
func boardStart() { boardStartOptions(false) }
func boardStartOptions(suite bool) {
	stateMu.Lock()
	if probeBusy || boardClosing {
		stateMu.Unlock()
		return
	}
	req := boardPublicRequest(boardSession, suite)
	ctx, cancel := context.WithCancel(context.Background())
	boardCancel = cancel
	if len(boardDevices) > 0 {
		req.Device = boardDevices[boardDeviceIndex%len(boardDevices)].PnPID
	}
	probeBusy = true
	appState = "connecting"
	lastResult = UsageResult{}
	boardLatest = boardReply{}
	boardStoragePage = 0
	boardSerialMode = 0
	boardStatus = "identifying"
	boardNotice = ""
	uiNotice = ""
	hasDelta = false
	stateMu.Unlock()
	procInvalidateRect.Call(hwndMain, 0, 0)
	go func() {
		defer cancel()
		out, _ := boardRunWorkerContext(ctx, req, func(v boardReply) {
			stateMu.Lock()
			boardLatest = v
			boardStatus = v.Kind
			if v.Result != nil {
				lastResult = *v.Result
			}
			stateMu.Unlock()
			procPostMessageW.Call(hwndMain, WM_APP_RESULT, 0, 0)
		})
		stateMu.Lock()
		defer stateMu.Unlock()
		boardLatest = out
		probeBusy = false
		boardCancel = nil
		boardStatus = "failed"
		appState = "error"
		if out.Kind == "select_device" {
			boardDevices = out.Devices
			boardDeviceIndex = 0
			boardStatus = "select"
			currentView = "diagnostics"
		} else {
			boardDevices = nil
			if out.Result != nil {
				lastResult = *out.Result
			}
			if out.Valid {
				appState = "connected"
				boardStatus = "complete"
				if out.Diagnostic.FailedStage != "" {
					boardStatus = "partial"
					for _, step := range out.Diagnostic.Steps {
						if (step.Stage == "close_session" || step.Stage == "worker") && (step.Status == "failed" || step.Status == "unconfirmed") {
							boardStatus = "cleanup"
						}
					}
				}
				if !suite && boardStoreError == nil && boardData.Auto {
					next := boardData
					next.Records = append([]boardRecord(nil), boardData.Records...)
					if e := boardAppend(&next, lastResult); e != nil {
						boardNotice = "serial_missing"
					} else if e = boardPersist(boardStorePath(), next); e != nil {
						boardNotice = "save_failed"
					} else {
						boardData = next
						if boardData.ExternalAuto && boardData.External != "" {
							if e = boardExternalSave(&boardData, boardData.External, false); e != nil {
								boardData.ExternalAuto = false
								boardNotice = "external_changed"
							}
							if e = boardPersist(boardStorePath(), boardData); e != nil {
								boardNotice = "save_failed"
							}
						}
					}
				} else if boardStoreError != nil {
					boardNotice = "store_error"
				}
			}
			if suite {
				boardSuiteRuns = append(boardSuiteRuns, out.Diagnostic)
				if len(boardSuiteRuns) > 10 {
					boardSuiteRuns = boardSuiteRuns[len(boardSuiteRuns)-10:]
				}
				boardStatus = "suite_done"
				if out.Diagnostic.Outcome != "completed" {
					boardStatus = "suite_stopped"
				}
			}
			if out.Diagnostic.FailedStage != "" {
				boardFailure = out
			}
		}
		procPostMessageW.Call(hwndMain, WM_APP_RESULT, 0, 0)
	}()
}
func boardSaveDialog(ext, name string) (string, bool, error) {
	buf := make([]uint16, 32768)
	copy(buf, syscall.StringToUTF16(name))
	filter := utf16MultiString(strings.ToUpper(ext)+" (*."+ext+")", "*."+ext)
	title := syscall.StringToUTF16(boardText("export_history"))
	if ext == "zip" {
		title = syscall.StringToUTF16(boardText("save_diag"))
	}
	def := syscall.StringToUTF16(ext)
	ofn := OPENFILENAMEW{LStructSize: uint32(unsafe.Sizeof(OPENFILENAMEW{})), HwndOwner: hwndMain, LpstrFilter: &filter[0], NFilterIndex: 1, LpstrFile: &buf[0], NMaxFile: uint32(len(buf)), LpstrTitle: &title[0], Flags: OFN_OVERWRITEPROMPT | OFN_NOCHANGEDIR | OFN_PATHMUSTEXIST, LpstrDefExt: &def[0]}
	ok, _, _ := procGetSaveFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ok == 0 {
		code, _, _ := procCommDlgExtendedError.Call()
		if code != 0 {
			return "", false, fmt.Errorf("dialog failed")
		}
		return "", false, nil
	}
	path := syscall.UTF16ToString(buf)
	if !strings.EqualFold(filepath.Ext(path), "."+ext) {
		path += "." + ext
	}
	return path, true, nil
}
func boardSaveDiagnostic() {
	stateMu.Lock()
	d := boardLatest.Diagnostic
	previous := append([]boardDiagnostic(nil), boardSuiteRuns...)
	if d.Started.IsZero() && boardFailure.Diagnostic.FailedStage != "" {
		d = boardFailure.Diagnostic
	}
	stateMu.Unlock()
	path, ok, e := boardSaveDialog("zip", "LUMERIAN_Diagnostic_"+time.Now().Format("20060102_150405")+".zip")
	if !ok && e == nil {
		return
	}
	if e == nil {
		var b []byte
		b, e = boardDiagnosticBundle(d, previous)
		if e == nil {
			e = boardAtomicWrite(path, b)
		}
	}
	stateMu.Lock()
	defer stateMu.Unlock()
	boardNotice = "saved"
	if e != nil {
		boardNotice = "save_failed"
	} else {
		lastSavedPath = path
	}
}
func boardSaveHistory(choose bool) {
	stateMu.Lock()
	path := boardData.External
	stateMu.Unlock()
	if choose {
		var ok bool
		var e error
		path, ok, e = boardSaveDialog("json", "LUMERIAN_History.json")
		if !ok {
			if e != nil {
				stateMu.Lock()
				boardNotice = "save_failed"
				stateMu.Unlock()
			}
			return
		}
	}
	stateMu.Lock()
	defer stateMu.Unlock()
	if boardStoreError != nil {
		boardNotice = "store_error"
		return
	}
	if strings.EqualFold(filepath.Clean(path), filepath.Clean(boardStorePath())) {
		boardNotice = "save_failed"
		return
	}
	if e := boardExternalSave(&boardData, path, choose); e != nil {
		boardNotice = "external_changed"
		return
	}
	if e := boardPersist(boardStorePath(), boardData); e != nil {
		boardNotice = "save_failed"
		return
	}
	boardNotice = "saved"
	lastSavedPath = path
}
func boardToggleSetting(external bool) {
	stateMu.Lock()
	defer stateMu.Unlock()
	if boardStoreError != nil {
		boardNotice = "store_error"
		return
	}
	next := boardData
	if external {
		next.ExternalAuto = !next.ExternalAuto
	} else {
		next.Auto = !next.Auto
	}
	if e := boardPersist(boardStorePath(), next); e != nil {
		boardNotice = "save_failed"
	} else {
		boardData = next
		boardNotice = "saved"
	}
}
func boardOpenIssues() {
	url := wstr("https://github.com/BieupHieut/LUMIX-Usage-Info/issues")
	verb := wstr("open")
	procShellExecute.Call(hwndMain, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(url)), 0, 0, 1)
}
func boardHistoryRows(s boardStore, body int) ([]boardRecord, []string) {
	bodies := []string{""}
	seen := map[string]bool{}
	for _, r := range s.Records {
		if !seen[r.Body] {
			bodies = append(bodies, r.Body)
			seen[r.Body] = true
		}
	}
	if body >= len(bodies) {
		body = 0
	}
	rows := []boardRecord{}
	for i := len(s.Records) - 1; i >= 0; i-- {
		r := s.Records[i]
		if body == 0 || r.Body == bodies[body] {
			rows = append(rows, r)
		}
	}
	return rows, bodies
}
func boardDataFolder() {
	path := filepath.Dir(boardStorePath())
	_ = os.MkdirAll(path, 0700)
	p := wstr(path)
	verb := wstr("open")
	procShellExecute.Call(hwndMain, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(p)), 0, 0, 1)
}
