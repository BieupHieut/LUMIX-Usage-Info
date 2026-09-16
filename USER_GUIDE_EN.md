# LUMIX Usage Info v1.0.0-beta.6 User Guide

## Connect
1. Turn on the camera.
2. Connect it by USB.
3. Set USB Mode to **PC(Tether)**.
4. Close LUMIX Tether if it is running.
5. Launch the tool.

## Verified values on DC-S1RM2 / FW 1.5
- **Shutter Actuations** — physical shutter actuation count
- **Power / Wake Activations** — power-on / sleep-wake related activation count

## Disconnected state
If a previously-read camera becomes unavailable, beta.6 keeps the last successful values visible and clearly labels them as **Disconnected / last known data**. Reconnect USB and press Refresh to obtain a new read.

## Validation & Info
Available even with no camera connected:
- Camera List
- Version History
- Developer Credits

## Privacy
- Screen serial: masked by default
- PNG: Masked / Last 4 / Full
- TXT report: Last 4 by default

## Safety
Read-only project path only. No camera write / EEPROM / ROM / firmware modification code is implemented.

## Camera validation status
- **DC-S1RM2 / FW 1.5:** FULLY VERIFIED.
- **DC-S5 / FW 2.9:** READ COMPATIBLE from an external community test by **아름프로** using v1.0.0-beta.3. Counter meanings are not yet verified, so the app keeps Raw Counter labels.
- Other listed cameras: NEEDS VALIDATION.
