# Validation Matrix

| Camera | Firmware | Tool | SetupInfo transport | Shutter meaning | Power/Wake meaning | Error descriptions | Validator | Date |
|---|---|---|---|---|---|---|---|---|
| DC-S1RM2 | 1.5 | 1.0.0-beta.3 | VERIFIED | VERIFIED | VERIFIED | LEGACY / PARTIAL | @bieup_hieut | 2026-09-16 |

## Status definitions

- **VERIFIED** — behavior checked on real hardware with controlled tests.
- **MODEL VERIFIED / FW UNVERIFIED** — model known but exact firmware semantics not hardware-tested.
- **LEGACY / PARTIAL** — decoding informed by older Panasonic service references and not claimed as an official mapping for the current body.
- **UNIDENTIFIED** — raw value exists but meaning is not established.

New rows should be added only after controlled hardware validation. Avoid marking a camera as verified merely because a raw value was readable.
