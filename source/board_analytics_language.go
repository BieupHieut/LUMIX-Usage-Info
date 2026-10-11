//go:build windows

package main

func init() {
	for key, row := range map[string][4]string{
		"analysis": {"Analysis", "분석", "分析", "Análisis"},
		"day":      {"Daily", "일별", "日別", "Diario"}, "month": {"Monthly", "월별", "月別", "Mensual"}, "year": {"Yearly", "연별", "年別", "Anual"},
		"settings_short": {"Save settings", "저장 설정", "保存設定", "Guardado"},
		"details_short":  {"Technical details", "세부 정보", "詳細情報", "Detalles"}, "errors_short": {"Camera errors", "오류 기록", "エラー履歴", "Errores"},
		"raw_short": {"Received values", "조회 원본", "受信値", "Valores recibidos"}, "card_short": {"Card capacity · test", "카드 용량 · 시험", "カード容量・試験", "Capacidad · prueba"},
		"zip_short": {"Diagnostic ZIP", "진단 ZIP 저장", "診断ZIP保存", "ZIP de diagnóstico"}, "contact_short": {"Report an issue", "문제 문의", "問題を報告", "Informar de un problema"},
		"delete_selected": {"Delete selected", "선택 삭제", "選択を削除", "Eliminar selección"},
		"clear_selection": {"Clear selection", "선택 해제", "選択を解除", "Deseleccionar"},
		"clean_short":     {"Clean duplicates", "중복 기록 정리", "重複を整理", "Limpiar duplicados"},
		"folder_short":    {"History folder", "이력 폴더", "履歴フォルダー", "Carpeta del historial"},
		"deleted_gap":     {"Comparison gap after deletion", "삭제 구간 · 비교 제외", "削除区間・比較なし", "Intervalo eliminado"},
		"delete_confirm": {
			"Delete %d selected records?\nA full local backup is created first. Manage this history with Beta.13 or later after deleting. Deleted intervals will not be joined to calculate increases.\nExternal copies stay unchanged and their automatic updates will be turned off.",
			"체크한 기록 %d개를 삭제할까요?\n전체 로컬 백업을 먼저 만듭니다. 삭제 후 이력은 Beta.13 이상에서 관리해 주세요. 삭제 구간을 건너뛰어 증가량을 계산하지 않습니다.\n외부 사본은 유지하며 외부 파일 자동 갱신은 꺼집니다.",
			"選択した%d件を削除しますか？\n先に全体をバックアップします。削除後の履歴はBeta.13以降で管理してください。削除区間をまたぐ増加量は計算しません。\n外部コピーは保持し、自動更新をオフにします。",
			"¿Eliminar los %d registros seleccionados?\nSe crea una copia local completa. Tras eliminar, use Beta.13 o posterior para este historial. No se calcularán aumentos a través de intervalos eliminados.\nSe conservan las copias externas y se desactiva su actualización automática."},
		"delete_done":    {"Deleted; backup in history-backups. External auto-update is off.", "선택 기록 삭제 완료 · history-backups에 백업 · 외부 자동 갱신 꺼짐", "削除完了・history-backupsに保存・外部自動更新はオフ", "Eliminado; copia en history-backups. Actualización externa desactivada."},
		"observed_total": {"Observed increase in view", "표시 구간 관측 증가량", "表示期間の観測増加量", "Aumento observado en pantalla"},
		"analysis_empty": {"Choose a camera with saved records. At least two comparable readings in one period are needed for a bar.", "저장된 이력이 있는 카메라를 선택하세요. 같은 기간에 비교 가능한 조회가 두 번 이상 있어야 막대를 표시합니다.", "履歴のあるカメラを選択してください。棒グラフには同じ期間内の比較可能な記録が2件以上必要です。", "Elija una cámara con historial. Cada barra necesita al menos dos lecturas comparables dentro del mismo periodo."},
		"unallocated": {
			"Across period boundaries, all history (not in bars): shutter +%s · power +%s",
			"기간 경계를 넘은 증가 · 전체 이력 (막대에서 제외): 셔터 +%s · 전원 +%s",
			"期間をまたぐ増加・全履歴（棒の対象外）：シャッター +%s・電源 +%s",
			"Entre periodos, todo el historial (fuera de barras): obturador +%s · encendidos +%s"},
		"analysis_note": {
			"Same-period differences only; bars are not total photos taken. — No comparable data; 0 = observed no increase. Local PC dates. Non-comparable intervals: %d.",
			"같은 기간 안의 조회 차이만 합산합니다. 실제 촬영량 전체는 아닙니다.\n— 비교 자료 없음 · 0 증가 없음 확인 · PC 날짜 기준 · 비교 제외 구간 %d개",
			"同期間内の差分のみで、全撮影枚数ではありません。— 比較資料なし、0 増加なしを観測。PCの現地日付。比較不可区間：%d。",
			"Solo diferencias dentro del mismo periodo; no son todas las fotos tomadas. — Sin datos comparables; 0 = sin aumento observado. Fechas del PC. Intervalos excluidos: %d."},
	} {
		boardStrings[key] = row
	}
}
