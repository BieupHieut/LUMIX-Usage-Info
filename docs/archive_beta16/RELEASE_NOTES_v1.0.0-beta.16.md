# v1.0.0-beta.16 — L10 manual USB diagnostics

## Added

- Dedicated experimental DC-L10 page under Validation & Info, with firmware entry and LUMIX Lab connection guidance in Korean/English.
- OS-only USB candidate discovery and optional disconnected baseline comparison.
- Standard PTP read and gated existing Panasonic read, each with explicit consent, isolated timeout/cancellation and no automatic retry.
- Redacted report preview and manual JSON + TXT export for positive and negative results.
- Clearly marked synthetic sample mode for offline UI/export inspection.
- Official Data Act research notes covering L10 and other camera groups, flash/power-saving leads, data-size differences, service-access uncertainty and an unsent inquiry draft.

## Verification

Source tests, Go vet and Windows x64 GUI build passed. Offline command-gate, parser, privacy, report rollback and Korean/English GDI checks passed. Existing window/DPI/keyboard checks remain in the test suite.

**No physical L10 connection or shutter-count meaning has been verified.** Successful raw read remains UNVERIFIED. S9 and existing verified combinations keep their previous route and validation scope.
