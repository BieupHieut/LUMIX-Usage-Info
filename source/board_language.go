//go:build windows

package main

import (
	"strings"
	"sync/atomic"
)

var boardWindowsLanguage uint32 = languageEnglish

func boardTranslate(text string) (string, bool) {
	m := boardLanguage()
	if m != languageJapanese && m != languageSpanish {
		return "", false
	}
	if strings.HasPrefix(text, "Preview: ") {
		label := "プレビュー："
		if m == languageSpanish {
			label = "Vista previa: "
		}
		return label + strings.TrimPrefix(text, "Preview: "), true
	}
	if strings.HasPrefix(text, "Raw Counter ") {
		label := "生カウンター "
		if m == languageSpanish {
			label = "Contador bruto "
		}
		return label + strings.TrimPrefix(text, "Raw Counter "), true
	}
	for key, row := range boardStrings {
		if text == row[0] {
			return boardText(key), true
		}
	}
	if row, ok := boardLegacyStrings[text]; ok {
		if m == languageJapanese {
			return row[0], true
		}
		return row[1], true
	}
	return "", false
}

var boardLegacyStrings = map[string][2]string{
	"BACK": {"戻る", "Volver"}, "VERSION HISTORY": {"バージョン履歴", "Historial de versiones"},
	"CAMERA LIST & VALIDATION": {"カメラ一覧・検証", "Cámaras y verificación"},
	"Developer Credits":        {"制作者・協力者", "Créditos"}, "CREDITS": {"クレジット", "Créditos"},
	"Camera model": {"カメラモデル", "Modelo"}, "Firmware": {"ファームウェア", "Firmware"}, "Serial Number": {"製造番号", "Número de serie"},
	"VERIFIED": {"検証済み", "Verificado"}, "UNVERIFIED MODEL": {"未検証モデル", "Modelo sin verificar"},
	"SHOW": {"表示", "Mostrar"}, "HIDE": {"非表示", "Ocultar"}, "TECHNICAL DETAILS": {"技術情報", "Detalles técnicos"},
	"ERROR HISTORY DETAILS":                   {"エラー履歴の詳細", "Detalles del historial de errores"},
	"No recorded non-zero error codes found.": {"エラーコードの記録はありません。", "No se encontraron códigos de error."},
	"DATE / TIME":                             {"日時", "Fecha / hora"}, "CODE": {"コード", "Código"}, "DESCRIPTION": {"説明", "Descripción"},
	"Windows Auto": {"Windowsに従う", "Automático (Windows)"},
	"PUBLIC SHARE": {"公開用", "Compartir"}, "DEVICE VALIDATION": {"機種検証用", "Verificación"},
	"SAVE PNG": {"PNGを保存", "Guardar PNG"}, "COPY SUMMARY": {"概要をコピー", "Copiar resumen"},
	"MASKED": {"非表示", "Oculto"}, "LAST 4 DIGITS": {"末尾4桁", "Últimos 4 dígitos"}, "FULL SERIAL": {"全桁表示", "Serie completa"},
	"Open saved file": {"保存ファイルを開く", "Abrir archivo"}, "Open folder": {"フォルダーを開く", "Abrir carpeta"},
}

func boardLanguage() uint32 {
	m := atomic.LoadUint32(&languageMode)
	if m != languageAuto {
		return m
	}
	if atomic.LoadUint32(&windowsKoreanUI) != 0 {
		return languageKorean
	}
	return atomic.LoadUint32(&boardWindowsLanguage)
}
func boardText(key string) string {
	row, ok := boardStrings[key]
	if !ok {
		return key
	}
	i := 0
	switch boardLanguage() {
	case languageKorean:
		i = 1
	case languageJapanese:
		i = 2
	case languageSpanish:
		i = 3
	}
	return row[i]
}

var boardStrings = map[string][4]string{
	"stable_release":     {"Stable release", "정식 버전", "正式版", "Versión estable"},
	"diagnostic_ready":   {"Read the camera, then review the stages or save a diagnostic ZIP if needed.", "카메라를 조회한 뒤 단계별 결과를 확인하거나 필요한 경우 진단 ZIP을 저장하세요.", "カメラを取得後、各段階の結果を確認し、必要に応じて診断ZIPを保存します。", "Lea la cámara y revise las etapas o guarde un ZIP de diagnóstico si hace falta."},
	"release_history":    {"Automatic history for multiple camera bodies and JSON file updates", "여러 카메라의 사용 이력 자동 저장 및 JSON 파일 갱신", "複数カメラの履歴を自動保存・JSONファイルを更新", "Historial automático por cámara y actualización de archivos JSON"},
	"release_analysis":   {"Daily / monthly / yearly charts and selected-record deletion", "일별·월별·연별 분석 및 선택 기록 삭제", "日・月・年別の分析と選択した記録の削除", "Gráficos diarios, mensuales y anuales; eliminación de registros seleccionados"},
	"release_diagnostic": {"Connection-stage results, retry, cancellation and diagnostic ZIP", "연결 단계별 결과·재조회·중단 및 진단 ZIP 저장", "接続段階の結果・再取得・中止・診断ZIPの保存", "Etapas de conexión, reintento, cancelación y ZIP de diagnóstico"},
	"release_privacy":    {"Serial visibility, private exports and a consistent interface", "시리얼 표시 전환·내보내기 공개 범위 및 화면 구성 통일", "製造番号の表示切替・公開範囲・統一した画面構成", "Visibilidad de serie, privacidad al exportar e interfaz uniforme"},
	"welcome_title":      {"Get to know your camera", "내 카메라의 사용 기록을 확인하세요", "カメラの使用記録を確認", "Conozca el uso de su cámara"},
	"welcome_subtitle":   {"Camera information, shutter count and power/wake count", "카메라 정보부터 셔터·전원 작동 횟수까지", "カメラ情報・シャッター回数・電源と復帰の回数", "Información de cámara, obturador y encendidos"},
	"welcome_usb":        {"Connect the powered-on camera by USB and select PC(Tether).", "카메라 전원을 켜고 USB로 연결한 뒤 PC(테더)를 선택하세요.", "カメラの電源を入れてUSB接続し、PC(Tether)を選択します。", "Encienda la cámara, conecte por USB y seleccione PC(Tether)."},
	"welcome_l10":        {"DC-L10: select LUMIX Lab. See camera compatibility for supported models.", "DC-L10은 LUMIX Lab을 선택하세요. 지원 기종은 검증 목록에서 확인할 수 있습니다.", "DC-L10はLUMIX Labを選択。対応機種は検証リストをご覧ください。", "DC-L10: seleccione LUMIX Lab. Consulte la lista de compatibilidad."},
	"welcome_connect":    {"Connect camera", "카메라 연결하기", "カメラを接続", "Conectar cámara"},
	"overview":           {"Your camera at a glance", "내 카메라 사용 정보", "カメラの使用情報", "Uso de su cámara"},
	"my_camera":          {"My camera", "내 카메라", "マイカメラ", "Mi cámara"},
	"recorded_total":     {"Cumulative count recorded by the camera", "카메라에 기록된 누적 횟수", "カメラに記録された累計回数", "Total acumulado registrado por la cámara"},
	"refresh_read":       {"Read again", "다시 조회", "再取得", "Volver a leer"},
	"card_trial":         {"Card capacity · final test", "카드 용량 · 최종 시험", "カード容量・最終試験", "Capacidad de tarjeta · prueba final"},

	"serial_change":      {"Serial · click to change", "시리얼 · 클릭하여 변경", "シリアル · クリックで切替", "Serie · clic para cambiar"},
	"serial_label":       {"Serial number", "시리얼 번호", "シリアル番号", "Número de serie"},
	"snapshot_note":      {"Snapshot at query time · Use Read camera to refresh.", "조회 시점의 상태입니다. 카메라 조회를 눌러 갱신하세요.", "取得時点の状態です。再取得で更新してください。", "Datos del momento de consulta. Vuelva a leer para actualizar."},
	"camera_status_note": {"Card information is experimental. Missing capacity does not erase usage counts.", "카드 정보는 시험 기능입니다. 용량이 조회되지 않아도 사용 횟수는 유지됩니다.", "カード情報は試験機能です。容量が取得できなくても使用回数は保持します。", "La información de tarjetas es experimental. La falta de capacidad no borra los contadores."},
	"storage_unexposed":  {"Capacity unavailable over this connection.\nCard status is shown separately.", "현재 연결에서는 용량을 조회할 수 없습니다.\n카드 상태와 용량 조회 여부는 별개입니다.", "この接続では容量を取得できません。\nカード状態は別に表示します。", "Capacidad no disponible en esta conexión.\nEl estado de la tarjeta se muestra aparte."},
	"storage_skipped":    {"Capacity query skipped: no addressable logical storage", "용량 조회 생략: 조회 가능한 논리 저장소 없음", "容量取得を省略：アクセス可能な論理ストレージなし", "Consulta omitida: sin almacenamiento lógico accesible"},
	"space_unavailable":  {"Capacity not available", "용량 조회되지 않음", "容量を取得できません", "Capacidad no disponible"},
	"no_storage":         {"No storage exposed over USB.", "USB로 제공된 저장소가 없습니다.", "USBで公開されたストレージはありません。", "No se expone almacenamiento por USB."},
	"no_cards_reported":  {"Camera reports no cards.", "카메라가 카드 없음으로 응답했습니다.", "カメラはカードなしと応答しました。", "La cámara indica que no hay tarjetas."},
	"card_error":         {"Error", "오류", "エラー", "Error"},
	"card_protected":     {"Write protected", "쓰기 보호", "書込保護", "Protegida"},
	"card_ready":         {"Available", "사용 가능", "使用可能", "Disponible"},
	"card_absent":        {"No card", "없음", "未挿入", "Sin tarjeta"},
	"state_unknown":      {"Unknown state", "알 수 없는 상태", "不明な状態", "Estado desconocido"},
	"state_stream_rec":   {"Streaming / remote recording", "스트리밍 · 원격 촬영 중", "配信・リモート撮影中", "Transmisión / grabación remota"},
	"state_remote":       {"Remote recording", "원격 촬영 중", "リモート撮影中", "Grabación remota"},
	"state_stream":       {"Live streaming", "라이브 스트리밍 중", "配信中", "Transmisión en directo"},
	"state_menu":         {"Menu displayed", "메뉴 표시 중", "メニュー表示中", "Menú abierto"},
	"state_nr":           {"Noise reduction", "노이즈 감소 처리 중", "ノイズ低減処理中", "Reducción de ruido"},
	"state_processing":   {"Processing", "처리 중", "処理中", "Procesando"},
	"state_playback":     {"Playback", "재생 중", "再生中", "Reproducción"},
	"state_movie":        {"Recording video", "동영상 촬영 중", "動画撮影中", "Grabando vídeo"},
	"state_standby":      {"Standby", "촬영 대기", "待機", "En espera"},
	"raw_short":          {"Raw / unverified", "원시 값 / 미검증", "生データ・未検証", "Datos brutos / sin verificar"},
	"slot_note":          {"Storage IDs may differ from card slots.", "저장소 ID는 카드 슬롯 번호와 다를 수 있습니다.", "ストレージIDとスロット番号は異なる場合があります。", "Los ID no siempre coinciden con las ranuras."},
	"free":               {"Free", "여유", "空き", "Libre"},
	"capacity":           {"Total", "전체", "合計", "Total"},
	"camera_store":       {"Storage", "저장소", "ストレージ", "Almacenamiento"},
	"cards_space":        {"Cards & storage", "메모리 카드 · 저장공간", "カード・空き容量", "Tarjetas y capacidad"},
	"remaining_movie":    {"Movie time remaining", "동영상 촬영 가능 시간", "動画撮影可能時間", "Tiempo de vídeo restante"},
	"remaining_photos":   {"Photos remaining", "남은 사진 매수", "撮影可能枚数", "Fotos restantes"},
	"shooting":           {"Shooting status", "촬영 상태", "撮影状態", "Estado de captura"},
	"snapshot":           {"Last read", "마지막 조회", "最終取得", "Última lectura"},
	"camera_ready":       {"Camera connection", "카메라 연결", "カメラ接続", "Conexión de cámara"},
	"suite_retry":        {"Run S1 diagnostic again", "종합 진단 재시도", "S1診断を再実行", "Repetir diagnóstico S1"},
	"suite":              {"S1 full diagnostic", "S1 종합 진단", "S1 総合診断", "Diagnóstico S1"},
	"stop":               {"Stop diagnostic", "검사 중단", "診断を中止", "Detener diagnóstico"},
	"suite_running":      {"Diagnostic running — keep USB connected.", "종합 진단 중 · USB 연결을 유지하세요.", "診断中・USB接続を維持してください。", "Diagnóstico en curso; mantenga el USB conectado."},
	"suite_done":         {"Diagnostic finished. Save the complete ZIP below.", "종합 진단 완료 · 아래에서 전체 결과 ZIP을 저장하세요.", "診断終了・下のボタンで結果ZIPを保存してください。", "Diagnóstico finalizado. Guarde el ZIP completo."},
	"suite_stopped":      {"Diagnostic stopped. Save ZIP; reconnect USB before retrying.", "진단 중단 · ZIP 저장 후 다음 검사 전 USB를 다시 연결하세요.", "診断中止・ZIP保存後、再試行前にUSBを再接続してください。", "Diagnóstico detenido. Guarde el ZIP y reconecte el USB."},
	"suite_steps":        {"Recorded steps", "기록된 단계", "記録した段階", "Pasos registrados"},
	"suite_note":         {"S1 / 2.4: baseline → session → conditional follow-up.\nNo history saved. If stopped, reconnect USB before another run.", "S1 / 2.4: 기존 방식 → 세션 → 조건부 재조회\n검사값은 이력에 저장하지 않습니다. 중단 후에는 USB를 다시 연결하세요.", "S1 / 2.4：通常→セッション→条件付き再取得。\n履歴保存なし。中止した場合は、次の診断前にUSBを再接続してください。", "S1 / 2.4: normal → sesión → consulta condicional.\nSin historial. Si se detiene, reconecte el USB antes de repetir."},
	"history_legend":     {"✓ Verified pair · ? Estimated / unverified · Compared with the previous record for the same body", "✓ 검증된 조합 · ? 추정/미검증 · 같은 바디의 직전 기록과 비교", "✓ 検証済みの組合せ・? 推定/未検証・同じカメラの前回記録と比較", "✓ Verificado · ? Estimado / sin verificar · Comparación con el registro anterior de la misma cámara"},
	"counters_verified":  {"Shutter / power: verified", "셔터·전원 횟수 검증됨", "シャッター・電源：検証済み", "Obturador / encendido: verificados"},
	"counters_estimated": {"Shutter / power: estimated", "셔터·전원 횟수 추정됨", "シャッター・電源：推定", "Obturador / encendido: estimados"},
	"shutter_short":      {"Shutter", "셔터", "シャッター", "Obturador"},
	"power_short":        {"Power/wake", "전원/깨우기", "電源・復帰", "Encendido"},
	"privacy_short":      {"Full serial and device path excluded · Save, then send the ZIP yourself.", "전체 시리얼·장치 경로 제외 · ZIP 저장 후 직접 전달합니다.", "全桁の製造番号・パスを除外。保存したZIPをご自身で送信してください。", "Sin serie completa ni ruta del dispositivo · Guarde y envíe el ZIP manualmente."},
	"reading_extra":      {"Reading camera status and storage…", "카메라 상태·저장공간 조회 중…", "カメラ状態・容量を取得中…", "Leyendo estado y almacenamiento…"},
	"language_unsaved":   {"Language changed; preference could not be saved.", "언어를 변경했으나 설정 저장에 실패했습니다.", "言語は変更されましたが設定を保存できませんでした。", "Idioma cambiado; no se pudo guardar la preferencia."},
	"partial":            {"Counters read; optional information unavailable.", "카운터 조회 성공 · 일부 추가 정보 조회 실패", "カウンター取得成功・追加情報の一部は取得不可", "Contadores leídos; faltan algunos datos opcionales."},
	"save_error":         {"Save last failure ZIP", "최근 오류 ZIP 저장", "直近のエラーZIPを保存", "Guardar ZIP del último error"},
	"text_report":        {"Save text report", "텍스트 리포트 저장", "テキスト保存", "Guardar texto"},
	"verified":           {"Verified", "검증됨", "検証済み", "Verificado"},
	"estimated":          {"Estimated / unverified", "추정됨 / 미검증", "推定・未検証", "Estimado / sin verificar"},
	"history_note":       {"Select a camera to compare its consecutive readings. Firmware changes start a new comparison.", "카메라를 선택해 연속 기록을 비교합니다. 펌웨어가 바뀌면 새 비교를 시작합니다.", "カメラを選択して連続する記録を比較します。FW変更時は比較をリセットします。", "Seleccione una cámara para comparar lecturas consecutivas. Un cambio de firmware inicia otra referencia."},
	"camera":             {"Camera", "카메라", "カメラ", "Cámara"},
	"history":            {"Usage history", "사용 이력", "使用履歴", "Historial de uso"},
	"diagnostic":         {"Connection & diagnostics", "연결 및 진단", "接続・診断", "Conexión y diagnóstico"},
	"settings":           {"History settings", "이력 저장 설정", "履歴の保存設定", "Ajustes del historial"},
	"connect":            {"Read camera", "카메라 조회", "カメラを読み取る", "Leer cámara"},
	"waiting":            {"Connect USB, then select PC(Tether). DC-L10: select LUMIX Lab.", "USB 연결 후 PC(테더)를 선택하세요. DC-L10: LUMIX Lab을 선택하세요.", "USB接続後、PC(Tether)を選択。DC-L10はLUMIX Labを選択します。", "Conecte por USB y seleccione PC(Tether). DC-L10: LUMIX Lab."},
	"close_apps":         {"Close other camera apps before reading.", "조회 전 다른 카메라 앱을 종료하세요.", "読み取り前に他のカメラアプリを終了してください。", "Cierre otras aplicaciones de cámara antes de leer."},
	"readonly":           {"Read-only · Unofficial tool", "읽기 전용 · 비공식 도구", "読み取り専用・非公式ツール", "Solo lectura · No oficial"},
	"safe":               {"Camera settings and stored data are not modified.", "카메라 설정과 저장된 데이터를 변경하지 않습니다.", "カメラの設定や保存データは変更しません。", "No se modifican los ajustes ni los datos de la cámara."},
	"about":              {"App information", "앱 정보", "アプリ情報", "Acerca de"},
	"shutter":            {"Shutter actuations", "셔터 작동 횟수", "シャッター作動回数", "Accionamientos del obturador"},
	"power":              {"Power / wake activations", "전원 / 깨우기 횟수", "電源・復帰回数", "Encendidos / activaciones"},
	"usb":                {"USB link information", "USB 연결 정보", "USB接続情報", "Información del enlace USB"},
	"raw":                {"Raw · interpretation unverified", "원시 값 · 의미 미검증", "生データ・意味未検証", "Valor bruto · sin verificar"},
	"unavailable":        {"Not available", "조회되지 않음", "取得できません", "No disponible"},
	"extras":             {"Read card information (test)", "카드 정보 조회 (시험)", "カード情報取得（試験）", "Leer información de tarjetas (prueba)"},
	"extra_note":         {"USB values are not measured file-transfer throughput.", "USB 값은 실제 파일 전송 속도를 측정한 결과가 아닙니다.", "USB値は実際のファイル転送速度の測定結果ではありません。", "Los valores USB no miden la velocidad real de transferencia."},
	"trial":              {"Session query (DC-S1 / 2.4 trial)", "세션 적용 조회 (DC-S1 / 2.4 시험)", "セッション取得（DC-S1 / 2.4試験）", "Lectura con sesión (DC-S1 / 2.4)"},
	"session_note":       {"Session mode uses Windows read/write access; saved camera data is unchanged.", "세션 연결에 Windows 읽기·쓰기 권한을 사용하며, 저장된 카메라 데이터는 변경하지 않습니다.", "接続にWindowsの読み書き権限を使います。カメラの保存データは変更しません。", "La sesión usa acceso de lectura/escritura de Windows; no cambia datos guardados."},
	"trial_note":         {"Session compatibility requires a camera test. No automatic retries.", "세션 호환성은 실기 확인이 필요합니다. 자동 재시도는 하지 않습니다.", "セッションの互換性は実機確認が必要です。自動再試行はしません。", "La compatibilidad requiere pruebas en cámara. Sin reintentos automáticos."},
	"save_diag":          {"Save diagnostic ZIP", "진단 ZIP 저장", "診断ZIPを保存", "Guardar ZIP de diagnóstico"},
	"issues":             {"Open GitHub Issues", "이슈 게시판 열기", "GitHub Issuesを開く", "Abrir GitHub Issues"},
	"privacy":            {"No full serial, device path or camera payload is included. Attach this file yourself to an issue or Instagram DM.", "전체 시리얼·장치 경로·카메라 원본 응답은 포함하지 않습니다. 이슈 게시판이나 인스타그램 DM에 직접 첨부해 주세요.", "完全な製造番号、デバイスパス、カメラの生応答は含みません。IssuesまたはInstagram DMにご自身で添付してください。", "No incluye el número de serie completo, rutas ni respuestas originales. Adjunte el archivo a una incidencia o mensaje de Instagram."},
	"none":               {"No records yet.", "아직 기록이 없습니다.", "記録はまだありません。", "Todavía no hay registros."},
	"latest":             {"Latest query", "최근 조회", "最新の取得", "Última consulta"},
	"last_error":         {"Last failure", "최근 오류", "直近のエラー", "Último error"},
	"complete":           {"Read complete · session released", "조회 완료 · 연결 해제됨", "取得完了・接続解放済み", "Lectura completada · conexión liberada"},
	"failed":             {"Read failed. Save diagnostics below.", "조회에 실패했습니다. 아래에서 진단 파일을 저장할 수 있습니다.", "取得に失敗しました。下のボタンから診断を保存できます。", "La lectura falló. Puede guardar el diagnóstico abajo."},
	"cleanup":            {"Counters read; session cleanup needs checking.", "카운터 조회 성공 · 세션 종료 상태 확인 필요", "取得成功・セッション終了の確認が必要", "Contadores leídos; revise el cierre de sesión."},
	"saved":              {"Saved successfully.", "저장했습니다.", "保存しました。", "Guardado correctamente."},
	"save_failed":        {"Save failed. Check folder access; existing files were preserved.", "저장 실패 · 폴더 권한을 확인하세요. 기존 파일은 보존했습니다.", "保存失敗。フォルダー権限をご確認ください。既存ファイルは保持しました。", "Error al guardar. Revise los permisos; se conservaron los archivos existentes."},
	"serial_missing":     {"History not saved: no reliable serial number.", "시리얼 확인 불가 · 이력을 저장하지 않았습니다.", "製造番号不明のため履歴を保存しませんでした。", "Historial no guardado: número de serie no disponible."},
	"store_error":        {"History store unavailable. Existing file was not overwritten.", "이력 저장소를 읽을 수 없습니다. 기존 파일을 덮어쓰지 않았습니다.", "履歴を読み込めません。既存ファイルは上書きしていません。", "Historial no disponible. No se sobrescribió el archivo existente."},
	"auto_history":       {"Automatic local history", "사용 이력 자동 저장", "履歴の自動保存", "Historial local automático"},
	"auto_external":      {"Update selected file automatically", "지정 파일 자동 갱신", "指定ファイルを自動更新", "Actualizar el archivo automáticamente"},
	"export_history":     {"Save history as JSON…", "사용 이력 JSON 저장…", "履歴をJSONで保存…", "Guardar historial como JSON…"},
	"update_history":     {"Update selected file", "지정 파일 갱신", "指定ファイルを更新", "Actualizar archivo elegido"},
	"external_changed":   {"File changed externally. Choose a file again before updating.", "외부에서 파일이 변경되었습니다. 저장 파일을 다시 선택하세요.", "ファイルが外部で変更されました。保存先を選び直してください。", "El archivo cambió externamente. Vuelva a elegir el destino."},
	"local_note":         {"History stays on this PC. Only the last four serial digits are stored.", "이력은 이 PC에 보관합니다. 시리얼은 뒤 네 자리만 저장합니다.", "履歴はこのPCに保存され、製造番号は末尾4桁のみ記録されます。", "El historial queda en este PC. Solo se guardan los últimos cuatro dígitos de serie."},
	"external_note":      {"Exports include every camera and record. External file changes stop automatic updates.", "내보내기는 모든 카메라·기록을 포함합니다. 외부 수정 감지 시 자동 갱신을 중단합니다.", "すべてのカメラ・記録を書き出します。外部変更を検出すると自動更新を停止します。", "Se exportan todas las cámaras y registros. Los cambios externos detienen las actualizaciones."},
	"on":                 {"ON", "켜짐", "オン", "Activado"}, "off": {"OFF", "꺼짐", "オフ", "Desactivado"},
	"all":      {"All cameras", "모든 카메라", "すべてのカメラ", "Todas las cámaras"},
	"previous": {"Previous", "이전", "前へ", "Anterior"}, "next": {"Next", "다음", "次へ", "Siguiente"},
	"comparison":     {"Change from previous record", "이전 기록 대비 변화", "前回記録との差分", "Cambio desde el registro anterior"},
	"boundary":       {"New baseline: firmware change, counter decrease, or no earlier record.", "새 비교 기준: 펌웨어 변경·카운터 감소 또는 이전 기록 없음", "新しい基準：FW変更、カウンター減少、または過去の記録なし。", "Nueva referencia: firmware distinto, contador menor o sin registro previo."},
	"select":         {"Multiple cameras detected. Select one to read.", "여러 카메라가 감지되었습니다. 조회할 카메라를 선택하세요.", "複数のカメラを検出しました。読み取るカメラを選択してください。", "Se detectaron varias cámaras. Seleccione una para leer."},
	"choose":         {"Next camera", "다음 카메라 선택", "次のカメラ", "Siguiente cámara"},
	"identified":     {"Camera identified", "카메라 확인됨", "カメラを確認", "Cámara identificada"},
	"preparing":      {"Starting session…", "세션 시작 중…", "セッション開始中…", "Iniciando sesión…"},
	"reading":        {"Reading camera…", "카메라 조회 중…", "カメラ読み取り中…", "Leyendo cámara…"},
	"identifying":    {"Identifying camera…", "카메라 확인 중…", "カメラ確認中…", "Identificando cámara…"},
	"closing":        {"Closing owned session…", "세션 종료 중…", "セッション終了中…", "Cerrando sesión propia…"},
	"usage_complete": {"Counters received…", "카운터 수신 완료…", "カウンター受信済み…", "Contadores recibidos…"},
	"closing_app":    {"Finishing query and session cleanup before closing…", "조회와 세션 종료를 마친 후 창을 닫습니다…", "取得とセッション終了後に閉じます…", "Finalizando la consulta y la sesión antes de cerrar…"},
	"details":        {"Technical details", "기술 세부 정보", "技術情報", "Detalles técnicos"},
	"errors":         {"Camera error history", "카메라 오류 기록", "カメラエラー履歴", "Errores de la cámara"},
	"export":         {"PNG / text reports", "PNG / 텍스트 리포트", "PNG・テキストレポート", "Informes PNG / texto"},
	"validation":     {"Camera compatibility", "기종 검증 목록", "カメラ互換性", "Compatibilidad"},
	"versions":       {"Version history", "버전 기록", "バージョン履歴", "Historial de versiones"},
	"beta":           {"Development build · not a stable release", "개발 빌드 · 정식 배포 전", "開発ビルド・正式版ではありません", "Versión de desarrollo · no estable"},
}
