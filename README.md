# LUMIX Usage Info v1.0.0-beta.6

Unofficial read-only camera usage reader for Panasonic LUMIX cameras.

## Verified configuration
- Camera: **DC-S1RM2**
- Firmware: **1.5**
- OS target: Windows x64
- Validated by: **@bieup_hieut**
- Validation date: **2026-09-16**

Counter meanings are fully hardware-verified only on **DC-S1RM2 / FW 1.5**.

A community tester has also confirmed that **DC-S5 / FW 2.9** can successfully return the usage data block using **v1.0.0-beta.3**. DC-S5 is therefore marked **READ COMPATIBLE**, but its raw counter meanings are **not yet behavior-verified**. Contributor: **아름프로**.

## What it reads
- Model / firmware / serial number
- Shutter Actuations (verified on DC-S1RM2 FW 1.5)
- Power / Wake Activations (verified on DC-S1RM2 FW 1.5)
- Panasonic SetupInfo diagnostic block
- Error History (legacy-code decoding is partial / informational)
- Raw counters 3–7 (unidentified)

## beta.6 highlights
- `VALIDATION` renamed to **VALIDATION & INFO**
- Validation & project pages are available even when no camera is connected
- **Disconnected / stale-data state**: last successful read remains visible with a clear warning
- USB device-tree changes trigger a non-blocking background re-check
- Refresh shows `REFRESHING...`
- Single-instance protection avoids two copies competing for one camera
- Save Report masks serial number except for the last 4 digits by default
- Clipboard copy retries briefly if another app temporarily owns the clipboard
- State-aware Pomeranian mascots: normal / info-success / PNG-export / error-disconnected
- Version History corrected (`beta.3` Sharing & Privacy entry)
- First external community validation: **DC-S5 / FW 2.9 = READ COMPATIBLE** (counter semantics unverified), contributed by **아름프로**, tested with **v1.0.0-beta.3**
- Existing beta.4 PNG-save stability fix retained

## Camera connection
1. Turn on the camera.
2. Connect it to the PC via USB.
3. Set camera USB mode to **PC(Tether)**.
4. Close LUMIX Tether if it is running.
5. Run `LUMIX_Usage_Info_v1.0.0-beta.6.exe`.

## Safety
The program intentionally implements only read paths used by this project:
- `0x1001` — standard PTP GetDeviceInfo
- `0x9414` + `0x15C00010` — Panasonic SetupInfo read

No EEPROM/ROM write, firmware modification, `SetProperty`, or unverified opcode brute-force path is implemented.

## Privacy
- Screen serial number is masked by default.
- PNG export offers Masked / Last 4 / Full Serial choices.
- TXT report uses **Last 4 digits** by default.
- Hide full serial numbers before public posting unless intentionally using Device Verification.

## Feedback
- GitHub Issues: preferred for reproducible bugs and validation results
- Instagram DM: **@bieup_hieut**

This is an unofficial community tool and is not affiliated with or endorsed by Panasonic.
