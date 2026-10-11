//go:build windows

package main

// Stable reads deliberately exclude the unverified capacity experiment.
func boardPublicRequest(session, suite bool) boardRequest {
	return boardRequest{Session: session, Extras: false, Suite: suite}
}
