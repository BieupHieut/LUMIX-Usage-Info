//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

func boardAcquireInstance(name string) (uintptr, bool, error) {
	// LazyProc.Call already captures GetLastError on the calling OS thread.
	// A second Win32 call can clear it or run on a different Go worker thread.
	h, _, lastErr := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(wstr(name))))
	if h == 0 {
		return 0, false, lastErr
	}
	if lastErr == syscall.Errno(ERROR_ALREADY_EXISTS) {
		procCloseHandle.Call(h)
		return 0, true, nil
	}
	return h, false, nil
}

func boardStartupError(stage string, err error) {
	message := boardText("startup_error") + "\n\n" + stage
	if code, ok := err.(syscall.Errno); ok {
		message += fmt.Sprintf(" (0x%08X)", uint32(code))
	}
	procMessageBoxW.Call(0, uintptr(unsafe.Pointer(wstr(message))), uintptr(unsafe.Pointer(wstr(productTitle()))), MB_OK|MB_ICONERROR)
}

func init() {
	boardStrings["patch_startup"] = [4]string{"Reliable single-instance startup and startup error messages", "중복 실행 방지 수정과 시작 오류 안내", "二重起動の防止を修正・起動エラーを案内", "Inicio único fiable y mensajes de error al iniciar"}
	boardStrings["patch_history"] = [4]string{"History cleanup preserves daily boundaries and observed increases", "이력 정리 시 날짜 경계와 관측 증가량 보존", "履歴整理時に日付境界と観測増加量を保持", "La limpieza conserva los límites diarios y los aumentos observados"}
	boardStrings["patch_privacy"] = [4]string{"Opening Save report resets serial visibility to hidden", "리포트 저장 화면을 열 때 시리얼 숨김으로 시작", "レポート保存画面を開くと製造番号を非表示に", "Al abrir Guardar informe, la serie vuelve a ocultarse"}
	boardStrings["patch_errors"] = [4]string{"Camera error history opens on its first page", "카메라 오류 기록을 첫 페이지부터 표시", "カメラのエラー履歴を最初のページから表示", "El historial de errores se abre en la primera página"}
	boardStrings["startup_error"] = [4]string{
		"The app could not start. Try again; if this repeats, report the stage and code below.",
		"앱을 시작하지 못했습니다. 다시 실행해 보시고, 반복되면 아래 단계와 코드를 알려 주세요.",
		"アプリを起動できませんでした。再実行し、続く場合は以下の段階とコードをお知らせください。",
		"No se pudo iniciar la aplicación. Inténtelo de nuevo; si se repite, comunique la etapa y el código indicados.",
	}
	boardStrings["already_running"] = [4]string{
		"LUMERIAN LUMIX Usage Board is already running. Use the open window.",
		"LUMERIAN LUMIX Usage Board가 이미 실행 중입니다. 열려 있는 창을 이용하세요.",
		"LUMERIAN LUMIX Usage Boardは起動済みです。開いているウィンドウをご利用ください。",
		"LUMERIAN LUMIX Usage Board ya está en ejecución. Use la ventana abierta.",
	}
}
