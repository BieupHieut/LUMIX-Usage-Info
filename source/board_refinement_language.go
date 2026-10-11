//go:build windows

package main

func init() {
	boardStrings["free"] = [4]string{"Free", "남은 용량", "空き容量", "Libre"}
	boardStrings["export_destination"] = [4]string{"PNG: choose destination · TXT: Local AppData / Lumerian / Reports", "PNG: 저장 위치 선택 · TXT: 로컬 AppData / Lumerian / Reports", "PNG：保存先を選択 · TXT：Local AppData / Lumerian / Reports", "PNG: elegir destino · TXT: Local AppData / Lumerian / Reports"}

	for key, row := range map[string][4]string{
		"raw_view":            {"View received values", "조회 원본 보기", "受信値を見る", "Ver valores recibidos"},
		"raw_compare":         {"Inspect the received card information and query results.\nSave the diagnostic ZIP if you need to report an issue.", "수신한 카드 정보와 조회 결과를 확인할 수 있습니다.\n문제가 있으면 진단 ZIP을 저장해 주세요.", "受信したカード情報と取得結果を確認できます。\n問題があれば診断ZIPを保存してください。", "Revise la información de tarjetas y los resultados.\nGuarde el ZIP de diagnóstico si necesita informar de un problema."},
		"received_unverified": {"Values received · interpretation pending", "값 수신됨 · 의미 확인 필요", "値を受信・意味は未確認", "Valores recibidos · sin interpretar"},
		"trace_title":         {"Query stages and response codes", "조회 단계·응답 코드", "取得段階・応答コード", "Etapas y códigos de respuesta"},
		"back":                {"Back", "뒤로", "戻る", "Volver"},
		"body":                {"Body", "바디", "ボディ", "Cuerpo"},
		"pick_body":           {"Choose a camera", "카메라 선택", "カメラを選択", "Elegir cámara"},
		"share":               {"Save report", "리포트 저장", "レポート保存", "Guardar informe"},
		"share_formats":       {"Save PNG / TXT or copy a summary", "PNG·TXT 저장 또는 요약 복사", "PNG・TXT保存、要約コピー", "Guardar PNG / TXT o copiar resumen"},
		"stage_identity":      {"Camera identification", "카메라 인식", "カメラ認識", "Identificación de cámara"},
		"stage_session":       {"Connection preparation", "연결 준비", "接続準備", "Preparación de conexión"},
		"stage_counters":      {"Usage counter query", "사용 횟수 조회", "使用回数の取得", "Consulta de contadores"},
		"stage_close":         {"Connection cleanup", "연결 종료 처리", "接続終了処理", "Cierre de conexión"},
		"stage_process":       {"Query processing", "조회 처리", "取得処理", "Proceso de consulta"},
		"failed_step":         {"Step needing attention", "확인할 단계", "確認が必要な段階", "Etapa que requiere revisión"},
		"retry_help":          {"Check USB and camera mode, close other camera apps, then read again. If it repeats, save the diagnostic ZIP and attach it to an issue.", "USB·카메라 모드를 확인하고 다른 카메라 앱을 종료한 뒤 다시 조회하세요. 반복되면 진단 ZIP을 저장해 이슈에 첨부하세요.", "USBとカメラモードを確認し、他のカメラアプリを閉じて再取得してください。続く場合は診断ZIPを保存して報告してください。", "Revise USB y modo de cámara, cierre otras apps y repita la lectura. Si persiste, guarde el ZIP y adjúntelo a una incidencia."},
		"partial_help":        {"Usage counts are available. Some additional information could not be read. View received values to check which fields arrived, then save a ZIP if needed.", "사용 횟수는 정상 조회되었습니다. 일부 부가 정보는 읽지 못했습니다. 조회 원본에서 수신 항목을 확인하고 필요하면 ZIP을 저장하세요.", "使用回数は取得できました。一部の追加情報は取得できません。受信値で項目を確認し、必要ならZIPを保存してください。", "Los contadores están disponibles; faltan datos adicionales. Revise los valores recibidos y guarde un ZIP si lo necesita."},
		"diagnostic_help":     {"View received values for numeric fields and response codes. Save the ZIP, then attach it yourself through GitHub Issues or Instagram DM @bieup_hieut.", "조회 원본에서 숫자 필드와 응답 코드를 확인하세요. ZIP을 저장한 뒤 GitHub 이슈 또는 인스타그램 @bieup_hieut DM으로 직접 첨부합니다.", "受信値で数値と応答コードを確認できます。ZIP保存後、GitHub IssuesまたはInstagram @bieup_hieutへご自身で添付してください。", "Revise los valores y códigos recibidos. Guarde el ZIP y adjúntelo en GitHub Issues o por DM de Instagram a @bieup_hieut."},
		"complete":            {"Read complete", "조회 완료", "取得完了", "Lectura completada"},
		"failed":              {"Could not complete the query", "조회를 완료하지 못했습니다", "取得を完了できませんでした", "No se pudo completar la consulta"},
		"partial":             {"Usage counts read · some additional information unavailable", "사용 횟수 조회 완료 · 일부 부가 정보 조회 불가", "使用回数取得済み・一部の追加情報は取得不可", "Contadores leídos · faltan datos adicionales"},
	} {
		boardStrings[key] = row
	}
}
