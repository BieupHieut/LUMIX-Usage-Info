//go:build windows

package main

import (
	"fmt"
	"strings"
)

func init() {
	for k, v := range map[string][2]string{
		"EXPORT / SHARE": {"書き出し・共有", "Exportar / compartir"},
		"Create a clean PNG or copy a text summary": {"PNG保存・テキスト概要のコピー", "Guarde PNG o copie el resumen"},
		"1. CHOOSE PURPOSE":                         {"1. 用途を選択", "1. Seleccione el uso"},
		"DEVICE VERIFICATION":                       {"機種検証用", "Verificación de cámara"},
		"Public Share masks the serial. Device Verification starts with Last 4 digits; Full Serial is optional.": {"公開用は製造番号を非表示にします。検証用は末尾4桁から選べます。", "Compartir oculta la serie. Verificación muestra los últimos 4 dígitos por defecto."},
		"2. SERIAL NUMBER IN EXPORT": {"2. 書き出す製造番号の範囲", "2. Número de serie al exportar"},
		"Full serial number will be included. Avoid posting the exported image publicly.": {"製造番号の全桁が含まれます。画像の公開は避けてください。", "Se incluirá la serie completa. Evite publicar la imagen."},
		"Choose where to save PNG. Default: Local AppData · Lumerian · Reports.":          {"PNGの保存先を選択。初期設定：Local AppData · Lumerian · Reports。", "Elija el destino PNG. Predeterminado: Local AppData · Lumerian · Reports."},
		"SAVING…": {"保存中…", "Guardando…"}, "Saving report...": {"レポート保存中…", "Guardando informe…"},
		"CAMERA INFORMATION": {"カメラ情報", "Información de la cámara"}, "Model": {"モデル", "Modelo"},
		"Shutter Actuations":                               {"シャッター作動回数", "Accionamientos del obturador"},
		"Power / Wake Activations":                         {"電源・復帰回数", "Encendidos / activaciones"},
		"Shutter Actuations (estimated)":                   {"シャッター作動回数（推定）", "Obturador (estimado)"},
		"Power / Wake Activations (estimated)":             {"電源・復帰回数（推定）", "Encendidos (estimado)"},
		"Flash Activations (estimated)":                    {"フラッシュ作動回数（推定）", "Activaciones del flash (estimado)"},
		"Power-saving Events (estimated)":                  {"省電力動作回数（推定）", "Eventos de ahorro de energía (estimado)"},
		"Unidentified":                                     {"意味不明", "Sin identificar"},
		"Read-only service data · No camera data modified": {"読み取り専用・データ変更なし", "Solo lectura · Sin modificar datos"},
		"3/4: legacy candidates STBCNT / PSVCNT · Meanings unverified. 5–7: unidentified.":                                     {"3/4：旧名STBCNT / PSVCNTに基づく推定。5～7：意味不明。", "3/4: STBCNT / PSVCNT, sin verificar. 5–7: sin identificar."},
		"Up to 16 service-history records · legacy Panasonic reference codes":                                                  {"最大16件・旧Panasonic参照コード", "Hasta 16 registros · Referencias antiguas"},
		"A slot may contain a date marker while the error code is 00000000; this app does not treat that as a recorded error.": {"日時があってもコード00000000はエラーとして表示しません。", "Una fecha con código 00000000 no se considera un error registrado."},
		"PREVIOUS": {"前へ", "Anterior"}, "NEXT": {"次へ", "Siguiente"},
		"%d verified · %d estimated": {"検証済み%d・推定%d", "%d verificados · %d estimados"},
		"ESTIMATED":                  {"対応推定", "Estimado"}, "✓ VERIFIED": {"✓ 検証済み", "✓ Verificado"},
		"PC(Tether) · Counter meanings unverified": {"PC(Tether)・意味未検証", "PC(Tether) · Valores sin verificar"},
		"Not yet verified":                         {"報告待ち", "Sin verificar"}, "FW ": {"FW ", "FW "},
		"Shutter ✓   Power / Wake ✓": {"シャッター ✓   電源・復帰 ✓", "Obturador ✓   Encendido ✓"},
		"Validated by ":              {"確認：", "Verificado por "},
		"Verified for the listed firmware only. Other entries: estimated support list.": {"表示されたFWのみ検証済み。その他は対応推定です。", "Verificado solo con el firmware indicado. El resto es compatibilidad estimada."},
		"Invalid external camera data. Using bundled verification data.":                {"外部機種データが無効のため内蔵データを使用しています。", "Datos externos no válidos. Se utilizan los datos incluidos."},
		"Counters 3/4 estimated · 5–7 and error mappings unverified.":                   {"3/4は推定・5～7とエラー解釈は未検証", "3/4 estimados · 5–7 e interpretación de errores sin verificar"},
		"DEVELOPER CREDITS": {"制作者・協力者", "Créditos"}, "CREATOR & CONTRIBUTORS": {"制作者と協力者", "Creador y colaboradores"},
		"CREATOR": {"制作者", "Creador"}, "Creator · Product design": {"制作・製品デザイン", "Creación · Diseño del producto"},
		"CONTRIBUTORS": {"協力者", "Colaboradores"}, "Camera tests and feedback": {"カメラ確認・フィードバック", "Pruebas de cámara y comentarios"},
		"SPECIAL THANKS":                                                      {"感謝", "Agradecimientos"},
		"Panasonic LUMIX Café Forum":                                          {"Panasonic LUMIXカフェフォーラム", "Foro Café Panasonic LUMIX"},
		"Thank you to community members for tests and feedback.":              {"確認とフィードバックに感謝します。", "Gracias por las pruebas y comentarios."},
		"Unofficial read-only community tool · Not affiliated with Panasonic": {"非公式の読み取り専用ツール・Panasonicとは無関係です", "Herramienta comunitaria no oficial · Sin afiliación con Panasonic"},
		"Meanings have not been verified for this model / firmware. Raw fields: shutter = Counter 2, power/wake = Counter 1.": {"このモデル・FWでは意味未検証です。シャッター＝カウンター2、電源＝1。", "Significados sin verificar en este modelo/FW. Obturador = contador 2; encendido = 1."},
		"Read directly from the connected camera · Read-only":                                                                 {"接続カメラから直接取得・読み取り専用", "Leído directamente de la cámara · Solo lectura"},
		"Generated": {"作成日時", "Generado"}, "Validation": {"検証状態", "Verificación"},
		"LUMIX Usage Info": {"LUMIX使用情報", "Información de uso LUMIX"}, "Camera data revision": {"機種データ更新日", "Revisión de datos"},
		"Made by @bieup_hieut": {"制作 @bieup_hieut", "Creado por @bieup_hieut"},
		"NOT VERIFIED":         {"未検証", "Sin verificar"}, "Not verified": {"未検証", "Sin verificar"}, "Not available": {"不明", "No disponible"},
		"PNG Image (*.png)": {"PNG画像 (*.png)", "Imagen PNG (*.png)"}, "All Files (*.*)": {"すべてのファイル (*.*)", "Todos los archivos (*.*)"},
		"Save LUMIX Usage Info as PNG":                          {"LUMIX使用情報をPNGで保存", "Guardar información LUMIX como PNG"},
		"Read-only — camera settings and data are not changed.": {"読み取り専用。カメラの設定やデータは変更しません。", "Solo lectura: no se modifican los ajustes ni los datos de la cámara."},
	} {
		boardLegacyStrings[k] = v
	}
}

func boardLocalizedReport(r UsageResult, serialMode string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "LUMERIAN LUMIX Usage Board v%s\r\n%s\r\n\r\n", appVersion, r.Timestamp.Format("20060102_150405"))
	fmt.Fprintf(&b, "%s: %s\r\n%s: %s\r\n%s: %s\r\n", localizeText("Model"), r.Model, localizeText("Firmware"), r.Firmware, localizeText("Serial Number"), serialForMode(r.Serial, serialMode))
	status := "estimated"
	if r.SemanticsVerified {
		status = "verified"
	}
	fmt.Fprintf(&b, "%s: %s\r\n", localizeText("Validation"), boardText(status))
	by, date := validationDetails(r)
	fmt.Fprintf(&b, "%s%s · %s\r\n\r\n", localizeText("Validated by "), localizeText(by), localizeText(date))
	counts := []uint64{uint64(r.PowerWake), uint64(r.Shutter), uint64(r.Raw3), uint64(r.Raw4), uint64(r.Raw5), uint64(r.Raw6), uint64(r.Raw7)}
	for i, v := range counts {
		fmt.Fprintf(&b, "%d · %s: %s\r\n", i+1, counterName(i+1, r), formatNumber(v))
	}
	fmt.Fprintf(&b, "\r\n%s\r\n", boardText("errors"))
	if len(r.Errors) == 0 {
		fmt.Fprintln(&b, localizeText("No recorded non-zero error codes found."))
	}
	for _, e := range r.Errors {
		fmt.Fprintf(&b, "%02d · %s · %s\r\n", e.Index, e.DateText, e.CodeText)
	}
	fmt.Fprintf(&b, "\r\n%s\r\n%s\r\n\r\nGitHub Issues: https://github.com/BieupHieut/LUMIX-Usage-Info/issues\r\nInstagram DM: @bieup_hieut\r\n", boardText("safe"), localizeText("Counters 3/4 estimated · 5–7 and error mappings unverified."))
	return b.String()
}
