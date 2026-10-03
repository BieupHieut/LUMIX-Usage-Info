//go:build windows

package main

type cameraStatusStyle struct{ Background, Accent, Detail uintptr }

// Color accompanies explicit status words; color alone never signals validation.
func cameraRowStyle(verified bool) cameraStatusStyle {
	if verified {
		return cameraStatusStyle{rgb(29, 43, 61), rgb(167, 200, 240), rgb(174, 197, 225)}
	}
	return cameraStatusStyle{rgb(38, 36, 33), rgb(230, 191, 130), rgb(199, 189, 173)}
}
