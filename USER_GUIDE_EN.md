# LUMIX Usage Info v1.0.0-beta.3 User Guide

LUMIX Usage Info is an unofficial, read-only Windows utility for reading selected usage and diagnostic information from Panasonic LUMIX cameras over USB PC(Tether).

The currently hardware-validated combination is **DC-S1RM2 / Firmware 1.5**.

## Connect

1. Turn on the camera.
2. Connect it to the PC via USB.
3. Set USB Mode to **PC(Tether)**.
4. Close LUMIX Tether if it is running.
5. Run the tool.

## Verified counters

**Shutter Actuations** represents verified physical shutter activity on the validated setup. Mechanical exposures increment it, while an electronic-shutter test did not.

**Power / Wake Activations** increments on OFF→ON and on a sleep→wake cycle. USB reconnect alone did not increment it.

The main screen can show a since-last-refresh delta when the same camera is read again.

## Serial privacy

Serial Number is privacy-masked by default on the main screen. Use `SHOW / HIDE` to toggle the full value.

## PNG export

Choose `SAVE PNG`, then select:

- **Public Share** — privacy-first preset.
- **Device Verification** — useful for ownership or used-camera verification.

Serial display options:

- Masked
- Last 4 digits
- Full Serial

Do not publicly post a Full Serial export unless that disclosure is intentional.

## Copy Summary

Copies model, firmware, serial according to the selected privacy mode, counters, and validation status to the clipboard.

## Error History

The tool can parse up to 16 service-history slots. Legacy Panasonic service references are used for partial descriptions; they are not claimed to be an official S1RM2 error-code map.

## Feedback

GitHub Issues are preferred. Instagram DM: **@bieup_hieut**.
