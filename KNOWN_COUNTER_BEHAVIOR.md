# Known Counter Behavior — DC-S1RM2 / Firmware 1.5

These are controlled behavior observations from a real camera, not official Panasonic counter-name definitions.

## Counter #2 — Shutter Actuations (verified behavior)
- Mechanical shutter exposure: +1
- Two mechanical actuations: +2
- Electronic shutter exposure: +0
- High Resolution test using the tested settings: +0
- Power OFF with [Power-off Shutter] = CLOSE: +1
- Sensor Cleaning + required restart: +1 total observed
- Pixel Refresh + required restart: +3 total observed with Power-off Shutter=CLOSE; exact breakdown not isolated

## Counter #1 — Power / Wake Activations (verified behavior)
- OFF → ON: +1
- Sleep → Wake cycle: +1
- USB disconnect/reconnect only: +0

## Counters #3–#7
Unidentified. Legacy candidate names may be shown only as unverified references.


# Community Read Compatibility — DC-S5 / Firmware 2.9

Contributor: **아름프로** · Tool: **v1.0.0-beta.3**

The camera successfully returned the usage data block, including Raw Counter 2 and Raw Counter 1. Their meanings have **not** been behavior-verified on DC-S5. Keep raw labels until controlled testing is complete.
