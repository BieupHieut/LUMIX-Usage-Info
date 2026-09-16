# LUMIX Usage Info

**Current release:** `v1.0.0-beta.3`  
**Verified hardware:** Panasonic LUMIX **DC-S1RM2**  
**Verified firmware:** **Ver. 1.5**  
**Platform:** Windows x64  
**Validated by:** `@bieup_hieut`

LUMIX Usage Info is an **unofficial, read-only community tool** that reads selected usage and diagnostic information from supported Panasonic LUMIX cameras over USB in **PC(Tether)** mode.

> This project is not affiliated with or endorsed by Panasonic. Panasonic and LUMIX are trademarks of their respective owner.

## What beta.3 adds

- Serial Number privacy toggle on the main screen.
- `Save PNG` export flow with **Public Share** and **Device Verification** presets.
- Serial Number export choices: **Masked / Last 4 digits / Full Serial**.
- `Copy Summary` for quick sharing in GitHub Issues, forums, or DMs.
- Since-last-refresh deltas for verified counters.
- Improved Error History empty-state and validation labeling.
- Validation Info page showing verified camera / firmware / tool version / validator.
- Worker-process camera communication with timeout so a stalled WPD/PTP request does not freeze the UI.

## Verified information on DC-S1RM2 / FW 1.5

### Shutter Actuations — VERIFIED

Behavior observed on real hardware:

- Mechanical shutter exposure: `+1` per physical actuation.
- Electronic shutter exposure: `+0`.
- Power OFF while **[Power-off Shutter] = CLOSE**: can add `+1` because the physical shutter closes.
- High Resolution test: `+0` in the tested setup (electronic shutter path).
- Pixel Refresh + required restart: a total increase was observed in testing, but the exact internal breakdown is **not isolated**. Do not assume Pixel Refresh always adds a fixed number.

### Power / Wake Activations — VERIFIED

- Camera OFF → ON: `+1`.
- Sleep → wake cycle: `+1`.
- USB disconnect / reconnect only: `+0`.

The exact moment inside the sleep/wake cycle at which the counter increments has not been isolated, so the broad label **Power / Wake Activations** is intentional.

### Counters 3–7 — UNIDENTIFIED

The tool exposes the raw values only in Technical Details. Legacy Panasonic names may be shown only as historical candidates and must not be treated as confirmed DC-S1RM2 meanings.

## PNG export privacy

`Save PNG` opens an export screen:

- **Public Share** — privacy-first preset.
- **Device Verification** — suitable for ownership or used-camera verification.
- Serial display can still be selected manually as **Masked / Last 4 digits / Full Serial**.

If Full Serial is selected, do not post the exported image publicly unless you intentionally want to expose the serial number.

## Connection

1. Turn on the camera.
2. Connect it to the PC by USB.
3. On the camera, set USB Mode to **PC(Tether)**.
4. Close LUMIX Tether if it is running.
5. Run `LUMIX_Usage_Info_v1.0.0-beta.3.exe`.

## Read-only safety

The public tool intentionally uses only read paths:

- Standard PTP `0x1001` — GetDeviceInfo
- Panasonic read `0x9414` with `0x15C00010`

Camera-write commands such as `0x9403`, `0x940B`, `0x9704`, EEPROM/ROM writes, and firmware modification are **not implemented**.

## Feedback / additional validation

GitHub Issues are preferred because test results remain searchable for other users. Instagram DM is also welcome: **@bieup_hieut**.

When reporting results, include:

- Camera model
- Firmware version
- Tool version
- Windows version
- What you did before and after Refresh
- Counter change observed

Please redact the Serial Number from public screenshots unless it is intentionally being used for device verification.

## Build

The program is written in Go and uses Win32/WPD directly.

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath -ldflags="-H=windowsgui -s -w" \
  -o LUMIX_Usage_Info_v1.0.0-beta.3.exe main.go
```

Recommended before release:

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go vet main.go
```

