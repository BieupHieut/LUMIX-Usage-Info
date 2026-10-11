//go:build windows

package main

func init() {
	for key, row := range map[string][4]string{
		"external_no_file": {"No external history file selected.", "외부 이력 파일을 아직 지정하지 않았습니다.", "外部の履歴ファイルは未指定です。", "No se ha elegido un archivo externo de historial."},
		"diag_retry_short": {"Check USB and camera mode, then read again. Save a ZIP if it repeats.", "USB와 카메라 모드를 확인한 뒤 다시 조회하세요. 반복되면 ZIP을 저장해 주세요.", "USBとカメラモードを確認して再取得してください。続く場合はZIPを保存してください。", "Revise USB y el modo de cámara y repita. Si persiste, guarde el ZIP."},
		"diag_send_short":  {"To report a problem, save the ZIP and attach it to an issue or DM @bieup_hieut.", "문제가 있으면 ZIP을 저장해 이슈 게시판 또는 @bieup_hieut DM으로 첨부하세요.", "問題があればZIPを保存し、Issuesまたは@bieup_hieutへのDMに添付してください。", "Para informar de un problema, adjunte el ZIP en Issues o por DM a @bieup_hieut."},
		"clean_history":    {"Clean unchanged records", "변화 없는 기록 정리", "変化のない記録を整理", "Limpiar registros sin cambios"},
		"data_folder":      {"Open history folder", "이력 폴더 열기", "履歴フォルダーを開く", "Abrir carpeta del historial"},
		"clean_confirm": {
			"%s\nRemove %d intermediate records with unchanged shutter AND power counts?\nThe first/latest records for each camera and change boundaries are kept. A local backup is created first.\nExternal files stay unchanged; automatic external updates will be turned off.",
			"%s\n셔터·전원 횟수가 모두 같은 중간 기록 %d개를 정리할까요?\n카메라별 최초·최신 기록과 변경 구간은 보존하고, 먼저 로컬 백업을 만듭니다.\n외부 파일은 유지하며 외부 파일 자동 갱신은 꺼집니다.",
			"%s\nシャッターと電源回数が同じ中間記録%d件を削除しますか？\n各カメラの最初・最新の記録と変更境界は保持し、先にローカルバックアップを作成します。\n外部ファイルは保持し、自動更新をオフにします。",
			"%s\n¿Eliminar %d registros intermedios sin cambios en ambos contadores?\nSe conservan el primero, el último y los límites de cambios por cámara. Se crea una copia local primero.\nLos archivos externos se conservan y su actualización automática se desactiva."},
		"clean_none":         {"No unchanged intermediate records to remove.", "정리할 중복 중간 기록이 없습니다.", "削除する重複した中間記録はありません。", "No hay registros intermedios duplicados."},
		"clean_changed":      {"History changed. Review the records and try again.", "이력이 변경되었습니다. 목록을 확인하고 다시 시도하세요.", "履歴が変更されました。確認して再試行してください。", "El historial cambió. Revíselo e inténtelo de nuevo."},
		"clean_done":         {"Cleaned; backup in history-backups. External auto-update is off.", "정리 완료 · history-backups에 백업 · 외부 파일 자동 갱신 꺼짐", "整理完了・history-backupsに保存・外部の自動更新はオフ", "Limpieza completada; copia en history-backups. Actualización externa desactivada."},
		"connection_policy":  {"Each read releases the connection; session cleanup is attempted if this tool opened it.", "조회 후 연결을 해제하며, 이 도구가 시작한 세션은 종료를 시도합니다.", "取得後に接続を解放し、このツールで開始したセッションの終了を試みます。", "Cada lectura libera la conexión; se intenta cerrar la sesión iniciada por esta herramienta."},
		"shutter_help_title": {"Why this can differ from photo count", "사진 수와 셔터 횟수가 다른 이유", "写真枚数と異なる理由", "Por qué puede diferir del número de fotos"},
		"shutter_help_body": {
			"[Shutter Behavior at Power Off] → [CLOSE] closes the shutter at shutdown. This action may also count.\nOur DC-S1RM2 / FW 1.5 test recorded +1. Other model/firmware combinations need confirmation.\nClick ! again to close.",
			"[전원 차단 시 셔터 동작]을 [CLOSE]로 설정하면 종료 시 셔터가 닫히며, 이 동작도 횟수에 포함될 수 있습니다.\nDC-S1RM2 / 펌웨어 1.5 시험에서 +1을 확인했습니다. 다른 기종·펌웨어는 별도 확인이 필요합니다.\n! 버튼을 다시 누르면 닫힙니다.",
			"[Shutter Behavior at Power Off] → [CLOSE]では電源オフ時にシャッターが閉じ、この動作も計数される場合があります。\nDC-S1RM2 / FW 1.5の試験で+1を確認。他の機種・FWは確認が必要です。\n!を再度押すと閉じます。",
			"[Shutter Behavior at Power Off] → [CLOSE] cierra el obturador al apagar. Esta acción puede contar.\nNuestra prueba con DC-S1RM2 / FW 1.5 registró +1. Otras combinaciones requieren confirmación.\nPulse ! de nuevo para cerrar."},
	} {
		boardStrings[key] = row
	}
}
