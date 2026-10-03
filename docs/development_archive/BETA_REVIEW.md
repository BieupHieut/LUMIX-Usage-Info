# beta.18 local QA — 2026-10-02

- Full Windows x64 source tests / go vet / GUI build.
- Exact firmware verification and validator attribution for all five pairs; other firmware/S1M2ES remain unverified.
- Mock L10/S1M2 regular transport and response parsing; Korean/English TXT/summary masking and verified labels.
- Pure scrollbar layout cases: fit, large, each single axis, cross-axis dependency, small, exact 125/200% fit.
- Hidden native Win32 window: small/large viewport, exact fit, forced legacy bars removed, move without resize, Enter/Escape navigation, keyboard reveal, synthetic DPI message. No physical camera access or native save dialog.
- Offscreen GDI real-font metrics/rendering: merged verification page, credits/Special Thanks, home, connection, version and export screens in Korean/English; 125/150/200% render checks. Not desktop screenshots.
- ZIP/root manifests and installed file hashes verified.

Remaining: actual beta.18 L10/S1M2 camera communication and native dialogs on target Windows PCs; cross-monitor DPI behavior in target environments. Community semantic confirmation is independent of developer mock/UI tests.
