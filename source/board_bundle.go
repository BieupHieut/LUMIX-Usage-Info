//go:build windows

package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
)

// Use the same whitelist for every attempt. Neither private worker replies nor
// identity strings are marshalled directly, including historical attempts.
func boardDiagnosticBundle(d boardDiagnostic, previous []boardDiagnostic) ([]byte, error) {
	base, e := boardDiagnosticZIP(d)
	if e != nil || d.Policy != "s1-comprehensive-v1" {
		return base, e
	}
	var out bytes.Buffer
	z := zip.NewWriter(&out)
	copyZIP := func(payload []byte, prefix string) error {
		r, e := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
		if e != nil {
			return e
		}
		for _, f := range r.File {
			src, e := f.Open()
			if e != nil {
				return e
			}
			dst, e := z.Create(prefix + f.Name)
			if e != nil {
				src.Close()
				return e
			}
			_, e = io.Copy(dst, src)
			src.Close()
			if e != nil {
				return e
			}
		}
		return nil
	}
	if e = copyZIP(base, ""); e != nil {
		return nil, e
	}
	prior := []boardDiagnostic{}
	for _, p := range previous {
		if p.Policy == "s1-comprehensive-v1" && (p.RunID == "" || p.RunID != d.RunID) {
			prior = append(prior, p)
		}
	}
	if len(prior) > 9 {
		prior = prior[len(prior)-9:]
	}
	count := 0
	for _, p := range prior {
		payload, e := boardDiagnosticZIP(p)
		if e != nil {
			return nil, e
		}
		count++
		if e = copyZIP(payload, fmt.Sprintf("previous_attempts/%02d/", count)); e != nil {
			return nil, e
		}
	}
	w, e := z.Create("ATTEMPTS.txt")
	if e != nil {
		return nil, e
	}
	_, e = fmt.Fprintf(w, "Current result: diagnostic.json / SUMMARY.txt\r\nPrevious attempts included: %d\r\nUp to ten recent attempts from this app session are retained. No automatic upload.\r\n현재 결과는 최상위 파일, 이전 시도 %d회는 previous_attempts 폴더에 있습니다.\r\n현재 앱 실행 중 최근 10회까지 보관합니다. 앱을 닫기 전에 ZIP을 저장하세요.\r\n", count, count)
	if e != nil {
		return nil, e
	}
	if e = z.Close(); e != nil {
		return nil, e
	}
	return out.Bytes(), nil
}
